package doc

import (
	"encoding/json"
	"fmt"
	"io"

	immuclient "github.com/codenotary/immudb/pkg/client"

	_ "github.com/codenotary/immudb/pkg/client"
)

// PropertyEntryList defines a list of property key-value pairs.
type PropertyEntryList []PropertyEntry

// Properties defined a list of property index hashes pairs.
type PropertyHashList []PropertyHash

// GlobalHash defines the global document hash.
type GlobalHash []byte

type DocumentManager struct {
	dbClient immuclient.ImmuClient
	docID    string
	docName  string

	PropertyEntryList PropertyEntryList
	PropertyHashList  PropertyHashList
	GlobalHash        GlobalHash
}

type PropertyEntry struct {
	KeyURI string
	Value  []byte
}

type PropertyHash struct {
	Index uint64 // Index of property DB entry.
	Hash  []byte // Hash of property DB entry.
}

func NewDocManager(docID, docName string, client immuclient.ImmuClient) *DocumentManager {
	return &DocumentManager{
		dbClient: client,
		docID:    docID,
		docName:  docName,
	}
}

func (d *DocumentManager) GeneratePropertyList(r io.Reader) error {
	var docMap map[string]interface{}
	if err := json.NewDecoder(r).Decode(&docMap); err != nil {
		return fmt.Errorf("unable to unmarshall payload: %v", err)
	}

	// https://www.w3schools.com/js/js_json_datatypes.asp
	for key, value := range docMap {
		switch value.(type) {
		case nil:

		case string:

		case int64:

		case uint64:

		case float32:

		case float64:

		case bool:

		case []interface{}:

		case map[string]interface{}:

		case []map[string]interface{}:

		}

		_ = key
	}

	return nil
}
