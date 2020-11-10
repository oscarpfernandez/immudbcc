package doc

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// RawToPropertyList creates the property list for given document provided
// a reader to the raw payload.
func RawToPropertyList(docID string, r io.Reader) (PropertyEntryList, error) {
	var docMap interface{}
	if err := json.NewDecoder(r).Decode(&docMap); err != nil {
		return nil, fmt.Errorf("unable to unmarshall payload: %v", err)
	}

	return rawToPropertyList([]string{docID}, docMap), nil
}

// rawToPropertyList recursively transverses the raw object tree, building a
// list of property entries for every leaf of said tree. Each property contains
// Key-Value pair. The Key, describes a path from the root to the leaf, and the
// Value is the leaf's value.
func rawToPropertyList(keys []string, value interface{}) PropertyEntryList {
	list := PropertyEntryList{}

	// https://www.w3schools.com/js/js_json_datatypes.asp
	switch v := value.(type) {
	case nil:
		list = append(list, PropertyNil(keys))
	case string:
		list = append(list, PropertyString(keys, v))
	case bool:
		list = append(list, PropertyBool(keys, v))
	case float64:
		list = append(list, PropertyFloat64(keys, v))
	case map[string]interface{}:
		for key, value := range v {
			keys = append(keys, key)
			list = append(list, rawToPropertyList(keys, value)...)
			removeLastElement(&keys)
		}
	case []interface{}:
		vLen := len(v)
		for idx, arrElem := range v {
			keys = append(keys, "["+strconv.Itoa(idx)+"."+strconv.Itoa(vLen)+"]")
			list = append(list, rawToPropertyList(keys, arrElem)...)
			removeLastElement(&keys)
		}
	}

	return list
}

// removeLastElement removes the last object of a non-empty slice of strings.
func removeLastElement(s *[]string) {
	if s == nil || len(*s) == 0 {
		return
	}
	*s = (*s)[:len(*s)-1]
}
