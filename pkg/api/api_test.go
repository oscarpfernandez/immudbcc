package api

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/oscarpfernandez/immudbcc/pkg/doc"

	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// Ensure that the ImmuClientMock implements ImmuClient interface.
var _ immuclient.ImmuClient = &ImmuClientMock{}

// ImmuClientMock defines an inversion of control mock.
type ImmuClientMock struct {
	mu *sync.RWMutex
	immuclient.ImmuClient
	safeSetFn func(ctx context.Context, key []byte, value []byte) (*immuclient.VerifiedIndex, error)
	safeGetFn func(ctx context.Context, key []byte, opts ...grpc.CallOption) (*immuclient.VerifiedItem, error)
	byIndexFn func(ctx context.Context, index uint64) (*immuschema.StructuredItem, error)
}

func (m *ImmuClientMock) SafeSet(ctx context.Context, key []byte, value []byte) (*immuclient.VerifiedIndex, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.safeSetFn(ctx, key, value)
}

func (m *ImmuClientMock) SafeGet(ctx context.Context, key []byte, opts ...grpc.CallOption) (*immuclient.VerifiedItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.safeGetFn(ctx, key)
}

func (m *ImmuClientMock) ByIndex(ctx context.Context, index uint64) (*immuschema.StructuredItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.byIndexFn(ctx, index)
}

