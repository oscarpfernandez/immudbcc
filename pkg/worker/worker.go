package worker

import (
	"context"
	"sync"

	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/oscarpfernandez/immudbcc/pkg/doc"
)

type WriteWorkerPool struct {
	numWorkers   int
	client       immuclient.ImmuClient
	jobChan      chan doc.PropertyEntry
	resultChan   chan doc.PropertyHash
	errChan      chan error
	shutdownChan chan bool

	mu        *sync.Mutex
	wg        *sync.WaitGroup
	isStarted bool
	closeOnce sync.Once
}

func NewWriteWorkerPool(numWorkers int, client immuclient.ImmuClient) *WriteWorkerPool {
	return &WriteWorkerPool{
		numWorkers:   numWorkers,
		client:       client,
		jobChan:      make(chan doc.PropertyEntry, 100),
		resultChan:   make(chan doc.PropertyHash, 100),
		errChan:      make(chan error, 100),
		shutdownChan: make(chan bool),
		wg:           &sync.WaitGroup{},
	}
}

func (w *WriteWorkerPool) Start(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isStarted {
		return
	}

	for i := 0; i < w.numWorkers; i++ {
		w.wg.Add(1)
		go w.worker(ctx)
	}
	w.isStarted = true
}

func (w *WriteWorkerPool) Write(properties doc.PropertyEntryList) <-chan error {
	go func() {
		for _, propEntry := range properties {
			w.jobChan <- propEntry
		}
	}()

	return w.errChan
}

func (w *WriteWorkerPool) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isStarted {
		return
	}

	w.closeOnce.Do(func() {
		close(w.shutdownChan) // Trigger workers shutdown.
		w.wg.Wait()           // Wait until all workers are stopped.
		close(w.jobChan)      // Close the underlying channels.
		close(w.resultChan)
		close(w.errChan)
	})
}

func (w *WriteWorkerPool) worker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case job := <-w.jobChan:
			func() {
				key, value := []byte(job.KeyURI), job.Value

				vi, err := w.client.SafeSet(ctx, key, value)
				if err != nil {
					w.errChan <- err
					return
				}

				w.resultChan <- doc.PropertyHashDigest(vi.Index, key, value)
			}()

		case <-w.shutdownChan:
			return

		case <-ctx.Done():
			return
		}
	}
}
