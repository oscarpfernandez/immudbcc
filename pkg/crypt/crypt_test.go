package crypt

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	token, err := GenerateEncryptionToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("Token: %s", token)

	for i := 0; i < 10; i++ {
		t.Logf("Test #%d", i)
		data := make([]byte, rand.Intn(1024)+1024)
		if _, err := rand.Read(data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		t.Logf("Sample Data:    %s", hex.EncodeToString(data[:50]))

		encData, err := Encrypt(data, token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("Sample EncData: %s", hex.EncodeToString(encData[:50]))

		decData, err := Decrypt(encData, token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("Sample DecData: %s", hex.EncodeToString(decData[:50]))

		assert.Equal(t, data, decData, "decrypted data should match original data")
	}
}
