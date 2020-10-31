package doc

import (
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

func CreatePropertyList(keys []string, rawMap map[string]interface{}) PropertyEntryList {
	list := PropertyEntryList{}

	// https://www.w3schools.com/js/js_json_datatypes.asp
	for key, value := range rawMap {
		// Build tree's traversal path.
		keys = append(keys, key)

		switch v := value.(type) {
		case nil:
			list = append(list, PropertyNil(keys))
		case string:
			list = append(list, PropertyString(keys, v))
		case int64:
			list = append(list, PropertyInt64(keys, v))
		case uint64:
			list = append(list, PropertyUInt64(keys, v))
		case bool:
			list = append(list, PropertyBool(keys, v))
		case float32:
			list = append(list, PropertyFloat32(keys, v))
		case float64:
			list = append(list, PropertyFloat64(keys, v))
		case map[string]interface{}:
			subList := CreatePropertyList(keys, v)
			list = append(list, subList...)
		case []map[string]interface{}:

		case []interface{}:

		}

		// Drop last element visited.
		RemoveLastElement(&keys)
	}

	return list
}

func PropertyNil(keys []string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/nil",
		Value:  nil,
	}
}

func PropertyString(keys []string, value string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/string",
		Value:  []byte(value),
	}
}

func PropertyBool(keys []string, value bool) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/bool",
		Value:  []byte(strconv.FormatBool(value)),
	}
}

func PropertyInt64(key []string, value int64) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(key, "/") + "/int64",
		Value:  int64ToBinary(value),
	}
}

func PropertyUInt64(key []string, value uint64) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(key, "/") + "/uint64",
		Value:  uint64ToBinary(value),
	}
}

func PropertyFloat64(key []string, value float64) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(key, "/") + "/float64",
		Value:  float64ToBinary(value),
	}
}

func PropertyFloat32(key []string, value float32) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(key, "/") + "/float32",
		Value:  float32ToBinary(value),
	}
}

func RemoveLastElement(s *[]string) {
	if s == nil || len(*s) == 0 {
		return
	}
	*s = (*s)[:len(*s)-1]
}

func int64ToBinary(v int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, v)
	return buf[:n]
}

func uint64ToBinary(v uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, v)
	return buf[:n]
}

func float32ToBinary(v float32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(v))
	return buf[:]
}

func float64ToBinary(v float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(v))
	return buf[:]
}
