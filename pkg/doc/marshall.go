package doc

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func GeneratePropertyList(docID string, r io.Reader) (PropertyEntryList, error) {
	var docMap interface{}
	if err := json.NewDecoder(r).Decode(&docMap); err != nil {
		return nil, fmt.Errorf("unable to unmarshall payload: %v", err)
	}

	return rawToPropertyList([]string{docID}, docMap), nil
}

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
