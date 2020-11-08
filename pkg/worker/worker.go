package worker

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/oscarpfernandez/immudbcc/pkg/doc"

	immuclient "github.com/codenotary/immudb/pkg/client"
)

// WriteWorkerPool represents the pool of DB writer go routines.
type WriteWorkerPool struct {
	numWorkers   int
	isStarted    bool
	client       immuclient.ImmuClient
	jobChan      chan *doc.PropertyEntry
	resultChan   chan *doc.PropertyHash
	errChan      chan error
	shutdownChan chan bool

	mu        *sync.Mutex
	wg        *sync.WaitGroup
	closeOnce sync.Once
}

// NewWriteWorkerPool creates a new object.
func NewWriteWorkerPool(numWorkers int, client immuclient.ImmuClient) *WriteWorkerPool {
	return &WriteWorkerPool{
		numWorkers:   numWorkers,
		client:       client,
		jobChan:      make(chan *doc.PropertyEntry, 50),
		resultChan:   make(chan *doc.PropertyHash, 50),
		errChan:      make(chan error, 50),
		shutdownChan: make(chan bool),
		wg:           &sync.WaitGroup{},
		mu:           &sync.Mutex{},
	}
}

// StartWorkers launches the worker pool writer goroutines.
func (w *WriteWorkerPool) StartWorkers(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isStarted {
		return errors.New("workers are already started")
	}

	w.wg.Add(w.numWorkers)
	for i := 0; i < w.numWorkers; i++ {
		go w.worker(ctx)
	}
	w.isStarted = true

	return nil
}

// Write performs the write of a list of property entry list.
// Return three channels to handle the processing response results:
// * <-chan *doc.PropertyHash: a read channel of elements inserted in the DB.
// * <-chan bool: read channel used as a go routine termination signal.
// * <-chan error: read channel collecting any errors that might occur during
// the data ingestion.
func (w *WriteWorkerPool) Write(properties doc.PropertyEntryList) (<-chan *doc.PropertyHash, <-chan bool, <-chan error) {
	go func() {
		for _, propEntry := range properties {
			pp := propEntry // lock value.
			w.jobChan <- &pp
		}
	}()

	return w.resultChan, w.shutdownChan, w.errChan
}

// Stop triggers the shutdown of all goroutines within the pool.
func (w *WriteWorkerPool) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isStarted {
		return
	}

	w.closeOnce.Do(func() {
		close(w.shutdownChan)
		w.wg.Wait()
		close(w.jobChan) // Close the underlying channels.
		close(w.resultChan)
		close(w.errChan)
	})
}

// worker defines the worker's processing control loop that can be launched as
// a goroutine.
func (w *WriteWorkerPool) worker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case job := <-w.jobChan:
			if job != nil {
				key, value := []byte(job.KeyURI), job.Value
				log.Printf("Writing property: key(%s)", key)
				index, err := w.client.Set(ctx, key, value)
				if err != nil {
					w.errChan <- err
					continue
				}
				w.resultChan <- doc.CreatePropertyHash(index.Index, key, value)
			}
		case <-w.shutdownChan:
			return
		case <-ctx.Done():
			w.errChan <- errors.New("context expiration timeout")
			return
		}
	}
}
