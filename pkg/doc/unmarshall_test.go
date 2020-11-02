package doc

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromPropertyList(t *testing.T) {
	tests := map[string]struct {
		prefix         string
		propertyList   PropertyEntryList
		expJSONPayload []byte
	}{
		"Transforms flat structure": {
			prefix: "prefix1",
			propertyList: PropertyEntryList{
				{KeyURI: "prefix1/employee/name/string", Value: []byte("John")},
				{KeyURI: "prefix1/employee/age/float64", Value: float64ToBinary(30)},
				{KeyURI: "prefix1/employee/index/float64", Value: float64ToBinary(18446744073709551615)},
				{KeyURI: "prefix1/employee/city/string", Value: []byte("New York")},
				{KeyURI: "prefix1/employee/active/bool", Value: []byte(strconv.FormatBool(true))},
			},
			expJSONPayload: []byte(`{
            "employee":{
               "index": 18446744073709551615,
               "name": "John",
               "age": 30,
               "city": "New York",
               "active": true
            }
         }`),
		},
		//"Transform nested object array": {
		// prefix: "prefix1",
		// expJSONPayload: []byte(`{
		//    "top-tags": ["tag1","tag2","tag3","tag4","tag5","tag6"],
		//    "nested": {
		//       "nested-tags": ["tag7","tag8","tag9","tag10","tag11","tag12"],
		//       "name": "tagger"
		//    }
		// }`),
		// propertyList: PropertyEntryList{
		//    {KeyURI: "prefix1/top-tags/[0.6]/string", Value: []byte("tag1")},
		//    {KeyURI: "prefix1/top-tags/[1.6]/string", Value: []byte("tag2")},
		//    {KeyURI: "prefix1/top-tags/[2.6]/string", Value: []byte("tag3")},
		//    {KeyURI: "prefix1/top-tags/[3.6]/string", Value: []byte("tag4")},
		//    {KeyURI: "prefix1/top-tags/[4.6]/string", Value: []byte("tag5")},
		//    {KeyURI: "prefix1/top-tags/[5.6]/string", Value: []byte("tag6")},
		//    {KeyURI: "prefix1/nested/name/string", Value: []byte("tagger")},
		//    {KeyURI: "prefix1/nested/nested-tags/[0.6]/string", Value: []byte("tag7")},
		//    {KeyURI: "prefix1/nested/nested-tags/[1.6]/string", Value: []byte("tag8")},
		//    {KeyURI: "prefix1/nested/nested-tags/[2.6]/string", Value: []byte("tag9")},
		//    {KeyURI: "prefix1/nested/nested-tags/[3.6]/string", Value: []byte("tag10")},
		//    {KeyURI: "prefix1/nested/nested-tags/[4.6]/string", Value: []byte("tag11")},
		//    {KeyURI: "prefix1/nested/nested-tags/[5.6]/string", Value: []byte("tag12")},
		// },
		//},
		"Transform object array": {
			prefix: "prefix1",
			expJSONPayload: []byte(`{
            "tags": ["tag1","tag2","tag3","tag4","tag5","tag6"]
         }`),
			propertyList: PropertyEntryList{
				{KeyURI: "prefix1/tags/[0.5]/string", Value: []byte("tag1")},
				{KeyURI: "prefix1/tags/[1.5]/string", Value: []byte("tag2")},
				{KeyURI: "prefix1/tags/[2.5]/string", Value: []byte("tag3")},
				{KeyURI: "prefix1/tags/[3.5]/string", Value: []byte("tag4")},
				{KeyURI: "prefix1/tags/[4.5]/string", Value: []byte("tag5")},
				{KeyURI: "prefix1/tags/[5]/string", Value: []byte("tag6")},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var object interface{}
			if err := json.NewDecoder(bytes.NewReader(test.expJSONPayload)).Decode(&object); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			rawMap := PropertyListToRaw(test.propertyList)

			gotPayload, err := json.Marshal(rawMap)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.JSONEq(t, string(test.expJSONPayload), string(gotPayload), "list should match")
		})
	}
}
