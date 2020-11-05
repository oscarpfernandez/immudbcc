package api

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/oscarpfernandez/immudbcc/pkg/doc"
	"github.com/oscarpfernandez/immudbcc/pkg/worker"

	immuclient "github.com/codenotary/immudb/pkg/client"
)

const (
	defaultNumWorkers = 20
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

// Manager represents the object required to use the API.
type Manager struct {
	numberWorkers int
	client        immuclient.ImmuClient
}

// New creates a new API manager object.
func New(c *Config) (*Manager, error) {
	client, err := immuclient.NewImmuClient(c.ClientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create ImmuDB client: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	return &Manager{
		numberWorkers: c.NumberWorkers,
		client:        client,
	}, nil
}

// StoreDocument saves a JSON document in the database, marshalling its structure
// into key-value properties, representing the transversal property paths of the
// original object.
func (m *Manager) StoreDocument(ctx context.Context, docID string, r io.Reader) ([]*doc.PropertyHash, error) {
	entryList, err := doc.GeneratePropertyList(docID, r)
	if err != nil {
		return nil, err
	}

	doc.PrintPropertyEntryList(entryList)

	workers := worker.NewWriteWorkerPool(m.numberWorkers, m.client)
	if err := workers.StartWorkers(ctx); err != nil {
		return nil, err
	}
	defer workers.Stop()

	// Process the entry list submitting them to the available workers.
	resultChan, done, errChan := workers.Write(entryList)

	counter := 0
	var resultHash []*doc.PropertyHash
	var errList []string
	for {
		select {
		case hash := <-resultChan:
			if hash != nil {
				resultHash = append(resultHash, hash)
				counter++
				fmt.Printf("Received result hash: %d\n", counter)
				if counter == len(entryList) {
					workers.Stop()
				}
			}
		case err := <-errChan:
			if err != nil {
				errList = append(errList, err.Error())
			}
		case <-done:
			goto finish
		case <-ctx.Done():
			break
		}
	}

finish:

	if len(errList) > 0 {
		return nil, fmt.Errorf("failed to store document ID '%s': %v", docID, strings.Join(errList, "; "))
	}

	return resultHash, nil
}
