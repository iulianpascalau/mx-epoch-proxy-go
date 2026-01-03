package common

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
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

// CronJobStarter is able to start a go routine that periodically calls the provided handler. The time between calls is
// provided as timeToCall
func CronJobStarter(ctx context.Context, handler func(), timeToCall time.Duration) {
	go func() {
		timer := time.NewTimer(timeToCall)
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				handler()
				timer.Reset(timeToCall)
			case <-ctx.Done():
				return
			}
		}
	}()
}
