package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/oscarpfernandez/immudbcc/pkg/doc"
	"github.com/oscarpfernandez/immudbcc/pkg/worker"

	immuclient "github.com/codenotary/immudb/pkg/client"
)

const (
	defaultNumWorkers = 50
)

// Config represents the required API options.
type Config struct {
	NumberWorkers int
	ClientOptions *immuclient.Options
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

	workers := worker.NewWriteWorkerPool(m.conf.NumberWorkers, m.client)
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
				// The goroutines were shutdown.
				return
			case <-ctx.Done():
				// Execution context expired.
				return
			}
		}
	}()

	if len(errList) > 0 {
		return nil, fmt.Errorf("failed to store document ID '%s': %v", docID, strings.Join(errList, "; "))
	}

	sort.Sort(resultHash)

	objectManifest := &doc.ObjectManifest{
		ObjectID:        docID,
		PropertyIndexes: resultHash.Indexes(),
		ObjectHash:      resultHash.Hash(),
	}

	index, err := m.writeDocumentManifest(ctx, objectManifest)
	if err != nil {
		return nil, fmt.Errorf("unable to store manifes of object '%s': %v", docID, err)
	}

	log.Printf("Object Write succesfull: index(%d) - keyID(%s)", index, docID)

	return &StoreDocumentResult{
		Index: index,
		Hash:  objectManifest.ObjectHash,
	}, nil
}

func (m *Manager) writeDocumentManifest(ctx context.Context, om *doc.ObjectManifest) (uint64, error) {
	objectManifestKey := []byte(fmt.Sprintf("manifest/%s", om.ObjectID))

	documentValue, err := json.Marshal(om)
	if err != nil {
		return 0, fmt.Errorf("unable to marshall object maifest: %v", err)
	}

	idx, err := m.client.Set(ctx, objectManifestKey, documentValue)
	if err != nil {
		return 0, err
	}
	return idx.Index, nil
}

type GetDocumentResult struct {
	ID      string
	Index   uint64
	Payload []byte
	Hash    []byte
}

func (m *Manager) GetDocument(ctx context.Context, docId string) (*GetDocumentResult, error) {
	docManifestKey := []byte("manifest/" + docId)

	log.Printf("Reading object manifest: DocumentID(%s)", docManifestKey)
	docManifestItem, err := m.client.Get(ctx, docManifestKey)
	if err != nil {
		return nil, err
	}
	log.Printf("Object manifest: Index(%d) - Key(%s)", docManifestItem.Index, string(docManifestItem.Key))

	objectManifest := &doc.ObjectManifest{}
	if err := json.Unmarshal(docManifestItem.Value.GetPayload(), objectManifest); err != nil {
		fmt.Printf("unmarshal failed")
		return nil, err
	}
	log.Printf("Object manifest: Key(%s) - PropertyIndexes(%v)", string(docManifestItem.Key), objectManifest.PropertyIndexes)

	propertyList := doc.PropertyEntryList{}
	for _, propertyIndex := range objectManifest.PropertyIndexes {
		object, err := m.client.ByIndex(ctx, propertyIndex)
		if err != nil {
			return nil, err
		}
		log.Printf("Reading property: Index(%d) - Key(%s)", object.Index, object.Key)

		propertyList = append(propertyList, doc.PropertyEntry{
			KeyURI: string(object.Key),
			Value:  object.Value.Payload,
		})
	}

	log.Print("Reconstructing JSON object...")
	rawObject := doc.PropertyListToRaw(propertyList)
	payload, err := json.MarshalIndent(rawObject, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("Object Read succesfull: index(%d) - keyID(%s)", docManifestItem.Index, string(docManifestItem.Key))

	return &GetDocumentResult{
		ID:      docId,
		Index:   docManifestItem.Index,
		Payload: payload,
		Hash:    nil,
	}, nil
}
