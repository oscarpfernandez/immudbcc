package doc

import (
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
		"Transform nested object array": {
			prefix: "prefix1",
			expJSONPayload: []byte(`{
				"top-tags": ["tag1","tag2","tag3","tag4","tag5","tag6"],
					"nested": {
						"nested-tags": ["tag7","tag8","tag9","tag10","tag11","tag12"],
						"name": "tagger"
					}
				}`),
			propertyList: PropertyEntryList{
				{KeyURI: "prefix1/top-tags/[0.6]/string", Value: []byte("tag1")},
				{KeyURI: "prefix1/top-tags/[1.6]/string", Value: []byte("tag2")},
				{KeyURI: "prefix1/top-tags/[2.6]/string", Value: []byte("tag3")},
				{KeyURI: "prefix1/top-tags/[3.6]/string", Value: []byte("tag4")},
				{KeyURI: "prefix1/top-tags/[4.6]/string", Value: []byte("tag5")},
				{KeyURI: "prefix1/top-tags/[5.6]/string", Value: []byte("tag6")},
				{KeyURI: "prefix1/nested/name/string", Value: []byte("tagger")},
				{KeyURI: "prefix1/nested/nested-tags/[0.6]/string", Value: []byte("tag7")},
				{KeyURI: "prefix1/nested/nested-tags/[1.6]/string", Value: []byte("tag8")},
				{KeyURI: "prefix1/nested/nested-tags/[2.6]/string", Value: []byte("tag9")},
				{KeyURI: "prefix1/nested/nested-tags/[3.6]/string", Value: []byte("tag10")},
				{KeyURI: "prefix1/nested/nested-tags/[4.6]/string", Value: []byte("tag11")},
				{KeyURI: "prefix1/nested/nested-tags/[5.6]/string", Value: []byte("tag12")},
			},
		},
		"Transform simple object array #1": {
			prefix: "prefix1",
			expJSONPayload: []byte(`{
				"tags": ["tag1","tag2","tag3","tag4","tag5","tag6"]
			}`),
			propertyList: PropertyEntryList{
				{KeyURI: "prefix1/tags/[0.6]/string", Value: []byte("tag1")},
				{KeyURI: "prefix1/tags/[1.6]/string", Value: []byte("tag2")},
				{KeyURI: "prefix1/tags/[2.6]/string", Value: []byte("tag3")},
				{KeyURI: "prefix1/tags/[3.6]/string", Value: []byte("tag4")},
				{KeyURI: "prefix1/tags/[4.6]/string", Value: []byte("tag5")},
				{KeyURI: "prefix1/tags/[5.6]/string", Value: []byte("tag6")},
			},
		},
		"Transforms simple object array #2": {
			prefix: "objectID",
			expJSONPayload: []byte(`{
				"people": [
					{"id": 0,"name": "Monroe Roth"},
					{"id": 1,"name": "Mullen Rhodes"},
					{"id": 2,"name": "Mcclure Welch"}
				]
			}`),
			propertyList: PropertyEntryList{
				{KeyURI: "objectID/people/[0.3]/id/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/people/[0.3]/name/string", Value: []byte("Monroe Roth")},
				{KeyURI: "objectID/people/[1.3]/id/float64", Value: float64ToBinary(1)},
				{KeyURI: "objectID/people/[1.3]/name/string", Value: []byte("Mullen Rhodes")},
				{KeyURI: "objectID/people/[2.3]/id/float64", Value: float64ToBinary(2)},
				{KeyURI: "objectID/people/[2.3]/name/string", Value: []byte("Mcclure Welch")},
			},
		},
		"Transform object with chained arrays": {
			prefix: "objectID",
			propertyList: PropertyEntryList{
				{KeyURI: "objectID/type/string", Value: []byte("FeatureCollection")},
				{KeyURI: "objectID/features/[0.2]/type/string", Value: []byte("Feature")},
				{KeyURI: "objectID/features/[0.2]/properties/MAPBLKLOT/string", Value: []byte("0001001")},
				{KeyURI: "objectID/features/[0.2]/properties/BLKLOT/string", Value: []byte("0001001")},
				{KeyURI: "objectID/features/[0.2]/properties/FROM_ST/string", Value: []byte("0")},
				{KeyURI: "objectID/features/[0.2]/properties/STREET/string", Value: []byte("UNKNOWN")},
				{KeyURI: "objectID/features/[0.2]/properties/ODD_EVEN/string", Value: []byte("E")},
				{KeyURI: "objectID/features/[0.2]/properties/BLOCK_NUM/string", Value: []byte("0001")},
				{KeyURI: "objectID/features/[0.2]/properties/LOT_NUM/string", Value: []byte("001")},
				{KeyURI: "objectID/features/[0.2]/properties/TO_ST/string", Value: []byte("0")},
				{KeyURI: "objectID/features/[0.2]/properties/ST_TYPE/nil", Value: nil},
				{KeyURI: "objectID/features/[0.2]/geometry/type/string", Value: []byte("Polygon")},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[0.5]/[0.3]/float64", Value: float64ToBinary(-122.42200352825247)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[0.5]/[1.3]/float64", Value: float64ToBinary(37.80848009696725)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[0.5]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[1.5]/[0.3]/float64", Value: float64ToBinary(-122.42207601332528)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[1.5]/[1.3]/float64", Value: float64ToBinary(37.808835019815085)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[1.5]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[2.5]/[0.3]/float64", Value: float64ToBinary(-122.42110217434863)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[2.5]/[1.3]/float64", Value: float64ToBinary(37.808803534992904)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[2.5]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[3.5]/[0.3]/float64", Value: float64ToBinary(-122.42106256906727)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[3.5]/[1.3]/float64", Value: float64ToBinary(37.80860105681815)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[3.5]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[4.5]/[0.3]/float64", Value: float64ToBinary(-122.42200352825247)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[4.5]/[1.3]/float64", Value: float64ToBinary(37.80848009696725)},
				{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[4.5]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[1.2]/type/string", Value: []byte("Feature")},
				{KeyURI: "objectID/features/[1.2]/properties/STREET/string", Value: []byte("UNKNOWN")},
				{KeyURI: "objectID/features/[1.2]/properties/ODD_EVEN/string", Value: []byte("E")},
				{KeyURI: "objectID/features/[1.2]/properties/MAPBLKLOT/string", Value: []byte("0004002")},
				{KeyURI: "objectID/features/[1.2]/properties/BLOCK_NUM/string", Value: []byte("0004")},
				{KeyURI: "objectID/features/[1.2]/properties/LOT_NUM/string", Value: []byte("002")},
				{KeyURI: "objectID/features/[1.2]/properties/FROM_ST/string", Value: []byte("0")},
				{KeyURI: "objectID/features/[1.2]/properties/BLKLOT/string", Value: []byte("0004002")},
				{KeyURI: "objectID/features/[1.2]/properties/TO_ST/string", Value: []byte("0")},
				{KeyURI: "objectID/features/[1.2]/properties/ST_TYPE/nil", Value: nil},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[0.4]/[0.3]/float64", Value: float64ToBinary(-122.41570120460688)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[0.4]/[1.3]/float64", Value: float64ToBinary(37.80832725267146)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[0.4]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[1.4]/[0.3]/float64", Value: float64ToBinary(-122.4157607435932)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[1.4]/[1.3]/float64", Value: float64ToBinary(37.808630700240904)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[1.4]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[2.4]/[0.3]/float64", Value: float64ToBinary(-122.4137878913324)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[2.4]/[1.3]/float64", Value: float64ToBinary(37.80856680131984)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[2.4]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[3.4]/[0.3]/float64", Value: float64ToBinary(-122.41570120460688)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[3.4]/[1.3]/float64", Value: float64ToBinary(37.80832725267146)},
				{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[3.4]/[2.3]/float64", Value: float64ToBinary(0)},
				{KeyURI: "objectID/features/[1.2]/geometry/type/string", Value: []byte("Polygon")},
			},
			expJSONPayload: []byte(`{
				"type": "FeatureCollection",
				"features": [
					{
						"type": "Feature",
						"properties": {
							"MAPBLKLOT": "0001001",
							"BLKLOT": "0001001",
							"BLOCK_NUM": "0001",
							"LOT_NUM": "001",
							"FROM_ST": "0",
							"TO_ST": "0",
							"STREET": "UNKNOWN",
							"ST_TYPE": null,
							"ODD_EVEN": "E"
						},
						"geometry": {
						"type": "Polygon",
						"coordinates": [
							[
								[-122.42200352825247,37.80848009696725,0],
								[-122.42207601332528,37.808835019815085,0],
								[-122.42110217434863,37.808803534992904,0],
								[-122.42106256906727,37.80860105681815,0],
								[-122.42200352825247,37.80848009696725,0]
							]
						]
					}
				},
				{
					"type": "Feature",
					"properties": {
						"MAPBLKLOT": "0004002",
						"BLKLOT": "0004002",
						"BLOCK_NUM": "0004",
						"LOT_NUM": "002",
						"FROM_ST": "0",
						"TO_ST": "0",
						"STREET": "UNKNOWN",
						"ST_TYPE": null,
						"ODD_EVEN": "E"
					},
					"geometry": {
					"type": "Polygon",
					"coordinates": [
						[
							[-122.41570120460688,37.80832725267146,0],
							[-122.4157607435932,37.808630700240904,0],
							[-122.4137878913324,37.80856680131984,0],
							[-122.41570120460688,37.80832725267146,0]
						]
					]
				}
			}
		]}`),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rawMap := PropertyListToRaw(test.propertyList)

			gotPayload, err := json.Marshal(rawMap)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.JSONEq(t, string(test.expJSONPayload), string(gotPayload), "list should match")
		})
	}
}
