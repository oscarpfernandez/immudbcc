package doc

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// PropertyListToRaw converts a list of PropertyEntry to the raw original
// document object.
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

// propertyListToRawMap recursively analyzes a PropertyEntry's Key, building
// the equivalent structure in the raw document object. This cases deals with
// the case where the current root element in the path being analyzed consists
// of a Map.
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

// propertyListToRawArrays recursively analyzes a PropertyEntry's Key, building
// the equivalent structure in the raw document object. This cases deals with
// the case where the current root element in the path being analyzed consists
// of an Array.
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

var arrayRegExp = regexp.MustCompile(`^\[\d+\.\d+]$`)

// hasArrayFormat checks if the current node of the path describes and array element.
func hasArrayFormat(s string) bool {
	// Checks for Arrays definitions of the format "[%d.%d]" where the first
	// parameter describes the current index and the second the total capacity.
	return arrayRegExp.MatchString(s)
}

// splitArrayFormat given a current array element, returns the associates index
// and total capacity.
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