func TestManagerStoreGetDocument(t *testing.T) {
	type KeyValue struct {
		Index uint64
		Key   string
		Value []byte
	}

	tests := map[string]struct {
		jsonPayload         []byte
		expStoredProperties []KeyValue
		expObjectManifest   *ObjectManifest
	}{
		"Stored document #1": {
			jsonPayload: []byte(`{
			  "squadName": "Super hero squad",
			  "homeTown": "Metro City",
			  "formed": 2016,
			  "secretBase": "Super tower",
			  "active": true,
			  "members": [
				{
				  "name": "Molecule Man",
				  "age": 29,
				  "secretIdentity": "Dan Jukes",
				  "powers": [
					"Radiation resistance",
					"Turning tiny",
					"Radiation blast"
				  ]
				},
				{
				  "name": "Madame Uppercut",
				  "age": 39,
				  "secretIdentity": "Jane Wilson",
				  "powers": [
					"Million tonne punch",
					"Damage resistance",
					"Superhuman reflexes"
				  ]
				},
				{
				  "name": "Eternal Flame",
				  "age": 1000000,
				  "secretIdentity": "Unknown",
				  "powers": [
					"Immortality",
					"Heat Immunity",
					"Inferno",
					"Teleportation",
					"Interdimensional travel"
				  ]
				}
			  ]
			}`),
			expStoredProperties: []KeyValue{
				{Key: "docID/secretBase/string", Value: []byte("Super tower")},
				{Key: "docID/members/[0.3]/powers/[0.3]/string", Value: []byte("Radiation resistance")},
				{Key: "docID/members/[0.3]/powers/[1.3]/string", Value: []byte("Turning tiny")},
				{Key: "docID/members/[0.3]/powers/[2.3]/string", Value: []byte("Radiation blast")},
				{Key: "docID/members/[0.3]/name/string", Value: []byte("Molecule Man")},
				{Key: "docID/members/[0.3]/age/float64", Value: doc.Float64ToBinary(29)},
				{Key: "docID/members/[0.3]/secretIdentity/string", Value: []byte("Dan Jukes")},
				{Key: "docID/members/[1.3]/name/string", Value: []byte("Madame Uppercut")},
				{Key: "docID/members/[1.3]/age/float64", Value: doc.Float64ToBinary(39)},
				{Key: "docID/members/[1.3]/secretIdentity/string", Value: []byte("Jane Wilson")},
				{Key: "docID/active/bool", Value: []byte("true")},
				{Key: "docID/members/[1.3]/powers/[1.3]/string", Value: []byte("Damage resistance")},
				{Key: "docID/members/[1.3]/powers/[2.3]/string", Value: []byte("Superhuman reflexes")},
				{Key: "docID/members/[2.3]/age/float64", Value: doc.Float64ToBinary(1000)},
				{Key: "docID/members/[2.3]/secretIdentity/string", Value: []byte("Unknown")},
				{Key: "docID/members/[2.3]/powers/[0.5]/string", Value: []byte("Immortality")},
				{Key: "docID/members/[2.3]/powers/[1.5]/string", Value: []byte("Heat Immunity")},
				{Key: "docID/members/[2.3]/powers/[2.5]/string", Value: []byte("Inferno")},
				{Key: "docID/members/[2.3]/powers/[3.5]/string", Value: []byte("Teleportation")},
				{Key: "docID/members/[2.3]/powers/[4.5]/string", Value: []byte("Interdimensional travel")},
				{Key: "docID/members/[2.3]/name/string", Value: []byte("Eternal Flame")},
				{Key: "docID/formed/float64", Value: doc.Float64ToBinary(2016)},
				{Key: "docID/squadName/string", Value: []byte("Super hero squad")},
				{Key: "docID/homeTown/string", Value: []byte("Metro City")},
				{Key: "docID/members/[1.3]/powers/[0.3]/string", Value: []byte("Million tonne punch")},
				{Key: "manifest/docID", Value: []byte(`{"id":"docID",
					"indexes":[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24],
					"hash":"51175166a32086593b4bdc5d992acca5b45fbaef487c297759e986e603227f27"
				}`)},
			},
			expObjectManifest: &ObjectManifest{
				ObjectID: "docID",
				Indexes:  []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24},
				Hash:     "51175166a32086593b4bdc5d992acca5b45fbaef487c297759e986e603227f27",
			},
		},
		"Stored document #2": {
			jsonPayload: []byte(`{
				"people": [
					{"id": 0,"name": "Monroe Roth"},
					{"id": 1,"name": "Mullen Rhodes"},
					{"id": 2,"name": "Mcclure Welch"}
				]
			}`),
			expStoredProperties: []KeyValue{
				{Key: "docID/people/[0.3]/id/float64", Value: doc.Float64ToBinary(0)},
				{Key: "docID/people/[0.3]/name/string", Value: []byte("Monroe Roth")},
				{Key: "docID/people/[1.3]/id/float64", Value: doc.Float64ToBinary(1)},
				{Key: "docID/people/[1.3]/name/string", Value: []byte("Mullen Rhodes")},
				{Key: "docID/people/[2.3]/id/float64", Value: doc.Float64ToBinary(2)},
				{Key: "docID/people/[2.3]/name/string", Value: []byte("Mcclure Welch")},
				{Key: "manifest/docID", Value: []byte(`{"id":"docID",
					"indexes":[0,1,2,3,4,5],
					"hash":"61486e6661a9fe498c8ff50b0f5232e3d2979328faa6316346795067acc0eee6"
				}`)},
			},
			expObjectManifest: &ObjectManifest{
				ObjectID: "docID",
				Indexes:  []uint64{0, 1, 2, 3, 4, 5},
				Hash:     "61486e6661a9fe498c8ff50b0f5232e3d2979328faa6316346795067acc0eee6",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var gotStoredProperties []KeyValue
			var index uint64 = 0

			// Define ImmuDB client mock.
			clientMock := &ImmuClientMock{
				mu: &sync.RWMutex{},
				safeSetFn: func(ctx context.Context, key []byte, value []byte) (*immuclient.VerifiedIndex, error) {
					gotStoredProperties = append(gotStoredProperties, KeyValue{Index: index, Key: string(key), Value: value})
					defer func() { index++ }()
					return &immuclient.VerifiedIndex{Index: index}, nil
				},
				safeGetFn: func(ctx context.Context, key []byte, opts ...grpc.CallOption) (*immuclient.VerifiedItem, error) {
					for idx, kv := range gotStoredProperties {
						if kv.Key == string(key) {
							return &immuclient.VerifiedItem{
								Key:   []byte(kv.Key),
								Value: kv.Value,
								Index: uint64(idx),
							}, nil
						}
					}
					return nil, errors.New("not found")
				},
				byIndexFn: func(ctx context.Context, index uint64) (*immuschema.StructuredItem, error) {
					if index < uint64(len(gotStoredProperties)-1) {
						return &immuschema.StructuredItem{
							Index: index,
							Key:   []byte(gotStoredProperties[index].Key),
							Value: &immuschema.Content{
								Payload: gotStoredProperties[index].Value,
							},
						}, nil
					}
					return nil, errors.New("not found")
				},
			}

			conf := DefaultConfig().WithNumberWorkers(1)
			manager := Manager{
				conf:   *conf,
				client: clientMock,
			}

			storeResult, err := manager.StoreDocument(context.Background(), "docID", bytes.NewReader(test.jsonPayload))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			details, err := manager.getDocumentDetails(context.Background(), "docID")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assert.Equal(t, test.expObjectManifest, details.objectManifest)

			getResult, err := manager.GetDocument(context.Background(), "docID")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			isValid, err := manager.VerifyDocument(context.Background(), "docID", test.expObjectManifest.Hash)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assert.True(t, isValid)

			assert.JSONEq(t, string(test.jsonPayload), string(getResult.Payload))
			assert.Equal(t, storeResult.Hash, getResult.Hash)
		})
	}
}
