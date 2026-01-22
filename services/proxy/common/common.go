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

// ProcessUserDetails implements a high-level logic to set 3 fields on the provided user details object: the
// ProcessedAccountType, CryptoPaymentInitiated and IsUnlimited
func ProcessUserDetails(userDetails *UsersDetails) {
	if userDetails == nil {
		return
	}

	if userDetails.MaxRequests == 0 && userDetails.DBAccountType == PremiumAccountType {
		// the account is premium (no heavy throttling) with unlimited request
		userDetails.IsUnlimited = true
		userDetails.ProcessedAccountType = PremiumAccountType
		userDetails.CryptoPaymentInitiated = false

		return
	}

	userDetails.CryptoPaymentInitiated = userDetails.CryptoPaymentID > 0

	if userDetails.MaxRequests > 0 && userDetails.GlobalCounter < userDetails.MaxRequests {
		// the account is premium (no heavy throttling) with credits still left
		userDetails.IsUnlimited = false
		userDetails.ProcessedAccountType = PremiumAccountType

		return
	}

	// the account should be treated as free because the user depleted it's purchased credits but requests should still work although heavily throttled
	// or
	// the account was declared free and no purchase completed

	userDetails.IsUnlimited = false
	userDetails.ProcessedAccountType = FreeAccountType
}
