package common

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

const showFirstLetters = 3
const showLastLetters = 3
const additionalLetters = 3
const apiKeySize = 16

// AnonymizeKey will anonymize the provided key
func AnonymizeKey(key string) string {
	if len(key) <= showFirstLetters+showLastLetters+additionalLetters {
		return strings.Repeat("*", len(key))
	}

	numHiddenLetters := len(key) - showFirstLetters - showLastLetters
	return key[:showFirstLetters] + strings.Repeat("*", numHiddenLetters) + key[showFirstLetters+numHiddenLetters:]
}

// GenerateKey will generate a new key
func GenerateKey() string {
	buff := make([]byte, apiKeySize)
	_, _ = rand.Read(buff)

	return hex.EncodeToString(buff)
}
