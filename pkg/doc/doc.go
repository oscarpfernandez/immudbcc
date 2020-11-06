package doc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

type ObjectManifest struct {
	ObjectID        string   `json:"object_id"`
	PropertyIndexes []uint64 `json:"property_indexes"`
	ObjectHash      Hash     `json:"object_hash"`
	ObjectEncHash   EncHash  `json:"object_enc_hash"`
}

type Hash []byte

func (h Hash) Encrypt(token string) (EncHash, error) {
	key, err := hex.DecodeString(token)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aesGCM.Seal(nonce, nonce, h, nil), nil
}

type EncHash []byte

func (e EncHash) Decrypt(token string) (Hash, error) {
	block, err := aes.NewCipher([]byte(token))
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()

	nonce, ciphertext := e[:nonceSize], e[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
