package doc

import (
	"strconv"
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
			keys = append(keys, keys[len(keys)-1]+"["+strconv.Itoa(idx)+"."+strconv.Itoa(vLen)+"]")
			list = append(list, RawToPropertyList(keys, arrElem)...)
			removeLastElement(&keys)
		}
	}

	return list
}
