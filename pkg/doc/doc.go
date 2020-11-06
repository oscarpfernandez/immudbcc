package doc

import (
	"crypto/sha256"
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

type Hash []byte
type EncHash []byte

type ObjectManifest struct {
	ObjectID        string   `json:"object_id"`
	PropertyIndexes []uint64 `json:"property_indexes"`
	ObjectHash      Hash     `json:"object_hash"`
	ObjectEncHash   EncHash  `json:"object_enc_hash"`
}

type PropertyHash struct {
	Index uint64 // Index of property DB entry.
	Hash  []byte // Hash of property DB entry.
}

func CreatePropertyHash(index uint64, key, value []byte) *PropertyHash {
	digest := immuapi.Digest(index, key, value)

	return &PropertyHash{Index: index, Hash: digest[:]}
}

// Properties defined a list of property index hashes pairs.
type PropertyHashList []*PropertyHash

func (p PropertyHashList) Len() int {
	return len(p)
}

func (p PropertyHashList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PropertyHashList) Less(i, j int) bool {
	return p[i].Index <= p[j].Index
}

func (p PropertyHashList) Hash() Hash {
	globalSum := sha256.New()
	for _, hash := range p {
		globalSum.Write(hash.Hash)
	}

	return globalSum.Sum(nil)
}

func (p PropertyHashList) Indexes() []uint64 {
	indexes := make([]uint64, len(p))
	for idx, pp := range p {
		indexes[idx] = pp.Index
	}

	return indexes
}
