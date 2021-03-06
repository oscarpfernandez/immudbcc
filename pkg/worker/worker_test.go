package worker

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/oscarpfernandez/immudbcc/pkg/doc"

	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/stretchr/testify/assert"
)

type ImmuClientMock struct {
	wg *sync.Mutex
	immuclient.ImmuClient
	safeSetFn func(ctx context.Context, key []byte, value []byte) (*immuclient.VerifiedIndex, error)
}

func (m *ImmuClientMock) SafeSet(ctx context.Context, key []byte, value []byte) (*immuclient.VerifiedIndex, error) {
	m.wg.Lock()
	defer m.wg.Unlock()
	return m.safeSetFn(ctx, key, value)
}

func TestWorker(t *testing.T) {
	tests := map[string]struct {
		numWorkers      int
		forceWriteError bool
		properties      doc.PropertyEntryList
		expErr          []string
	}{
		"Successful write": {
			numWorkers: 5,
			properties: doc.PropertyEntryList{
				{KeyURI: "prefix2/cars/car1/string", Value: []byte(`Ford`)},
				{KeyURI: "prefix2/cars/car2/string", Value: []byte(`BMW`)},
				{KeyURI: "prefix2/cars/car3/string", Value: []byte(`Fiat`)},
				{KeyURI: "prefix2/name/string", Value: []byte(`John`)},
				{KeyURI: "prefix2/age/float64", Value: doc.Float64ToBinary(30)},
			},
			forceWriteError: false,
		},
		"Failed write": {
			numWorkers: 5,
			properties: doc.PropertyEntryList{
				{KeyURI: "prefix2/cars/car1/string", Value: []byte(`Ford`)},
				{KeyURI: "prefix2/cars/car2/string", Value: []byte(`BMW`)},
				{KeyURI: "prefix2/cars/car3/string", Value: []byte(`Fiat`)},
				{KeyURI: "prefix2/name/string", Value: []byte(`John`)},
				{KeyURI: "prefix2/age/float64", Value: doc.Float64ToBinary(30)},
			},
			forceWriteError: true,
			expErr:          []string{"write error 1", "write error 2", "write error 3", "write error 4", "write error 5"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			counter := 0
			resultHashList := doc.PropertyHashList{}
			var errList []string

			func() {
				index := 0
				mock := &ImmuClientMock{
					wg: &sync.Mutex{},
					safeSetFn: func(ctx context.Context, key []byte, value []byte) (*immuclient.VerifiedIndex, error) {
						index++
						if test.forceWriteError {
							return nil, fmt.Errorf("write error %d", index)
						}
						return &immuclient.VerifiedIndex{Index: uint64(index)}, nil
					},
				}

				workers := NewWriteWorkerPool(test.numWorkers, mock)
				err := workers.StartWorkers(context.Background())
				assert.Nil(t, err)
				err = workers.StartWorkers(context.Background()) // second start should take no effect.
				assert.EqualError(t, err, "workers are already started")
				defer func() {
					workers.Stop()
					workers.Stop() // double stop should no crash.
				}()

				resultChan, _, errChan := workers.Write(test.properties)

				for {
					select {
					case hash := <-resultChan:
						if hash != nil {
							counter++
							resultHashList = append(resultHashList, hash)
							if counter == len(test.properties) {
								return
							}
						}
					case err := <-errChan:
						if err != nil {
							counter++
							errList = append(errList, err.Error())
							if counter == len(test.properties) {
								return
							}
						}
					}
				}
			}()
			assert.ElementsMatch(t, test.expErr, errList)
		})
	}
}
