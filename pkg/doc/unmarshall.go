package doc

import (
	"sort"
	"strconv"
	"strings"
)

func PropertyListToRaw(properties PropertyEntryList) interface{} {
	sort.Sort(properties)

	var rawObject interface{}

	for idx, property := range properties {
		keys, vType := property.DissectKeyURI()
		value := property.Value
		//fmt.Printf("property: %s\n", property)

		if strings.HasPrefix(keys[0], "[") && strings.HasSuffix(keys[0], "]") {
			// Arrays case.
			if idx == 0 {
				rawObject = []interface{}{}
			}
			propertyListToRaw(rawObject, keys[1:], vType, value)
		} else {
			// Map case.
			if idx == 0 {
				rawObject = map[string]interface{}{}
			}
			propertyListToRaw(rawObject, keys[1:], vType, value)
		}
	}

	return rawObject
}

func propertyListToRaw(parentObject interface{}, keys []string, valueType string, value []byte) {
	//fmt.Printf("keys: %s\n", keys)
	//fmt.Printf("valueType :%s\n", valueType)

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
			index, _ := strconv.Atoi(strings.Trim(keys[0], "[]"))
			// Arrays case.
			if object[keys[0]] == nil {
				object[keys[0]] = []interface{}{}
			}
			childSlice := object[keys[0]].([]interface{})
			if childSlice == nil {
				childSlice = make([]interface{}, index+1)
			} else {
				newSlice := make([]interface{}, index+1)
				copy(newSlice, childSlice)
				childSlice = newSlice
			}
			propertyListToRaw(childSlice, keys[1:], valueType, value)
		} else {
			var childMap map[string]interface{}
			if object[keys[0]] == nil {
				childMap = map[string]interface{}{}
				object[keys[0]] = childMap
			} else {
				childMap = object[keys[0]].(map[string]interface{})
			}
			propertyListToRaw(childMap, keys[1:], valueType, value)
		}
	case []interface{}:
		// If we are in an array the only elements admissible are objects or
		// basic types, given that arrays of arrays are not possible in JSON.
		propertyListToRaw(object, keys[1:], valueType, value)
	}
}
