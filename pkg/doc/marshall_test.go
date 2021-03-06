package doc

import (
	"bytes"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCases = map[string]struct {
	prefix       string
	jsonPayload  []byte
	propertyList PropertyEntryList
}{
	"Transforms flat object": {
		prefix: "objectID",
		jsonPayload: []byte(`{
			"index": 18446744073709551615,
			"name": "John",
			"age": 30,
			"city": "New York",
			"active": true,
            "random": null
		}`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/name/string", Value: []byte("John")},
			{KeyURI: "objectID/age/float64", Value: Float64ToBinary(30)},
			{KeyURI: "objectID/index/float64", Value: Float64ToBinary(18446744073709551615)},
			{KeyURI: "objectID/city/string", Value: []byte("New York")},
			{KeyURI: "objectID/active/bool", Value: []byte(strconv.FormatBool(true))},
			{KeyURI: "objectID/random/nil", Value: nil},
		},
	},
	"Transforms nested objects #1": {
		prefix: "objectID",
		jsonPayload: []byte(`{
				"employee":{
					"index": 18446744073709551615,
					"name": "John",
					"age": 30,
					"city": "New York",
					"active": true
				}
			}`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/employee/name/string", Value: []byte("John")},
			{KeyURI: "objectID/employee/age/float64", Value: Float64ToBinary(30)},
			{KeyURI: "objectID/employee/index/float64", Value: Float64ToBinary(18446744073709551615)},
			{KeyURI: "objectID/employee/city/string", Value: []byte("New York")},
			{KeyURI: "objectID/employee/active/bool", Value: []byte(strconv.FormatBool(true))},
		},
	},
	"Transforms nested objects #2": {
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
		propertyList: PropertyEntryList{
			{KeyURI: "prefix2/cars/car1/string", Value: []byte(`Ford`)},
			{KeyURI: "prefix2/cars/car2/string", Value: []byte(`BMW`)},
			{KeyURI: "prefix2/cars/car3/string", Value: []byte(`Fiat`)},
			{KeyURI: "prefix2/name/string", Value: []byte(`John`)},
			{KeyURI: "prefix2/age/float64", Value: Float64ToBinary(30)},
		},
	},
	"Transforms simple object array #1": {
		prefix: "objectID",
		jsonPayload: []byte(`{ 
				"tags": ["tag1","tag2","tag3","tag4","tag5","tag6"],
				"nested": { 
					"tags": ["tag7","tag8","tag9","tag10","tag11","tag12"],
					"name": "tagger"
				}
			}`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/tags/[0.6]/string", Value: []byte("tag1")},
			{KeyURI: "objectID/tags/[1.6]/string", Value: []byte("tag2")},
			{KeyURI: "objectID/tags/[2.6]/string", Value: []byte("tag3")},
			{KeyURI: "objectID/tags/[3.6]/string", Value: []byte("tag4")},
			{KeyURI: "objectID/tags/[4.6]/string", Value: []byte("tag5")},
			{KeyURI: "objectID/tags/[5.6]/string", Value: []byte("tag6")},
			{KeyURI: "objectID/nested/tags/[0.6]/string", Value: []byte("tag7")},
			{KeyURI: "objectID/nested/tags/[1.6]/string", Value: []byte("tag8")},
			{KeyURI: "objectID/nested/tags/[2.6]/string", Value: []byte("tag9")},
			{KeyURI: "objectID/nested/tags/[3.6]/string", Value: []byte("tag10")},
			{KeyURI: "objectID/nested/tags/[4.6]/string", Value: []byte("tag11")},
			{KeyURI: "objectID/nested/tags/[5.6]/string", Value: []byte("tag12")},
			{KeyURI: "objectID/nested/name/string", Value: []byte("tagger")},
		},
	},
	"Transforms simple object array #2": {
		prefix: "objectID",
		jsonPayload: []byte(`{
				"people": [
					{"id": 0,"name": "Monroe Roth"},
					{"id": 1,"name": "Mullen Rhodes"},
					{"id": 2,"name": "Mcclure Welch"}
				]
			}`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/people/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/people/[0.3]/name/string", Value: []byte("Monroe Roth")},
			{KeyURI: "objectID/people/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/people/[1.3]/name/string", Value: []byte("Mullen Rhodes")},
			{KeyURI: "objectID/people/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/people/[2.3]/name/string", Value: []byte("Mcclure Welch")},
		},
	},
	"Transforms simple object array #3": {
		prefix: "objectID",
		jsonPayload: []byte(`{ 
				"tags": ["tag1","tag2","tag3","tag4","tag5","tag6"]
			}`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/tags/[0.6]/string", Value: []byte("tag1")},
			{KeyURI: "objectID/tags/[1.6]/string", Value: []byte("tag2")},
			{KeyURI: "objectID/tags/[2.6]/string", Value: []byte("tag3")},
			{KeyURI: "objectID/tags/[3.6]/string", Value: []byte("tag4")},
			{KeyURI: "objectID/tags/[4.6]/string", Value: []byte("tag5")},
			{KeyURI: "objectID/tags/[5.6]/string", Value: []byte("tag6")},
		},
	},
	"Transforms simple object array #4": {
		prefix: "objectID",
		jsonPayload: []byte(`[
				{ "name":"Ford", "models":[ "Fiesta", "Focus", "Mustang" ] },
				{ "name":"BMW", "models":[ "320", "X3", "X5" ] },
				{ "name":"Fiat", "models":[ "500", "Panda" ] }
			]`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/[0.3]/name/string", Value: []byte("Ford")},
			{KeyURI: "objectID/[0.3]/models/[0.3]/string", Value: []byte("Fiesta")},
			{KeyURI: "objectID/[0.3]/models/[1.3]/string", Value: []byte("Focus")},
			{KeyURI: "objectID/[0.3]/models/[2.3]/string", Value: []byte("Mustang")},
			{KeyURI: "objectID/[1.3]/name/string", Value: []byte("BMW")},
			{KeyURI: "objectID/[1.3]/models/[0.3]/string", Value: []byte("320")},
			{KeyURI: "objectID/[1.3]/models/[1.3]/string", Value: []byte("X3")},
			{KeyURI: "objectID/[1.3]/models/[2.3]/string", Value: []byte("X5")},
			{KeyURI: "objectID/[2.3]/name/string", Value: []byte("Fiat")},
			{KeyURI: "objectID/[2.3]/models/[0.2]/string", Value: []byte("500")},
			{KeyURI: "objectID/[2.3]/models/[1.2]/string", Value: []byte("Panda")},
		},
	},
	"Transforms complex object #1": {
		prefix: "objectID",
		jsonPayload: []byte(`{
				"id": "0001",
				"type": "donut",
				"name": "Cake",
				"ppu": 0.55,
				"price": null,
				"batters": {
					"batter": [
						{ "id": "1001", "type": "Regular" },
						{ "id": "1002", "type": "Chocolate" },
						{ "id": "1003", "type": "Blueberry" },
						{ "id": "1004", "type": "Devil's Food" }
					]
				},
				"topping":
					[
						{ "id": "5001", "type": "None" },
						{ "id": "5002", "type": "Glazed" },
						{ "id": "5005", "type": "Sugar" },
						{ "id": "5007", "type": "Powdered Sugar" },
						{ "id": "5006", "type": "Chocolate with Sprinkles" },
						{ "id": "5003", "type": "Chocolate" },
						{ "id": "5004", "type": "Maple" }
					]
			}`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/id/string", Value: []byte("0001")},
			{KeyURI: "objectID/type/string", Value: []byte("donut")},
			{KeyURI: "objectID/name/string", Value: []byte("Cake")},
			{KeyURI: "objectID/ppu/float64", Value: Float64ToBinary(0.55)},
			{KeyURI: "objectID/price/nil", Value: nil},
			{KeyURI: "objectID/batters/batter/[0.4]/id/string", Value: []byte("1001")},
			{KeyURI: "objectID/batters/batter/[0.4]/type/string", Value: []byte("Regular")},
			{KeyURI: "objectID/batters/batter/[1.4]/id/string", Value: []byte("1002")},
			{KeyURI: "objectID/batters/batter/[1.4]/type/string", Value: []byte("Chocolate")},
			{KeyURI: "objectID/batters/batter/[2.4]/id/string", Value: []byte("1003")},
			{KeyURI: "objectID/batters/batter/[2.4]/type/string", Value: []byte("Blueberry")},
			{KeyURI: "objectID/batters/batter/[3.4]/id/string", Value: []byte("1004")},
			{KeyURI: "objectID/batters/batter/[3.4]/type/string", Value: []byte("Devil's Food")},
			{KeyURI: "objectID/topping/[0.7]/type/string", Value: []byte("None")},
			{KeyURI: "objectID/topping/[0.7]/id/string", Value: []byte("5001")},
			{KeyURI: "objectID/topping/[1.7]/id/string", Value: []byte("5002")},
			{KeyURI: "objectID/topping/[1.7]/type/string", Value: []byte("Glazed")},
			{KeyURI: "objectID/topping/[2.7]/id/string", Value: []byte("5005")},
			{KeyURI: "objectID/topping/[2.7]/type/string", Value: []byte("Sugar")},
			{KeyURI: "objectID/topping/[3.7]/id/string", Value: []byte("5007")},
			{KeyURI: "objectID/topping/[3.7]/type/string", Value: []byte("Powdered Sugar")},
			{KeyURI: "objectID/topping/[4.7]/id/string", Value: []byte("5006")},
			{KeyURI: "objectID/topping/[4.7]/type/string", Value: []byte("Chocolate with Sprinkles")},
			{KeyURI: "objectID/topping/[5.7]/id/string", Value: []byte("5003")},
			{KeyURI: "objectID/topping/[5.7]/type/string", Value: []byte("Chocolate")},
			{KeyURI: "objectID/topping/[6.7]/id/string", Value: []byte("5004")},
			{KeyURI: "objectID/topping/[6.7]/type/string", Value: []byte("Maple")},
		},
	},
	"Transforms complex object #2": {
		prefix: "objectID",
		jsonPayload: []byte(`[
			  {
				"_id": "5f9deaa12b81ec174f75e315",
				"index": 0,
				"guid": "5e3ba5a0-7d5f-4b70-bf41-7c882b4da1ef",
				"isActive": true,
				"balance": "$3,166.17",
				"picture": "http://placehold.it/32x32",
				"age": 29,
				"eyeColor": "blue",
				"name": {
				  "first": "Rosario",
				  "last": "Camacho"
				},
				"company": "TERAPRENE",
				"email": "rosario.camacho@teraprene.co.uk",
				"phone": "+1 (940) 577-3244",
				"address": "766 Keap Street, Smock, California, 1121",
				"about": "Pariatur sint do pariatur dolor eiusmod reprehenderit non ex minim ullamco quis consequat.",
				"registered": "Friday, November 7, 2014 7:18 PM",
				"latitude": "-16.92488",
				"longitude": "151.382606",
				"tags": ["non","anim","esse","nostrud","veniam"],
				"range": [0,1,2,3,4,5,6,7,8,9],
				"friends": [
				  {"id": 0,"name": "Clay Nash"},
				  {"id": 1,"name": "Amanda Warner"},
				  {"id": 2,"name": "Kirsten Whitehead"}
				],
				"greeting": "Hello, Rosario! You have 7 unread messages.",
				"favoriteFruit": "apple"
			  },
			  {
				"_id": "5f9deaa1feed31fcbcc6e53e",
				"index": 1,
				"guid": "b9de6d8c-365e-4ca4-8ddb-218e1ce6f112",
				"isActive": true,
				"balance": "$1,739.73",
				"picture": "http://placehold.it/32x32",
				"age": 31,
				"eyeColor": "green",
				"name": {
				  "first": "Randall",
				  "last": "Conley"
				},
				"company": "BESTO",
				"email": "randall.conley@besto.me",
				"phone": "+1 (891) 595-2961",
				"address": "682 Meserole Avenue, Sexton, Georgia, 7532",
				"about": "Exercitation cillum sint dolore aute ex anim deserunt veniam excepteur.",
				"registered": "Wednesday, April 1, 2015 7:06 AM",
				"latitude": "-61.200296",
				"longitude": "81.917956",
				"tags": ["ut","exercitation","nostrud","consequat","esse"],
				"range": [0,1,2,3,4,5,6,7,8,9],
				"friends": [
				  {"id": 0,"name": "Mcfarland Pickett"},
				  {"id": 1,"name": "Briana Avery"},
				  {"id": 2,"name": "Noel Hobbs"}
				],
				"greeting": "Hello, Randall! You have 8 unread messages.",
				"favoriteFruit": "apple"
			  },
			  {
				"_id": "5f9deaa1174da743ea0f9cab",
				"index": 2,
				"guid": "f3c1a57c-622c-413f-a55b-fea9c741feb7",
				"isActive": false,
				"balance": "$2,843.35",
				"picture": "http://placehold.it/32x32",
				"age": 21,
				"eyeColor": "brown",
				"name": {
				  "first": "Pam",
				  "last": "Stein"
				},
				"company": "QUILTIGEN",
				"email": "pam.stein@quiltigen.biz",
				"phone": "+1 (854) 493-3172",
				"address": "716 Sackman Street, Riner, Rhode Island, 4681",
				"about": "Nisi incididunt deserunt irure non excepteur sint amet tempor irure Lorem veniam cillum et in.",
				"registered": "Monday, June 30, 2014 3:51 AM",
				"latitude": "-42.93735",
				"longitude": "-39.854437",
				"tags": ["tempor","eiusmod","pariatur","mollit","Lorem"
				],
				"range": [0,1,2,3,4,5,6,7,8,9],
				"friends": [
				  {"id": 0,"name": "Jewel Page"},
				  {"id": 1,"name": "Merle Fernandez"},
				  {"id": 2,"name": "Maynard Cohen"}
				],
				"greeting": "Hello, Pam! You have 9 unread messages.",
				"favoriteFruit": "apple"
			  },
			  {
				"_id": "5f9deaa11b81c7fcdd208a25",
				"index": 3,
				"guid": "04e86b5b-76a8-4841-953a-1dd617ff0f6d",
				"isActive": true,
				"balance": "$2,774.83",
				"picture": "http://placehold.it/32x32",
				"age": 31,
				"eyeColor": "brown",
				"name": {
				  "first": "Margaret",
				  "last": "Lamb"
				},
				"company": "VIAGRAND",
				"email": "margaret.lamb@viagrand.name",
				"phone": "+1 (830) 474-2690",
				"address": "748 Vandalia Avenue, Brule, Nevada, 1252",
				"about": "Pariatur nisi minim nostrud irure veniam reprehenderit excepteur eu duis.",
				"registered": "Friday, July 10, 2020 6:12 AM",
				"latitude": "-41.341565",
				"longitude": "159.003298",
				"tags": ["anim","velit","irure","adipisicing","nulla"],
				"range": [0,1,2,3,4,5,6,7,8,9],
				"friends": [
				  {"id": 0,"name": "Monroe Roth"},
				  {"id": 1,"name": "Mullen Rhodes"},
				  {"id": 2,"name": "Mcclure Welch"}
				],
				"greeting": "Hello, Margaret! You have 10 unread messages.",
				"favoriteFruit": "banana"
			  },
			  {
				"_id": "5f9deaa1551e387447129d68",
				"index": 4,
				"guid": "3698670f-afb5-441a-abdd-bf3b8fb805ff",
				"isActive": true,
				"balance": "$1,036.21",
				"picture": "http://placehold.it/32x32",
				"age": 34,
				"eyeColor": "green",
				"name": {
				  "first": "Sherman",
				  "last": "Stone"
				},
				"company": "GEEKOSIS",
				"email": "sherman.stone@geekosis.org",
				"phone": "+1 (903) 461-3017",
				"address": "259 Rapelye Street, Sandston, Connecticut, 8045",
				"about": "Consequat incididunt aliqua laboris qui. Ex minim voluptate et nostrud.",
				"registered": "Friday, October 31, 2014 4:57 PM",
				"latitude": "9.795218",
				"longitude": "-65.600476",
				"tags": ["cillum","labore","incididunt","sit","amet"
				],
				"range": [0,1,2,3,4,5,6,7,8,9],
				"friends": [
				  {"id": 0,"name": "Whitehead Pugh"},
				  {"id": 1,"name": "Mable Villarreal"},
				  {"id": 2,"name": "Lowery Floyd"}
				],
				"greeting": "Hello, Sherman! You have 5 unread messages.",
				"favoriteFruit": "strawberry"
			  },
			  {
				"_id": "5f9deaa13006ef7b913a5cc0",
				"index": 5,
				"guid": "63df9954-ec43-4fda-b7a5-214a88681b3a",
				"isActive": true,
				"balance": "$3,870.62",
				"picture": "http://placehold.it/32x32",
				"age": 21,
				"eyeColor": "brown",
				"name": {
				  "first": "Duran",
				  "last": "Small"
				},
				"company": "TALKOLA",
				"email": "duran.small@talkola.info",
				"phone": "+1 (926) 516-2079",
				"address": "435 Calyer Street, Sperryville, Oregon, 112",
				"about": "Nostrud aliquip consequat do sint ipsum pariatur ut mollit.",
				"registered": "Monday, September 29, 2014 1:24 PM",
				"latitude": "16.614664",
				"longitude": "-131.206816",
				"tags": ["quis","duis","enim","culpa","est"],
				"range": [0,1,2,3,4,5,6,7,8,9],
				"friends": [
				  {"id": 0,"name": "Maritza Gordon"},
				  {"id": 1,"name": "Araceli Carey"},
				  {"id": 2,"name": "Stark Payne"}
				],
				"greeting": "Hello, Duran! You have 10 unread messages.",
				"favoriteFruit": "strawberry"
			  },
			  {
				"_id": "5f9deaa15a54a8ab95b82e76",
				"index": 6,
				"guid": "c34629ad-de7a-4f88-b62d-99505ab9bcc5",
				"isActive": false,
				"balance": "$1,692.48",
				"picture": "http://placehold.it/32x32",
				"age": 30,
				"eyeColor": "brown",
				"name": {
				  "first": "Alana",
				  "last": "Hart"
				},
				"company": "SKYPLEX",
				"email": "alana.hart@skyplex.ca",
				"phone": "+1 (824) 504-3286",
				"address": "285 Veranda Place, Kula, Wisconsin, 3318",
				"about": "Ea dolore enim proident sint do commodo irure reprehenderit fugiat.",
				"registered": "Sunday, April 10, 2016 11:35 AM",
				"latitude": "42.68548",
				"longitude": "68.660711",
				"tags": ["nostrud","fugiat","aute","labore","et"],
				"range": [0,1,2,3,4,5,6,7,8,9],
				"friends": [
				  {"id": 0,"name": "Atkinson Thomas"},
				  {"id": 1,"name": "Iva Curtis"},
				  {"id": 2,"name": "King Burnett"}
				],
				"greeting": "Hello, Alana! You have 8 unread messages.",
				"favoriteFruit": "apple"
			  }
			]`),
		propertyList: PropertyEntryList{
			{KeyURI: "objectID/[0.7]/age/float64", Value: Float64ToBinary(29)},
			{KeyURI: "objectID/[0.7]/email/string", Value: []byte("rosario.camacho@teraprene.co.uk")},
			{KeyURI: "objectID/[0.7]/phone/string", Value: []byte("+1 (940) 577-3244")},
			{KeyURI: "objectID/[0.7]/registered/string", Value: []byte("Friday, November 7, 2014 7:18 PM")},
			{KeyURI: "objectID/[0.7]/range/[0.10]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[0.7]/range/[1.10]/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[0.7]/range/[2.10]/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[0.7]/range/[3.10]/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[0.7]/range/[4.10]/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[0.7]/range/[5.10]/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[0.7]/range/[6.10]/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[0.7]/range/[7.10]/float64", Value: Float64ToBinary(7)},
			{KeyURI: "objectID/[0.7]/range/[8.10]/float64", Value: Float64ToBinary(8)},
			{KeyURI: "objectID/[0.7]/range/[9.10]/float64", Value: Float64ToBinary(9)},
			{KeyURI: "objectID/[0.7]/_id/string", Value: []byte("5f9deaa12b81ec174f75e315")},
			{KeyURI: "objectID/[0.7]/isActive/bool", Value: []byte("true")},
			{KeyURI: "objectID/[0.7]/picture/string", Value: []byte("http://placehold.it/32x32")},
			{KeyURI: "objectID/[0.7]/friends/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[0.7]/friends/[0.3]/name/string", Value: []byte("Clay Nash")},
			{KeyURI: "objectID/[0.7]/friends/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[0.7]/friends/[1.3]/name/string", Value: []byte("Amanda Warner")},
			{KeyURI: "objectID/[0.7]/friends/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[0.7]/friends/[2.3]/name/string", Value: []byte("Kirsten Whitehead")},
			{KeyURI: "objectID/[0.7]/index/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[0.7]/longitude/string", Value: []byte("151.382606")},
			{KeyURI: "objectID/[0.7]/address/string", Value: []byte("766 Keap Street, Smock, California, 1121")},
			{KeyURI: "objectID/[0.7]/latitude/string", Value: []byte("-16.92488")},
			{KeyURI: "objectID/[0.7]/greeting/string", Value: []byte("Hello, Rosario! You have 7 unread messages.")},
			{KeyURI: "objectID/[0.7]/favoriteFruit/string", Value: []byte("apple")},
			{KeyURI: "objectID/[0.7]/eyeColor/string", Value: []byte("blue")},
			{KeyURI: "objectID/[0.7]/name/last/string", Value: []byte("Camacho")},
			{KeyURI: "objectID/[0.7]/name/first/string", Value: []byte("Rosario")},
			{KeyURI: "objectID/[0.7]/company/string", Value: []byte("TERAPRENE")},
			{KeyURI: "objectID/[0.7]/tags/[0.5]/string", Value: []byte("non")},
			{KeyURI: "objectID/[0.7]/tags/[1.5]/string", Value: []byte("anim")},
			{KeyURI: "objectID/[0.7]/tags/[2.5]/string", Value: []byte("esse")},
			{KeyURI: "objectID/[0.7]/tags/[3.5]/string", Value: []byte("nostrud")},
			{KeyURI: "objectID/[0.7]/tags/[4.5]/string", Value: []byte("veniam")},
			{KeyURI: "objectID/[0.7]/guid/string", Value: []byte("5e3ba5a0-7d5f-4b70-bf41-7c882b4da1ef")},
			{KeyURI: "objectID/[0.7]/balance/string", Value: []byte("$3,166.17")},
			{KeyURI: "objectID/[0.7]/about/string", Value: []byte("Pariatur sint do pariatur dolor eiusmod reprehenderit non ex minim ullamco quis consequat.")},
			{KeyURI: "objectID/[1.7]/picture/string", Value: []byte("http://placehold.it/32x32")},
			{KeyURI: "objectID/[1.7]/eyeColor/string", Value: []byte("green")},
			{KeyURI: "objectID/[1.7]/registered/string", Value: []byte("Wednesday, April 1, 2015 7:06 AM")},
			{KeyURI: "objectID/[1.7]/_id/string", Value: []byte("5f9deaa1feed31fcbcc6e53e")},
			{KeyURI: "objectID/[1.7]/isActive/bool", Value: []byte("true")},
			{KeyURI: "objectID/[1.7]/age/float64", Value: Float64ToBinary(31)},
			{KeyURI: "objectID/[1.7]/email/string", Value: []byte("randall.conley@besto.me")},
			{KeyURI: "objectID/[1.7]/friends/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[1.7]/friends/[0.3]/name/string", Value: []byte("Mcfarland Pickett")},
			{KeyURI: "objectID/[1.7]/friends/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[1.7]/friends/[1.3]/name/string", Value: []byte("Briana Avery")},
			{KeyURI: "objectID/[1.7]/friends/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[1.7]/friends/[2.3]/name/string", Value: []byte("Noel Hobbs")},
			{KeyURI: "objectID/[1.7]/guid/string", Value: []byte("b9de6d8c-365e-4ca4-8ddb-218e1ce6f112")},
			{KeyURI: "objectID/[1.7]/balance/string", Value: []byte("$1,739.73")},
			{KeyURI: "objectID/[1.7]/company/string", Value: []byte("BESTO")},
			{KeyURI: "objectID/[1.7]/phone/string", Value: []byte("+1 (891) 595-2961")},
			{KeyURI: "objectID/[1.7]/about/string", Value: []byte("Exercitation cillum sint dolore aute ex anim deserunt veniam excepteur.")},
			{KeyURI: "objectID/[1.7]/longitude/string", Value: []byte("81.917956")},
			{KeyURI: "objectID/[1.7]/tags/[0.5]/string", Value: []byte("ut")},
			{KeyURI: "objectID/[1.7]/tags/[1.5]/string", Value: []byte("exercitation")},
			{KeyURI: "objectID/[1.7]/tags/[2.5]/string", Value: []byte("nostrud")},
			{KeyURI: "objectID/[1.7]/tags/[3.5]/string", Value: []byte("consequat")},
			{KeyURI: "objectID/[1.7]/tags/[4.5]/string", Value: []byte("esse")},
			{KeyURI: "objectID/[1.7]/range/[0.10]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[1.7]/range/[1.10]/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[1.7]/range/[2.10]/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[1.7]/range/[3.10]/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[1.7]/range/[4.10]/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[1.7]/range/[5.10]/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[1.7]/range/[6.10]/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[1.7]/range/[7.10]/float64", Value: Float64ToBinary(7)},
			{KeyURI: "objectID/[1.7]/range/[8.10]/float64", Value: Float64ToBinary(8)},
			{KeyURI: "objectID/[1.7]/range/[9.10]/float64", Value: Float64ToBinary(9)},
			{KeyURI: "objectID/[1.7]/index/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[1.7]/greeting/string", Value: []byte("Hello, Randall! You have 8 unread messages.")},
			{KeyURI: "objectID/[1.7]/address/string", Value: []byte("682 Meserole Avenue, Sexton, Georgia, 7532")},
			{KeyURI: "objectID/[1.7]/latitude/string", Value: []byte("-61.200296")},
			{KeyURI: "objectID/[1.7]/favoriteFruit/string", Value: []byte("apple")},
			{KeyURI: "objectID/[1.7]/name/last/string", Value: []byte("Conley")},
			{KeyURI: "objectID/[1.7]/name/first/string", Value: []byte("Randall")},
			{KeyURI: "objectID/[2.7]/balance/string", Value: []byte("$2,843.35")},
			{KeyURI: "objectID/[2.7]/address/string", Value: []byte("716 Sackman Street, Riner, Rhode Island, 4681")},
			{KeyURI: "objectID/[2.7]/longitude/string", Value: []byte("-39.854437")},
			{KeyURI: "objectID/[2.7]/friends/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[2.7]/friends/[0.3]/name/string", Value: []byte("Jewel Page")},
			{KeyURI: "objectID/[2.7]/friends/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[2.7]/friends/[1.3]/name/string", Value: []byte("Merle Fernandez")},
			{KeyURI: "objectID/[2.7]/friends/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[2.7]/friends/[2.3]/name/string", Value: []byte("Maynard Cohen")},
			{KeyURI: "objectID/[2.7]/index/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[2.7]/guid/string", Value: []byte("f3c1a57c-622c-413f-a55b-fea9c741feb7")},
			{KeyURI: "objectID/[2.7]/age/float64", Value: Float64ToBinary(21)},
			{KeyURI: "objectID/[2.7]/eyeColor/string", Value: []byte("brown")},
			{KeyURI: "objectID/[2.7]/latitude/string", Value: []byte("-42.93735")},
			{KeyURI: "objectID/[2.7]/favoriteFruit/string", Value: []byte("apple")},
			{KeyURI: "objectID/[2.7]/_id/string", Value: []byte("5f9deaa1174da743ea0f9cab")},
			{KeyURI: "objectID/[2.7]/picture/string", Value: []byte("http://placehold.it/32x32")},
			{KeyURI: "objectID/[2.7]/name/first/string", Value: []byte("Pam")},
			{KeyURI: "objectID/[2.7]/name/last/string", Value: []byte("Stein")},
			{KeyURI: "objectID/[2.7]/company/string", Value: []byte("QUILTIGEN")},
			{KeyURI: "objectID/[2.7]/about/string", Value: []byte("Nisi incididunt deserunt irure non excepteur sint amet tempor irure Lorem veniam cillum et in.")},
			{KeyURI: "objectID/[2.7]/tags/[0.5]/string", Value: []byte("tempor")},
			{KeyURI: "objectID/[2.7]/tags/[1.5]/string", Value: []byte("eiusmod")},
			{KeyURI: "objectID/[2.7]/tags/[2.5]/string", Value: []byte("pariatur")},
			{KeyURI: "objectID/[2.7]/tags/[3.5]/string", Value: []byte("mollit")},
			{KeyURI: "objectID/[2.7]/tags/[4.5]/string", Value: []byte("Lorem")},
			{KeyURI: "objectID/[2.7]/range/[0.10]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[2.7]/range/[1.10]/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[2.7]/range/[2.10]/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[2.7]/range/[3.10]/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[2.7]/range/[4.10]/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[2.7]/range/[5.10]/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[2.7]/range/[6.10]/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[2.7]/range/[7.10]/float64", Value: Float64ToBinary(7)},
			{KeyURI: "objectID/[2.7]/range/[8.10]/float64", Value: Float64ToBinary(8)},
			{KeyURI: "objectID/[2.7]/range/[9.10]/float64", Value: Float64ToBinary(9)},
			{KeyURI: "objectID/[2.7]/isActive/bool", Value: []byte("false")},
			{KeyURI: "objectID/[2.7]/phone/string", Value: []byte("+1 (854) 493-3172")},
			{KeyURI: "objectID/[2.7]/registered/string", Value: []byte("Monday, June 30, 2014 3:51 AM")},
			{KeyURI: "objectID/[2.7]/greeting/string", Value: []byte("Hello, Pam! You have 9 unread messages.")},
			{KeyURI: "objectID/[2.7]/email/string", Value: []byte("pam.stein@quiltigen.biz")},
			{KeyURI: "objectID/[3.7]/email/string", Value: []byte("margaret.lamb@viagrand.name")},
			{KeyURI: "objectID/[3.7]/about/string", Value: []byte("Pariatur nisi minim nostrud irure veniam reprehenderit excepteur eu duis.")},
			{KeyURI: "objectID/[3.7]/range/[0.10]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[3.7]/range/[1.10]/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[3.7]/range/[2.10]/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[3.7]/range/[3.10]/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[3.7]/range/[4.10]/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[3.7]/range/[5.10]/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[3.7]/range/[6.10]/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[3.7]/range/[7.10]/float64", Value: Float64ToBinary(7)},
			{KeyURI: "objectID/[3.7]/range/[8.10]/float64", Value: Float64ToBinary(8)},
			{KeyURI: "objectID/[3.7]/range/[9.10]/float64", Value: Float64ToBinary(9)},
			{KeyURI: "objectID/[3.7]/picture/string", Value: []byte("http://placehold.it/32x32")},
			{KeyURI: "objectID/[3.7]/age/float64", Value: Float64ToBinary(31)},
			{KeyURI: "objectID/[3.7]/company/string", Value: []byte("VIAGRAND")},
			{KeyURI: "objectID/[3.7]/guid/string", Value: []byte("04e86b5b-76a8-4841-953a-1dd617ff0f6d")},
			{KeyURI: "objectID/[3.7]/balance/string", Value: []byte("$2,774.83")},
			{KeyURI: "objectID/[3.7]/name/first/string", Value: []byte("Margaret")},
			{KeyURI: "objectID/[3.7]/name/last/string", Value: []byte("Lamb")},
			{KeyURI: "objectID/[3.7]/eyeColor/string", Value: []byte("brown")},
			{KeyURI: "objectID/[3.7]/phone/string", Value: []byte("+1 (830) 474-2690")},
			{KeyURI: "objectID/[3.7]/address/string", Value: []byte("748 Vandalia Avenue, Brule, Nevada, 1252")},
			{KeyURI: "objectID/[3.7]/greeting/string", Value: []byte("Hello, Margaret! You have 10 unread messages.")},
			{KeyURI: "objectID/[3.7]/favoriteFruit/string", Value: []byte("banana")},
			{KeyURI: "objectID/[3.7]/_id/string", Value: []byte("5f9deaa11b81c7fcdd208a25")},
			{KeyURI: "objectID/[3.7]/index/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[3.7]/isActive/bool", Value: []byte("true")},
			{KeyURI: "objectID/[3.7]/tags/[0.5]/string", Value: []byte("anim")},
			{KeyURI: "objectID/[3.7]/tags/[1.5]/string", Value: []byte("velit")},
			{KeyURI: "objectID/[3.7]/tags/[2.5]/string", Value: []byte("irure")},
			{KeyURI: "objectID/[3.7]/tags/[3.5]/string", Value: []byte("adipisicing")},
			{KeyURI: "objectID/[3.7]/tags/[4.5]/string", Value: []byte("nulla")},
			{KeyURI: "objectID/[3.7]/friends/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[3.7]/friends/[0.3]/name/string", Value: []byte("Monroe Roth")},
			{KeyURI: "objectID/[3.7]/friends/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[3.7]/friends/[1.3]/name/string", Value: []byte("Mullen Rhodes")},
			{KeyURI: "objectID/[3.7]/friends/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[3.7]/friends/[2.3]/name/string", Value: []byte("Mcclure Welch")},
			{KeyURI: "objectID/[3.7]/registered/string", Value: []byte("Friday, July 10, 2020 6:12 AM")},
			{KeyURI: "objectID/[3.7]/latitude/string", Value: []byte("-41.341565")},
			{KeyURI: "objectID/[3.7]/longitude/string", Value: []byte("159.003298")},
			{KeyURI: "objectID/[4.7]/about/string", Value: []byte("Consequat incididunt aliqua laboris qui. Ex minim voluptate et nostrud.")},
			{KeyURI: "objectID/[4.7]/favoriteFruit/string", Value: []byte("strawberry")},
			{KeyURI: "objectID/[4.7]/index/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[4.7]/guid/string", Value: []byte("3698670f-afb5-441a-abdd-bf3b8fb805ff")},
			{KeyURI: "objectID/[4.7]/name/first/string", Value: []byte("Sherman")},
			{KeyURI: "objectID/[4.7]/name/last/string", Value: []byte("Stone")},
			{KeyURI: "objectID/[4.7]/company/string", Value: []byte("GEEKOSIS")},
			{KeyURI: "objectID/[4.7]/longitude/string", Value: []byte("-65.600476")},
			{KeyURI: "objectID/[4.7]/greeting/string", Value: []byte("Hello, Sherman! You have 5 unread messages.")},
			{KeyURI: "objectID/[4.7]/isActive/bool", Value: []byte("true")},
			{KeyURI: "objectID/[4.7]/picture/string", Value: []byte("http://placehold.it/32x32")},
			{KeyURI: "objectID/[4.7]/eyeColor/string", Value: []byte("green")},
			{KeyURI: "objectID/[4.7]/email/string", Value: []byte("sherman.stone@geekosis.org")},
			{KeyURI: "objectID/[4.7]/latitude/string", Value: []byte("9.795218")},
			{KeyURI: "objectID/[4.7]/range/[0.10]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[4.7]/range/[1.10]/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[4.7]/range/[2.10]/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[4.7]/range/[3.10]/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[4.7]/range/[4.10]/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[4.7]/range/[5.10]/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[4.7]/range/[6.10]/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[4.7]/range/[7.10]/float64", Value: Float64ToBinary(7)},
			{KeyURI: "objectID/[4.7]/range/[8.10]/float64", Value: Float64ToBinary(8)},
			{KeyURI: "objectID/[4.7]/range/[9.10]/float64", Value: Float64ToBinary(9)},
			{KeyURI: "objectID/[4.7]/_id/string", Value: []byte("5f9deaa1551e387447129d68")},
			{KeyURI: "objectID/[4.7]/age/float64", Value: Float64ToBinary(34)},
			{KeyURI: "objectID/[4.7]/address/string", Value: []byte("259 Rapelye Street, Sandston, Connecticut, 8045")},
			{KeyURI: "objectID/[4.7]/registered/string", Value: []byte("Friday, October 31, 2014 4:57 PM")},
			{KeyURI: "objectID/[4.7]/tags/[0.5]/string", Value: []byte("cillum")},
			{KeyURI: "objectID/[4.7]/tags/[1.5]/string", Value: []byte("labore")},
			{KeyURI: "objectID/[4.7]/tags/[2.5]/string", Value: []byte("incididunt")},
			{KeyURI: "objectID/[4.7]/tags/[3.5]/string", Value: []byte("sit")},
			{KeyURI: "objectID/[4.7]/tags/[4.5]/string", Value: []byte("amet")},
			{KeyURI: "objectID/[4.7]/friends/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[4.7]/friends/[0.3]/name/string", Value: []byte("Whitehead Pugh")},
			{KeyURI: "objectID/[4.7]/friends/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[4.7]/friends/[1.3]/name/string", Value: []byte("Mable Villarreal")},
			{KeyURI: "objectID/[4.7]/friends/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[4.7]/friends/[2.3]/name/string", Value: []byte("Lowery Floyd")},
			{KeyURI: "objectID/[4.7]/balance/string", Value: []byte("$1,036.21")},
			{KeyURI: "objectID/[4.7]/phone/string", Value: []byte("+1 (903) 461-3017")},
			{KeyURI: "objectID/[5.7]/phone/string", Value: []byte("+1 (926) 516-2079")},
			{KeyURI: "objectID/[5.7]/range/[0.10]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[5.7]/range/[1.10]/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[5.7]/range/[2.10]/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[5.7]/range/[3.10]/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[5.7]/range/[4.10]/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[5.7]/range/[5.10]/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[5.7]/range/[6.10]/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[5.7]/range/[7.10]/float64", Value: Float64ToBinary(7)},
			{KeyURI: "objectID/[5.7]/range/[8.10]/float64", Value: Float64ToBinary(8)},
			{KeyURI: "objectID/[5.7]/range/[9.10]/float64", Value: Float64ToBinary(9)},
			{KeyURI: "objectID/[5.7]/isActive/bool", Value: []byte("true")},
			{KeyURI: "objectID/[5.7]/balance/string", Value: []byte("$3,870.62")},
			{KeyURI: "objectID/[5.7]/email/string", Value: []byte("duran.small@talkola.info")},
			{KeyURI: "objectID/[5.7]/about/string", Value: []byte("Nostrud aliquip consequat do sint ipsum pariatur ut mollit.")},
			{KeyURI: "objectID/[5.7]/latitude/string", Value: []byte("16.614664")},
			{KeyURI: "objectID/[5.7]/longitude/string", Value: []byte("-131.206816")},
			{KeyURI: "objectID/[5.7]/tags/[0.5]/string", Value: []byte("quis")},
			{KeyURI: "objectID/[5.7]/tags/[1.5]/string", Value: []byte("duis")},
			{KeyURI: "objectID/[5.7]/tags/[2.5]/string", Value: []byte("enim")},
			{KeyURI: "objectID/[5.7]/tags/[3.5]/string", Value: []byte("culpa")},
			{KeyURI: "objectID/[5.7]/tags/[4.5]/string", Value: []byte("est")},
			{KeyURI: "objectID/[5.7]/friends/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[5.7]/friends/[0.3]/name/string", Value: []byte("Maritza Gordon")},
			{KeyURI: "objectID/[5.7]/friends/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[5.7]/friends/[1.3]/name/string", Value: []byte("Araceli Carey")},
			{KeyURI: "objectID/[5.7]/friends/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[5.7]/friends/[2.3]/name/string", Value: []byte("Stark Payne")},
			{KeyURI: "objectID/[5.7]/guid/string", Value: []byte("63df9954-ec43-4fda-b7a5-214a88681b3a")},
			{KeyURI: "objectID/[5.7]/age/float64", Value: Float64ToBinary(21)},
			{KeyURI: "objectID/[5.7]/address/string", Value: []byte("435 Calyer Street, Sperryville, Oregon, 112")},
			{KeyURI: "objectID/[5.7]/greeting/string", Value: []byte("Hello, Duran! You have 10 unread messages.")},
			{KeyURI: "objectID/[5.7]/favoriteFruit/string", Value: []byte("strawberry")},
			{KeyURI: "objectID/[5.7]/index/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[5.7]/picture/string", Value: []byte("http://placehold.it/32x32")},
			{KeyURI: "objectID/[5.7]/company/string", Value: []byte("TALKOLA")},
			{KeyURI: "objectID/[5.7]/registered/string", Value: []byte("Monday, September 29, 2014 1:24 PM")},
			{KeyURI: "objectID/[5.7]/_id/string", Value: []byte("5f9deaa13006ef7b913a5cc0")},
			{KeyURI: "objectID/[5.7]/eyeColor/string", Value: []byte("brown")},
			{KeyURI: "objectID/[5.7]/name/first/string", Value: []byte("Duran")},
			{KeyURI: "objectID/[5.7]/name/last/string", Value: []byte("Small")},
			{KeyURI: "objectID/[6.7]/index/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[6.7]/email/string", Value: []byte("alana.hart@skyplex.ca")},
			{KeyURI: "objectID/[6.7]/registered/string", Value: []byte("Sunday, April 10, 2016 11:35 AM")},
			{KeyURI: "objectID/[6.7]/favoriteFruit/string", Value: []byte("apple")},
			{KeyURI: "objectID/[6.7]/range/[0.10]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[6.7]/range/[1.10]/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[6.7]/range/[2.10]/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[6.7]/range/[3.10]/float64", Value: Float64ToBinary(3)},
			{KeyURI: "objectID/[6.7]/range/[4.10]/float64", Value: Float64ToBinary(4)},
			{KeyURI: "objectID/[6.7]/range/[5.10]/float64", Value: Float64ToBinary(5)},
			{KeyURI: "objectID/[6.7]/range/[6.10]/float64", Value: Float64ToBinary(6)},
			{KeyURI: "objectID/[6.7]/range/[7.10]/float64", Value: Float64ToBinary(7)},
			{KeyURI: "objectID/[6.7]/range/[8.10]/float64", Value: Float64ToBinary(8)},
			{KeyURI: "objectID/[6.7]/range/[9.10]/float64", Value: Float64ToBinary(9)},
			{KeyURI: "objectID/[6.7]/_id/string", Value: []byte("5f9deaa15a54a8ab95b82e76")},
			{KeyURI: "objectID/[6.7]/balance/string", Value: []byte("$1,692.48")},
			{KeyURI: "objectID/[6.7]/company/string", Value: []byte("SKYPLEX")},
			{KeyURI: "objectID/[6.7]/phone/string", Value: []byte("+1 (824) 504-3286")},
			{KeyURI: "objectID/[6.7]/about/string", Value: []byte("Ea dolore enim proident sint do commodo irure reprehenderit fugiat.")},
			{KeyURI: "objectID/[6.7]/longitude/string", Value: []byte("68.660711")},
			{KeyURI: "objectID/[6.7]/tags/[0.5]/string", Value: []byte("nostrud")},
			{KeyURI: "objectID/[6.7]/tags/[1.5]/string", Value: []byte("fugiat")},
			{KeyURI: "objectID/[6.7]/tags/[2.5]/string", Value: []byte("aute")},
			{KeyURI: "objectID/[6.7]/tags/[3.5]/string", Value: []byte("labore")},
			{KeyURI: "objectID/[6.7]/tags/[4.5]/string", Value: []byte("et")},
			{KeyURI: "objectID/[6.7]/guid/string", Value: []byte("c34629ad-de7a-4f88-b62d-99505ab9bcc5")},
			{KeyURI: "objectID/[6.7]/isActive/bool", Value: []byte("false")},
			{KeyURI: "objectID/[6.7]/eyeColor/string", Value: []byte("brown")},
			{KeyURI: "objectID/[6.7]/name/first/string", Value: []byte("Alana")},
			{KeyURI: "objectID/[6.7]/name/last/string", Value: []byte("Hart")},
			{KeyURI: "objectID/[6.7]/address/string", Value: []byte("285 Veranda Place, Kula, Wisconsin, 3318")},
			{KeyURI: "objectID/[6.7]/picture/string", Value: []byte("http://placehold.it/32x32")},
			{KeyURI: "objectID/[6.7]/age/float64", Value: Float64ToBinary(30)},
			{KeyURI: "objectID/[6.7]/latitude/string", Value: []byte("42.68548")},
			{KeyURI: "objectID/[6.7]/friends/[0.3]/id/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/[6.7]/friends/[0.3]/name/string", Value: []byte("Atkinson Thomas")},
			{KeyURI: "objectID/[6.7]/friends/[1.3]/id/float64", Value: Float64ToBinary(1)},
			{KeyURI: "objectID/[6.7]/friends/[1.3]/name/string", Value: []byte("Iva Curtis")},
			{KeyURI: "objectID/[6.7]/friends/[2.3]/id/float64", Value: Float64ToBinary(2)},
			{KeyURI: "objectID/[6.7]/friends/[2.3]/name/string", Value: []byte("King Burnett")},
			{KeyURI: "objectID/[6.7]/greeting/string", Value: []byte("Hello, Alana! You have 8 unread messages.")},
		},
	},
	"Transform object with chained arrays": {
		prefix: "objectID",
		jsonPayload: []byte(`{
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
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[0.5]/[0.3]/float64", Value: Float64ToBinary(-122.42200352825247)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[0.5]/[1.3]/float64", Value: Float64ToBinary(37.80848009696725)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[0.5]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[1.5]/[0.3]/float64", Value: Float64ToBinary(-122.42207601332528)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[1.5]/[1.3]/float64", Value: Float64ToBinary(37.808835019815085)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[1.5]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[2.5]/[0.3]/float64", Value: Float64ToBinary(-122.42110217434863)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[2.5]/[1.3]/float64", Value: Float64ToBinary(37.808803534992904)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[2.5]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[3.5]/[0.3]/float64", Value: Float64ToBinary(-122.42106256906727)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[3.5]/[1.3]/float64", Value: Float64ToBinary(37.80860105681815)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[3.5]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[4.5]/[0.3]/float64", Value: Float64ToBinary(-122.42200352825247)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[4.5]/[1.3]/float64", Value: Float64ToBinary(37.80848009696725)},
			{KeyURI: "objectID/features/[0.2]/geometry/coordinates/[0.1]/[4.5]/[2.3]/float64", Value: Float64ToBinary(0)},
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
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[0.4]/[0.3]/float64", Value: Float64ToBinary(-122.41570120460688)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[0.4]/[1.3]/float64", Value: Float64ToBinary(37.80832725267146)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[0.4]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[1.4]/[0.3]/float64", Value: Float64ToBinary(-122.4157607435932)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[1.4]/[1.3]/float64", Value: Float64ToBinary(37.808630700240904)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[1.4]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[2.4]/[0.3]/float64", Value: Float64ToBinary(-122.4137878913324)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[2.4]/[1.3]/float64", Value: Float64ToBinary(37.80856680131984)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[2.4]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[3.4]/[0.3]/float64", Value: Float64ToBinary(-122.41570120460688)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[3.4]/[1.3]/float64", Value: Float64ToBinary(37.80832725267146)},
			{KeyURI: "objectID/features/[1.2]/geometry/coordinates/[0.1]/[3.4]/[2.3]/float64", Value: Float64ToBinary(0)},
			{KeyURI: "objectID/features/[1.2]/geometry/type/string", Value: []byte("Polygon")},
		},
	},
}

func TestCreatePropertyListFromRaw(t *testing.T) {
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			gotList, err := RawToPropertyList(test.prefix, bytes.NewReader(test.jsonPayload))
			if err != nil {
				t.Fatalf("unexpected error :%v", err)
			}

			assert.ElementsMatch(t, gotList, test.propertyList, "list should match")
		})
	}
}

func TestCreatePropertyListFromRaw_BadReader(t *testing.T) {
	_, err := RawToPropertyList("somePrefix", bytes.NewReader([]byte("")))
	if err == nil {
		t.Fatalf("expected error :%v", err)
	}
	assert.Equal(t, errors.New("unable to unmarshall payload: EOF"), err)
}

func TestRemoveLastElement(t *testing.T) {
	tests := map[string]struct {
		slice    []string
		expSlice []string
	}{
		"Removed last": {
			slice:    []string{"a", "b", "c", "d", "e"},
			expSlice: []string{"a", "b", "c", "d"},
		},
		"Empty slice": {
			slice:    []string{},
			expSlice: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			orig := make([]string, len(test.slice))
			copy(orig, test.slice)
			removeLastElement(&orig)

			assert.Equal(t, test.expSlice, orig)
		})
	}
}
