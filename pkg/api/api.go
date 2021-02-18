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

// WithClientOptions set the client options used to initialize the ImmuDB client.
func (c *Config) WithClientOptions(options *immuclient.Options) *Config {
	c.ClientOptions = options
	return c
}

// ObjectManifest defines the top level object that describes a document in the
// Database, including the object ID of said document, the indexes of each of its
// properties and the global hash of the document (comprised by the hash of hashes,
// sorted according to the associated property index).
type ObjectManifest struct {
	ObjectID string   `json:"id"`
	Indexes  []uint64 `json:"indexes"`
	Hash     string   `json:"hash"`
}

// StoreDocumentResult represents the insertion result of a document.
type StoreDocumentResult struct {
	Index uint64
	Hash  string
}

// GetDocumentResult represents the result of fetching a document.
type GetDocumentResult struct {
	ID      string
	Index   uint64
	Payload []byte
	Hash    string
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

// StoreDocument saves a JSON document in the database, marshaling its structure
// into key-value properties, representing the transversal property paths of the
// original object.
func (m *Manager) StoreDocument(ctx context.Context, docID string, r io.Reader) (*StoreDocumentResult, error) {
	entryList, err := doc.RawToPropertyList(docID, r)
	if err != nil {
		return nil, err
	}

	sort.Sort(entryList)

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

	manifest := &ObjectManifest{
		ObjectID: docID,
		Indexes:  resultHash.Indexes(),
		Hash:     resultHash.Hash(),
	}

	index, err := m.writeDocumentManifest(ctx, manifest)
	if err != nil {
		return nil, fmt.Errorf("unable to store manifes of object '%s': %v", docID, err)
	}

	log.Printf("Object Write succesfull: index(%d) - keyID(%s)", index, docID)

	return &StoreDocumentResult{
		Index: index,
		Hash:  manifest.Hash,
	}, nil
}

// GetDocument allows the extraction of a document provided its global ID.
func (m *Manager) GetDocument(ctx context.Context, docId string) (*GetDocumentResult, error) {
	docDetails, err := m.getDocumentDetails(ctx, docId)
	if err != nil {
		return nil, err
	}

	log.Print("Reconstructing JSON object...")
	rawObject := doc.PropertyListToRaw(docDetails.propertyEntryList)
	payload, err := json.MarshalIndent(rawObject, "", "  ")
	if err != nil {
		return nil, err
	}

	log.Printf("Object Read succesfull: index(%d) - keyID(%s)", docDetails.objectManifestIndex, docDetails.objectManifestKey)

	return &GetDocumentResult{
		ID:      docId,
		Payload: payload,
		Index:   docDetails.objectManifestIndex,
		Hash:    docDetails.propertyHashList.Hash(),
	}, nil
}

// documentDetails represents the underlying properties required to describe
// and store a document.
type documentDetails struct {
	objectManifestIndex uint64
	objectManifestKey   string
	objectManifest      *ObjectManifest
	propertyEntryList   doc.PropertyEntryList
	propertyHashList    doc.PropertyHashList
}

// getDocumentDetails fetches from the database the details of a given document.
func (m *Manager) getDocumentDetails(ctx context.Context, docId string) (*documentDetails, error) {
	docManifestKey := []byte("manifest/" + docId)

	log.Printf("Reading object objectManifest: DocumentID(%s)", docManifestKey)
	docManifestItem, err := m.client.SafeGet(ctx, docManifestKey)
	if err != nil {
		return nil, err
	}
	log.Printf("Object objectManifest: Index(%d) - Key(%s)", docManifestItem.Index, string(docManifestItem.Key))

	objectManifest := &ObjectManifest{}
	if err := json.Unmarshal(docManifestItem.Value, objectManifest); err != nil {
		fmt.Printf("unmarshal failed")
		return nil, err
	}
	log.Printf("Object objectManifest: Key(%s) - Indexes(%v)", string(docManifestItem.Key), objectManifest.Indexes)

	propertyList := doc.PropertyEntryList{}
	propertyHashList := doc.PropertyHashList{}
	for _, propertyIndex := range objectManifest.Indexes {
		object, err := m.client.ByIndex(ctx, propertyIndex)
		if err != nil {
			return nil, err
		}
		log.Printf("Reading property: Index(%d) - Key(%s)", object.Index, object.Key)

		propertyList = append(propertyList, doc.PropertyEntry{
			KeyURI: string(object.Key),
			Value:  object.Value.Payload,
		})
		hash := doc.CreatePropertyHash(object.Index, object.Key, object.Value.GetPayload())
		propertyHashList = append(propertyHashList, hash)
	}

	sort.Sort(propertyHashList)

	return &documentDetails{
		objectManifestIndex: docManifestItem.Index,
		objectManifestKey:   string(docManifestItem.Key),
		objectManifest:      objectManifest,
		propertyEntryList:   propertyList,
		propertyHashList:    propertyHashList,
	}, nil
}

// VerifyDocument ensures that the stored document hash matches a known global
// hash, returning True if the integrity of the document is ensured, and False
// otherwise.
func (m *Manager) VerifyDocument(ctx context.Context, docID, globalHash string) (bool, error) {
	result, err := m.getDocumentDetails(ctx, docID)
	if err != nil {
		return false, err
	}

	if result.propertyHashList.Hash() == globalHash {
		return true, nil
	}

	return false, nil
}

// UpdateDocument allows the update of a given property of a document.
// Here the underlying assumption for the implementation is that updates
// are fairly rare and limited in scope.
func (m *Manager) UpdateDocument(ctx context.Context, docID string, key string, value []byte) (*GetDocumentResult, error) {
	docDetails, err := m.getDocumentDetails(ctx, docID)
	if err != nil {
		return nil, err
	}

	propertyKey := docID + "/" + key

	hashList := docDetails.propertyHashList
	manifest := docDetails.objectManifest

	found := false
	// Search for the property in the object manifest, and only replace that.
	for i, hash := range hashList {
		if hash.Key == propertyKey {
			oldDBIndex := hash.Index

			// Set the new property.
			idx, err := m.client.SafeSet(ctx, []byte(propertyKey), value)
			if err != nil {
				return nil, err
			}

			// Update the hash list to include the new property.
			newDBIndex := idx.Index
			docDetails.propertyHashList[i] = doc.CreatePropertyHash(newDBIndex, []byte(key), value)

			manifIndexes := manifest.Indexes
			for pos, idx := range manifIndexes {
				if idx == oldDBIndex {
					manifIndexes[pos] = newDBIndex
				}
			}
			// Update the manifest's hash list.
			manifest.Indexes = manifIndexes

			// Update the manifest's global hash.
			manifest.Hash = hashList.Hash()

			found = true
			break
		}
	}

	// Could not find and updated property.
	if !found {
		return nil, fmt.Errorf("document docID=%s does not have key=%s", key, docID)
	}

	// Save the new object manifest.
	index, err := m.writeDocumentManifest(ctx, manifest)
	if err != nil {
		return nil, fmt.Errorf("unable to store manifes of object '%s': %v", docID, err)
	}

	return &GetDocumentResult{
		ID:    docID,
		Index: index,
		Hash:  manifest.Hash,
	}, nil
}

// writeDocumentManifest persists in the Database the document manifest descriptor.
func (m *Manager) writeDocumentManifest(ctx context.Context, om *ObjectManifest) (uint64, error) {
	objectManifestKey := []byte(fmt.Sprintf("manifest/%s", om.ObjectID))

	documentValue, err := json.Marshal(om)
	if err != nil {
		return 0, fmt.Errorf("unable to marshall object maifest: %v", err)
	}

	idx, err := m.client.SafeSet(ctx, objectManifestKey, documentValue)
	if err != nil {
		return 0, err
	}
	return idx.Index, nil
}
