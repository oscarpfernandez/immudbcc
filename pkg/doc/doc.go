package doc

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	immuapi "github.com/codenotary/immudb/pkg/api"
)

type PropertyEntry struct {
	KeyURI string
	Value  []byte
}

func (p PropertyEntry) DissectKeyURI() (string, []string, string) {
	keys := strings.Split(p.KeyURI, "/")
	lastElemIdx := len(keys) - 1

	return keys[0], keys[1:lastElemIdx], keys[lastElemIdx]
}

// PropertyEntryList defines a list of property key-value pairs.
type PropertyEntryList []PropertyEntry

func (p PropertyEntryList) Len() int {
	return len(p)
}
func (p PropertyEntryList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PropertyEntryList) Less(i, j int) bool {
	return strings.Compare(p[i].KeyURI, p[j].KeyURI) <= 0
}

type PropertyHash struct {
	Index uint64 // Index of property DB entry.
	Hash  []byte // Hash of property DB entry.
}

func PropertyHashDigest(index uint64, key, value []byte) *PropertyHash {
	digest := immuapi.Digest(index, key, value)

	return &PropertyHash{Index: index, Hash: digest[:]}
}

// Properties defined a list of property index hashes pairs.
type PropertyHashList []PropertyHash

// GlobalHash defines the global document hash.
type GlobalHash []byte

func GeneratePropertyList(docID string, r io.Reader) (PropertyEntryList, error) {
	var docMap map[string]interface{}
	if err := json.NewDecoder(r).Decode(&docMap); err != nil {
		return nil, fmt.Errorf("unable to unmarshall payload: %v", err)
	}

	return RawToPropertyList([]string{docID}, docMap), nil
}
