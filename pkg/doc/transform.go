package doc

import (
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

func RawToPropertyList(keys []string, value interface{}) PropertyEntryList {
	list := PropertyEntryList{}

	// https://www.w3schools.com/js/js_json_datatypes.asp
	switch v := value.(type) {
	case nil:
		list = append(list, propertyNil(keys))

	case string:
		list = append(list, propertyString(keys, v))

	case bool:
		list = append(list, propertyBool(keys, v))

	case float64:
		list = append(list, propertyFloat64(keys, v))

	case map[string]interface{}:
		for key, value := range v {
			keys = append(keys, key)
			list = append(list, RawToPropertyList(keys, value)...)
			removeLastElement(&keys)
		}

	case []interface{}:
		vLen := len(v)
		for idx, arrElem := range v {
			keys = append(keys, "["+strconv.Itoa(idx)+"."+strconv.Itoa(vLen)+"]")
			subList := RawToPropertyList(keys, arrElem)
			list = append(list, subList...)
			removeLastElement(&keys)
		}
	}

	return list
}

func propertyNil(keys []string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/nil",
		Value:  nil,
	}
}

func propertyString(keys []string, value string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/string",
		Value:  []byte(value),
	}
}

func propertyBool(keys []string, value bool) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/bool",
		Value:  []byte(strconv.FormatBool(value)),
	}
}

func propertyFloat64(keys []string, value float64) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/float64",
		Value:  float64ToBinary(value),
	}
}

func removeLastElement(s *[]string) {
	if s == nil || len(*s) == 0 {
		return
	}
	*s = (*s)[:len(*s)-1]
}

func float64ToBinary(v float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(v))
	return buf[:]
}

func binaryToFloat64(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}
