package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"unicode"
)

type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox(rawKey string) (*SecretBox, error) {

	keyBytes, err := decodeKey(rawKey)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecretBox{aead: aead}, nil
}

// decodeKey accepts either 32 raw bytes (string) or 64-char hex.
// It trims whitespace/quotes and tolerates a "0x" prefix for hex.

func decodeKey(k string) ([]byte, error) {
	if k == "" {
		return nil, errors.New("encryption key is empty (set ENCRYPTION_KEY)")
	}

	// SAFE DEBUG: log only length, not value
	println("[SecretBox] raw ENCRYPTION_KEY length:", len(k))

	// Normalize
	k = strings.TrimSpace(k)
	k = strings.Trim(k, `"'`)
	if strings.HasPrefix(k, "0x") || strings.HasPrefix(k, "0X") {
		k = k[2:]
	}

	println("[SecretBox] normalized key length:", len(k))

	if isLikelyHex(k) && len(k)%2 == 0 {
		b, err := hex.DecodeString(k)
		if err != nil {
			return nil, errors.New("invalid hex string: " + err.Error())
		}
		if len(b) != 32 {
			return nil, errors.New("hex key must represent exactly 32 bytes (64 hex chars)")
		}
		return b, nil
	}

	if len(k) == 32 {
		return []byte(k), nil
	}

	return nil, errors.New("encryption key must be 32 raw bytes or 64-char hex")
}

func isLikelyHex(s string) bool {
	for _, r := range s {
		if !(unicode.IsDigit(r) ||
			(r >= 'a' && r <= 'f') ||
			(r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func (s *SecretBox) Seal(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ct := s.aead.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ct...), nil // nonce || ciphertext
}

func (s *SecretBox) Open(data []byte) ([]byte, error) {
	ns := s.aead.NonceSize()
	if len(data) < ns {
		return nil, errors.New("ciphertext too short")
	}
	nonce := data[:ns]
	ct := data[ns:]
	return s.aead.Open(nil, nonce, ct, nil)
}
