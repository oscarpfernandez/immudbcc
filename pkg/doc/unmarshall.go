package doc

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func PropertyListToRaw(properties PropertyEntryList) interface{} {
	sort.Sort(properties)

	var rawObject interface{}

	for _, property := range properties {
		_, keys, vType := property.DissectKeyURI()
		value := property.Value

		if hasArrayFormat(keys[0]) {
			index, capacity := splitArrayFormat(keys[0])
			// Arrays case.
			if rawObject == nil {
				rawObject = make([]interface{}, capacity)
			}
			propertyListToRawArrays(index, rawObject, 0, keys, vType, value)
		} else {
			// Map case.
			if rawObject == nil {
				rawObject = map[string]interface{}{}
			}
			propertyListToRawMap(rawObject, 0, keys, vType, value)
		}
	}

	return rawObject
}

func propertyListToRawMap(parentObject interface{}, curKeyIndex int, keys []string, valueType string, value []byte) {
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
				object[keys[curKeyIndex]] = BinaryToFloat64(value)
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
			propertyListToRawMap(object[keys[curKeyIndex]], curKeyIndex+1, keys, valueType, value)
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
				object[curArrayIndex] = BinaryToFloat64(value)
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
			propertyListToRawMap(object[curArrayIndex], curKeyIndex+1, keys, valueType, value)
		}
	}
}

func hasArrayFormat(s string) bool {
	// Checks for Arrays definitions of the format "[%d.%d]"
	return regexp.MustCompile(`^\[\d+\.\d+]$`).MatchString(s)
}

func splitArrayFormat(s string) (index int, capacity int) {
	if !hasArrayFormat(s) {
		panic("not array format")
	}

	indexCapStr := strings.Trim(s, "[]")
	valuesStr := strings.Split(indexCapStr, ".")
	index, _ = strconv.Atoi(valuesStr[0])
	capacity, _ = strconv.Atoi(valuesStr[1])

	return index, capacity
}
