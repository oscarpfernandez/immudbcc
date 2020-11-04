package doc

import (
	"sort"
)

func PropertyListToRaw(properties PropertyEntryList) interface{} {
	sort.Sort(properties)

	var rawObject interface{}

	for idx, property := range properties {
		_, keys, vType := property.DissectKeyURI()
		value := property.Value

		if hasArrayFormat(keys[0]) {
			// Arrays case.
			if idx == 0 {
				rawObject = []interface{}{}
			}
			propertyListToRaw(rawObject, 0, keys, vType, value)
		} else {
			// Map case.
			if idx == 0 {
				rawObject = map[string]interface{}{}
			}
			propertyListToRaw(rawObject, 0, keys, vType, value)
		}
	}

	return rawObject
}

func propertyListToRaw(parentObject interface{}, curKeyIndex int, keys []string, valueType string, value []byte) {
	// Leaf object
	if len(keys) == curKeyIndex+1 {
		switch object := parentObject.(type) {
		// Leaf object is a map.
		case map[string]interface{}:
			switch valueType {
			case "nil":
				object[keys[curKeyIndex]] = nil
			case "string":
				object[keys[curKeyIndex]] = string(value)
			case "bool":
				object[keys[curKeyIndex]] = string(value) == "true"
			case "float64":
				object[keys[curKeyIndex]] = binaryToFloat64(value)
			}
		}

		// backtrack
		return
	}

	switch object := parentObject.(type) {
	// Intermediate node object is a map.
	case map[string]interface{}:
		if hasArrayFormat(keys[curKeyIndex+1]) {
			index, capacity := splitArrayFormat(keys[curKeyIndex+1])
			// Arrays case.
			if object[keys[curKeyIndex]] == nil {
				object[keys[curKeyIndex]] = make([]interface{}, capacity)
			}
			propertyListToRawArrays(index, object[keys[curKeyIndex]], curKeyIndex+1, keys, valueType, value)
		} else {
			if object[keys[curKeyIndex]] == nil {
				object[keys[curKeyIndex]] = map[string]interface{}{}
			}
			propertyListToRaw(object[keys[curKeyIndex]], curKeyIndex+1, keys, valueType, value)
		}
	}
}

func propertyListToRawArrays(curArrayIndex int, parentArray interface{}, curKeyIndex int, keys []string, valueType string, value []byte) {
	// Leaf object
	if len(keys) == curKeyIndex+1 {
		switch object := parentArray.(type) {
		case []interface{}:
			switch valueType {
			case "nil":
				object[curArrayIndex] = nil
			case "string":
				object[curArrayIndex] = string(value)
			case "bool":
				object[curArrayIndex] = string(value) == "true"
			case "float64":
				object[curArrayIndex] = binaryToFloat64(value)
			}
		}
		// backtrack
		return
	}

	switch object := parentArray.(type) {
	// Intermediate node object is a map.
	case []interface{}:
		if hasArrayFormat(keys[curKeyIndex+1]) {
			index, capacity := splitArrayFormat(keys[curKeyIndex+1])
			// Arrays case.
			if object[curArrayIndex] == nil {
				object[curArrayIndex] = make([]interface{}, capacity)
			}
			propertyListToRawArrays(index, object[curArrayIndex], curKeyIndex+1, keys, valueType, value)
		} else {
			if object[curArrayIndex] == nil {
				object[curArrayIndex] = map[string]interface{}{}
			}
			propertyListToRaw(object[curArrayIndex], curKeyIndex+1, keys, valueType, value)
		}
	}
}
