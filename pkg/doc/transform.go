package doc

import (
	"encoding/binary"
	"math"
	"sort"
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
		for idx, arrElem := range v {
			keys = append(keys, "["+strconv.Itoa(idx)+"]")
			subList := RawToPropertyList(keys, arrElem)
			list = append(list, subList...)
			removeLastElement(&keys)
		}
	}

	return list
}

func PropertyListToRaw(properties PropertyEntryList) interface{} {
	sort.Sort(properties)

	var rawObject interface{}

	for _, property := range properties {
		keys, vType := property.DissectKeyURI()
		value := property.Value

		if strings.HasPrefix(keys[0], "[") && strings.HasSuffix(keys[0], "]") {
			// Arrays case.
			rawObject = []interface{}{}
			propertyListToRaw(rawObject, keys[1:], vType, value)
		} else {
			// Map case.
			rawObject = map[string]interface{}{}
			propertyListToRaw(rawObject, keys[1:], vType, value)
		}
	}

	return rawObject
}

func propertyListToRaw(parentObject interface{}, keys []string, valueType string, value []byte) {
	// Leaf object
	if len(keys) == 1 {
		switch object := parentObject.(type) {
		// Leaf object is a map.
		case map[string]interface{}:
			switch valueType {
			case "nil":
				object[keys[0]] = nil
			case "string":
				object[keys[0]] = string(value)
			case "bool":
				object[keys[0]] = string(value) == "true"
			case "float64":
				object[keys[0]] = binaryToFloat64(value)
			}

		// Leaf object is a map.
		case []interface{}:
			switch valueType {
			case "nil":
				object = append(object, nil)
			case "string":
				object = append(object, string(value))
			case "bool":
				object = append(object, string(value) == "true")
			case "float64":
				object = append(object, binaryToFloat64(value))
			}
		}

		// backtrack
		return
	}

	switch object := parentObject.(type) {
	// Intermediate node object is a map.
	case map[string]interface{}:
		if strings.HasPrefix(keys[0], "[") && strings.HasSuffix(keys[0], "]") {
			// Arrays case.
			if object[keys[0]] == nil {
				object[keys[0]] = []interface{}{}
			}
			propertyListToRaw(object[keys[0]], keys[1:], valueType, value)
		} else {
			// Map case.
			if object[keys[0]] == nil {
				object[keys[0]] = map[string]interface{}{}
			}
			propertyListToRaw(object[keys[0]], keys[1:], valueType, value)
		}
	case []interface{}:
		// If we are in an array the only elements admissible are objects or
		// basic types, given that arrays of arrays are not possible in JSON.
		newMap := map[string]interface{}{}
		object = append(object, newMap)
		propertyListToRaw(newMap, keys[1:], valueType, value)
	}
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
