package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/oscarpfernandez/immudbcc/pkg/doc"
)

type WriteWorkerPool struct {
	numWorkers   int
	client       immuclient.ImmuClient
	jobChan      chan *doc.PropertyEntry
	resultChan   chan *doc.PropertyHash
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
		jobChan:      make(chan *doc.PropertyEntry, 100),
		resultChan:   make(chan *doc.PropertyHash, 100),
		errChan:      make(chan error, 100),
		shutdownChan: make(chan bool),
		wg:           &sync.WaitGroup{},
		mu:           &sync.Mutex{},
	}
}

func (w *WriteWorkerPool) StartWorkers(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isStarted {
		return errors.New("workers are already started")
	}

	w.wg.Add(w.numWorkers)
	for i := 0; i < w.numWorkers; i++ {
		log.Printf("Starting worker %d", i)
		go w.worker(ctx)
	}
	w.isStarted = true

	return nil
}

func (w *WriteWorkerPool) Write(properties doc.PropertyEntryList) (<-chan *doc.PropertyHash, <-chan bool, <-chan error) {
	go func() {
		for idx, propEntry := range properties {
			fmt.Printf("Sending job %d\n", idx)
			w.jobChan <- &propEntry
		}
	}()

	return w.resultChan, w.shutdownChan, w.errChan
}

func (w *WriteWorkerPool) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isStarted {
		return
	}

	fmt.Println("Waiting for go routines to finish")

	w.closeOnce.Do(func() {
		close(w.shutdownChan)
		w.wg.Wait()
		close(w.jobChan) // Close the underlying channels.
		close(w.resultChan)
		close(w.errChan)
	})
}

func (w *WriteWorkerPool) worker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case job := <-w.jobChan:
			if job != nil {
				func() {
					key, value := []byte(job.KeyURI), job.Value

					vi, err := w.client.SafeSet(ctx, key, value)
					if err != nil {
						fmt.Printf("Error :%v", err)
						w.errChan <- err
					}
					fmt.Printf("Stored: %v\n", vi.Index)
					w.resultChan <- doc.PropertyHashDigest(vi.Index, key, value)
				}()
			}

		case <-w.shutdownChan:
			fmt.Println("Shutdown called")
			return

		case <-ctx.Done():
			return
		}
	}
}
