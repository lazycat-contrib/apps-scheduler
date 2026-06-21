package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const (
	Prefix      = "as_mcp_"
	displaySize = len(Prefix) + 8
	randomBytes = 32
)

func Generate() (string, error) {
	b := make([]byte, randomBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return Prefix + base64.RawURLEncoding.EncodeToString(b), nil
}

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func DisplayPrefix(token string) string {
	if len(token) <= displaySize {
		return token
	}
	return token[:displaySize]
}
