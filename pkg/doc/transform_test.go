package doc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePropertyList(t *testing.T) {
	tests := map[string]struct {
		prefix      string
		jsonPayload []byte
		expList     PropertyEntryList
	}{
		"Transforms flat JSON": {
			prefix: "prefix1",
			jsonPayload: []byte(`{
				"employee":{ 
					"name":"John", 
					"age":30, 
					"city":"New York" 
				}
			}`),
			expList: PropertyEntryList{
				{KeyURI: "prefix1/employee/name/string", Value: []byte("John")},
				{KeyURI: "prefix1/employee/age/float64", Value: float64ToBinary(30)},
				{KeyURI: "prefix1/employee/city/string", Value: []byte("New York")},
			},
		},
		"Transforms 2-nested JSON objects": {
			prefix: "prefix2",
			jsonPayload: []byte(`{
				"name":"John",
  				"age":30,
				"cars": {
					"car1":"Ford",
					"car2":"BMW",
					"car3":"Fiat"
				}
			}`),
			expList: PropertyEntryList{
				{KeyURI: "prefix2/cars/car1/string", Value: []byte(`Ford`)},
				{KeyURI: "prefix2/cars/car2/string", Value: []byte(`BMW`)},
				{KeyURI: "prefix2/cars/car3/string", Value: []byte(`Fiat`)},
				{KeyURI: "prefix2/name/string", Value: []byte(`John`)},
				{KeyURI: "prefix2/age/float64", Value: float64ToBinary(30)},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var rawMap map[string]interface{}
			if err := json.NewDecoder(bytes.NewReader(test.jsonPayload)).Decode(&rawMap); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotList := CreatePropertyList([]string{test.prefix}, rawMap)
			assert.ElementsMatch(t, gotList, test.expList, "list should match")
		})
	}
}
