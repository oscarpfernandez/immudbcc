package crypt

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	token, err := GenerateEncryptionToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 100; i++ {
		data := make([]byte, rand.Intn(1024)+1024)
		if _, err := rand.Read(data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		encData, err := Encrypt(data, token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		decData, err := Decrypt(encData, token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, data, decData, "decrypted data should match original data")
	}
}
