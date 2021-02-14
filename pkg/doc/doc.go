package doc

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	immuapi "github.com/codenotary/immudb/pkg/api"
)

// PropertyEntry represents a given entry from a document, more specifically a
// given full path key and its associated value.
type PropertyEntry struct {
	KeyURI string
	Value  []byte
}

// DissectKeyURI returns the underlying parts of a given key format.
func (p PropertyEntry) DissectKeyURI() (string, []string, string) {
	if !hasKeyFormat(p.KeyURI) {
		panic(fmt.Sprintf("property '%s' has invalid format", p.KeyURI))
	}
	keys := strings.Split(p.KeyURI, "/")
	lastElemIdx := len(keys) - 1

	return keys[0], keys[1:lastElemIdx], keys[lastElemIdx]
}

// Key format: <docID>/<s>/(<s>/<s>)*/<type>
var keyRegExp = regexp.MustCompile(`^\S+\/\S+(\/\S+)*\/(?:nil|string|bool|float64)$`)

// hasKeyFormat checks if the current node of the path describes and array element.
func hasKeyFormat(s string) bool {
	// Checks for Arrays definitions of the format "[%d.%d]" where the first
	// parameter describes the current index and the second the total capacity.
	return keyRegExp.MatchString(s)
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

// PropertyHash represents a hash sum of a given property.
type PropertyHash struct {
	Index uint64 // Index of property DB entry.
	Key   string
	Hash  []byte // Hash of property DB entry.
}

// CreatePropertyHash returns a property hash of a given key-value pair.
func CreatePropertyHash(index uint64, key, value []byte) *PropertyHash {
	digest := immuapi.Digest(index, key, value)

	return &PropertyHash{
		Index: index,
		Key:   string(key),
		Hash:  digest[:],
	}
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

// Hash returns the global hash of a property hash list.
func (p PropertyHashList) Hash() string {
	globalSum := sha256.New()
	for _, hash := range p {
		_, _ = globalSum.Write(hash.Hash)
	}

	sum := globalSum.Sum(nil)

	return hex.EncodeToString(sum)
}

// Indexes returns the associated indexes of a given property hash list.
func (p PropertyHashList) Indexes() []uint64 {
	indexes := make([]uint64, len(p))
	for idx, pp := range p {
		indexes[idx] = pp.Index
	}

	return indexes
}

// PropertyNil converts a property path with null value to a PropertyEntry.
func PropertyNil(keys []string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/nil",
		Value:  nil,
	}
}
// PropertyString converts a property path with string value to a PropertyEntry.
func PropertyString(keys []string, value string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/string",
		Value:  []byte(value),
	}
}

// PropertyBool converts a property path with boolean values to a PropertyEntry.
func PropertyBool(keys []string, value bool) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/bool",
		Value:  []byte(strconv.FormatBool(value)),
	}
}

// PropertyFloat64 converts a property path with float64 values to a PropertyEntry.
func PropertyFloat64(keys []string, value float64) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/float64",
		Value:  Float64ToBinary(value),
	}
}

// Float64ToBinary marshals a float64 to its binary representation.
func Float64ToBinary(v float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(v))
	return buf[:]
}

// BinaryToFloat64 un-marshals a float64 from its binary representation.
func BinaryToFloat64(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}
