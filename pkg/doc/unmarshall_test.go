package doc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRawFromPropertyList(t *testing.T) {
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			rawObject := PropertyListToRaw(test.propertyList)

			gotPayload, err := json.Marshal(rawObject)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.JSONEq(t, string(test.jsonPayload), string(gotPayload), "list should match")
		})
	}
}
