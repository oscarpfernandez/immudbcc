package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/oscarpfernandez/immudbcc/pkg/doc"

	"github.com/codenotary/immudb/pkg/api/schema"
	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/stretchr/testify/assert"
)

// ImmuClientMock defines an inversion of control mock.
type ImmuClientMock struct {
	immuclient.ImmuClient
	setFn     func(ctx context.Context, key []byte, value []byte) (*immuschema.Index, error)
	getFn     func(ctx context.Context, key []byte) (*schema.StructuredItem, error)
	byIndexFn func(ctx context.Context, index uint64) (*schema.StructuredItem, error)
}

func (m *ImmuClientMock) Set(ctx context.Context, key []byte, value []byte) (*immuschema.Index, error) {
	return m.setFn(ctx, key, value)
}

func (m *ImmuClientMock) Get(ctx context.Context, key []byte) (*schema.StructuredItem, error) {
	return m.getFn(ctx, key)
}

func (m *ImmuClientMock) ByIndex(ctx context.Context, index uint64) (*schema.StructuredItem, error) {
	return m.byIndexFn(ctx, index)
}

func TestManager_StoreDocument(t *testing.T) {
	type KeyValue struct {
		Index uint64
		Key   string
		Value []byte
	}

	tests := map[string]struct {
		jsonPayload         []byte
		expStoredProperties []KeyValue
		expResult           *StoreDocumentResult
	}{
		"Stored document": {
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
					"hash":"8a882c474f519a42bbf13adc6f5b0343ba56afa162bdc110078a9c6c49cba9de"
				}`)},
			},
			expResult: &StoreDocumentResult{
				Index: 25,
				Hash:  "",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var gotStoredProperties []KeyValue
			var index uint64 = 0
			clientMock := &ImmuClientMock{
				setFn: func(ctx context.Context, key []byte, value []byte) (*immuschema.Index, error) {
					gotStoredProperties = append(gotStoredProperties, KeyValue{Index: index, Key: string(key), Value: value})
					defer func() { index++ }()
					return &immuschema.Index{Index: index}, nil
				},
				getFn: func(ctx context.Context, key []byte) (*immuschema.StructuredItem, error) {
					for idx, kv := range gotStoredProperties {
						if kv.Key == string(key) {
							return &immuschema.StructuredItem{
								Key:   []byte(kv.Key),
								Value: &immuschema.Content{Payload: kv.Value},
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

			conf := DefaultConfig().WithNumberWorkers(5)
			manager := Manager{
				conf:   *conf,
				client: clientMock,
			}

			storeResult, err := manager.StoreDocument(context.Background(), "docID", bytes.NewReader(test.jsonPayload))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			getResult, err := manager.GetDocument(context.Background(), "docID")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, storeResult.Hash, getResult.Hash)
			assert.JSONEq(t, string(test.jsonPayload), string(getResult.Payload), "stored and retrieved payloads should match")
		})
	}
}

func MustMarshall(s interface{}) string {
	payload, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(payload)
}
