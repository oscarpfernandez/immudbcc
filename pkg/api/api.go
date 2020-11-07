package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/oscarpfernandez/immudbcc/pkg/crypt"
	"github.com/oscarpfernandez/immudbcc/pkg/doc"
	"github.com/oscarpfernandez/immudbcc/pkg/worker"

	immuclient "github.com/codenotary/immudb/pkg/client"
)

const (
	defaultNumWorkers = 500
)

// Config represents the required API options.
type Config struct {
	EncryptionToken string
	NumberWorkers   int
	IsSafeSet       bool
	ClientOptions   *immuclient.Options
}

// DefaultConfig defines a configuration with stock options.
func DefaultConfig() *Config {
	return &Config{
		NumberWorkers: defaultNumWorkers,
		ClientOptions: immuclient.DefaultOptions().WithAuth(false),
	}
}

// WithNumberWorkers set the number of workers used on Store actions.
func (c *Config) WithNumberWorkers(numWorkers int) *Config {
	c.NumberWorkers = numWorkers
	return c
}

// WithClientOptions set the client options used to initialise the ImmuDB client.
func (c *Config) WithClientOptions(options *immuclient.Options) *Config {
	c.ClientOptions = options
	return c
}

func (c *Config) WithEncryptionToken(token string) *Config {
	c.EncryptionToken = token
	return c
}

func (c *Config) WithSafeSet(isSafeSet bool) *Config {
	c.IsSafeSet = isSafeSet
	return c
}

type StoreDocumentResult struct {
	Index   uint64
	Hash    []byte
	HashEnc []byte
}

// Manager represents the object required to use the API.
type Manager struct {
	conf   Config
	client immuclient.ImmuClient
}

// New creates a new API manager object.
func New(c *Config) (*Manager, error) {
	client, err := immuclient.NewImmuClient(c.ClientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create ImmuDB client: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	return &Manager{
		conf:   *c,
		client: client,
	}, nil
}

// StoreDocument saves a JSON document in the database, marshalling its structure
// into key-value properties, representing the transversal property paths of the
// original object.
func (m *Manager) StoreDocument(ctx context.Context, docID string, r io.Reader) (*StoreDocumentResult, error) {
	entryList, err := doc.GeneratePropertyList(docID, r)
	if err != nil {
		return nil, err
	}

	//doc.PrintPropertyEntryList(entryList)

	workers := worker.NewWriteWorkerPool(m.conf.NumberWorkers, m.conf.IsSafeSet, m.client)
	if err := workers.StartWorkers(ctx); err != nil {
		return nil, err
	}
	defer workers.Stop()

	// Process the entry list submitting them to the available workers.
	resultChan, shutdown, errChan := workers.Write(entryList)

	var resultHash doc.PropertyHashList
	var errList []string
	func() {
		counter := 0
		for {
			select {
			case hash := <-resultChan:
				if hash != nil {
					resultHash = append(resultHash, hash)
					counter++
					if counter == len(entryList) {
						// All entries are processed. We can finish.
						workers.Stop()
					}
				}
			case err := <-errChan:
				if err != nil {
					counter++
					errList = append(errList, err.Error())
					if counter == len(entryList) {
						// All entries are processed. We can finish.
						workers.Stop()
					}
				}
			case <-shutdown:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	if len(errList) > 0 {
		return nil, fmt.Errorf("failed to store document ID '%s': %v", docID, strings.Join(errList, "; "))
	}

	sort.Sort(resultHash)

	indexes := resultHash.Indexes()
	objectHash := resultHash.Hash()
	encObjectHash, err := crypt.Encrypt(objectHash, m.conf.EncryptionToken)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt global hash :%v", err)
	}

	objectManif := &doc.ObjectManifest{
		ObjectID:        docID,
		PropertyIndexes: indexes,
		ObjectHash:      objectHash,
		ObjectEncHash:   encObjectHash,
	}

	vi, err := m.writeDocumentManifest(ctx, objectManif)
	if err != nil {
		return nil, fmt.Errorf("unable to store manifes of object '%s': %v", docID, err)
	}

	return &StoreDocumentResult{
		Index:   vi.Index,
		Hash:    objectHash,
		HashEnc: encObjectHash,
	}, nil
}

func (m *Manager) writeDocumentManifest(ctx context.Context, om *doc.ObjectManifest) (*immuclient.VerifiedIndex, error) {
	documentKey := []byte(fmt.Sprintf("manifest/%s", om.ObjectID))

	documentValue, err := json.Marshal(om)
	if err != nil {
		return nil, fmt.Errorf("unable to marshall object maifest: %v", err)
	}

	return m.client.SafeSet(ctx, documentKey, documentValue)
}

type DocumentProof struct {
	ObjectID string
	HashList doc.PropertyHashList
	Hash     doc.Hash
	EncHash  doc.EncHash
}

func (m *Manager) GetDocumentProof(ctx context.Context, docId string, docIndex uint64) (*DocumentProof, error) {
	itemList, err := m.client.Scan(ctx, []byte("manifest/"+docId))
	if err != nil {
		return nil, err
	}
	found := false
	for _, item := range itemList.GetItems() {
		if item.Index == docIndex {
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("unable to find document")
	}

	return nil, nil

}
