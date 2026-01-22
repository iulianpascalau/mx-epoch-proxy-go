package common

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	t.Parallel()

	key := GenerateKey()
	assert.Equal(t, apiKeySize*2, len(key))
}

func TestAnonymizeKey(t *testing.T) {
	t.Parallel()

	t.Run("empty string should return empty string", func(t *testing.T) {
		assert.Empty(t, AnonymizeKey(""))
	})
	t.Run("small keys should anonymize completely", func(t *testing.T) {
		assert.Equal(t, "*", AnonymizeKey("0"))
		assert.Equal(t, "**", AnonymizeKey("01"))
		assert.Equal(t, "***", AnonymizeKey("012"))
		assert.Equal(t, "****", AnonymizeKey("0123"))
		assert.Equal(t, "*****", AnonymizeKey("01234"))
		assert.Equal(t, "******", AnonymizeKey("012345"))
		assert.Equal(t, "*******", AnonymizeKey("0123456"))
		assert.Equal(t, "********", AnonymizeKey("01234567"))
		assert.Equal(t, "*********", AnonymizeKey("012345678"))
	})
	t.Run("should work", func(t *testing.T) {
		assert.Equal(t, "012****789", AnonymizeKey("0123456789"))

		key := GenerateKey()
		fmt.Printf("Generated API key: %s\n", key)
		anonymizedKey := AnonymizeKey(key)
		fmt.Printf("Anonymized API key: %s\n", anonymizedKey)
		assert.Equal(t, key[:3]+"**************************"+key[29:], anonymizedKey)
	})
}

func TestCronJob(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		counter := uint64(0)
		handler := func() {
			atomic.AddUint64(&counter, 1)
		}

		CronJobStarter(ctx, handler, time.Millisecond*100)

		time.Sleep(time.Millisecond * 350) // 350ms => 3 calls => counter should be 3

		assert.Equal(t, uint64(3), atomic.LoadUint64(&counter))
	})
	t.Run("context done should stop", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		counter := uint64(0)
		handler := func() {
			atomic.AddUint64(&counter, 1)
		}

		CronJobStarter(ctx, handler, time.Millisecond*100)

		time.Sleep(time.Millisecond * 350) // 35oms => 3 calls => counter should be 3
		cancel()

		time.Sleep(time.Millisecond * 350) // wait another 350ms just to be safe

		assert.Equal(t, uint64(3), atomic.LoadUint64(&counter))
	})
}

func TestProcessUserDetailss(t *testing.T) {
	t.Parallel()

	t.Run("nil account should not panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("Panic occurred: %v", r))
			}
		}()

		ProcessUserDetails(nil)
	})
	t.Run("premium - unlimited", func(t *testing.T) {
		userDetails := &UsersDetails{
			DBAccountType:   PremiumAccountType,
			GlobalCounter:   100,
			MaxRequests:     0,
			CryptoPaymentID: 0,
		}

		ProcessUserDetails(userDetails)
		assert.Equal(t, PremiumAccountType, userDetails.ProcessedAccountType)
		assert.Equal(t, true, userDetails.IsUnlimited)
		assert.Equal(t, false, userDetails.CryptoPaymentInitiated)

		assert.Equal(t, uint64(0), userDetails.MaxRequests)
		assert.Equal(t, uint64(100), userDetails.GlobalCounter)
		assert.Equal(t, uint64(0), userDetails.CryptoPaymentID)
	})

	t.Run("premium - limited", func(t *testing.T) {
		userDetails := &UsersDetails{
			GlobalCounter:   100,
			MaxRequests:     200,
			CryptoPaymentID: 1,
		}

		ProcessUserDetails(userDetails)
		assert.Equal(t, PremiumAccountType, userDetails.ProcessedAccountType)
		assert.Equal(t, false, userDetails.IsUnlimited)
		assert.Equal(t, true, userDetails.CryptoPaymentInitiated)

		assert.Equal(t, uint64(200), userDetails.MaxRequests)
		assert.Equal(t, uint64(100), userDetails.GlobalCounter)
		assert.Equal(t, uint64(1), userDetails.CryptoPaymentID)
	})

	t.Run("premium - limited and max reached", func(t *testing.T) {
		userDetails := &UsersDetails{
			GlobalCounter:   200,
			MaxRequests:     200,
			CryptoPaymentID: 1,
		}

		ProcessUserDetails(userDetails)
		assert.Equal(t, FreeAccountType, userDetails.ProcessedAccountType)
		assert.Equal(t, false, userDetails.IsUnlimited)
		assert.Equal(t, true, userDetails.CryptoPaymentInitiated)

		assert.Equal(t, uint64(200), userDetails.MaxRequests)
		assert.Equal(t, uint64(200), userDetails.GlobalCounter)
		assert.Equal(t, uint64(1), userDetails.CryptoPaymentID)
	})

	t.Run("free - no payment id", func(t *testing.T) {
		userDetails := &UsersDetails{
			GlobalCounter:   0,
			MaxRequests:     0,
			CryptoPaymentID: 0,
		}

		ProcessUserDetails(userDetails)
		assert.Equal(t, FreeAccountType, userDetails.ProcessedAccountType)
		assert.Equal(t, false, userDetails.IsUnlimited)
		assert.Equal(t, false, userDetails.CryptoPaymentInitiated)

		assert.Equal(t, uint64(0), userDetails.MaxRequests)
		assert.Equal(t, uint64(0), userDetails.GlobalCounter)
		assert.Equal(t, uint64(0), userDetails.CryptoPaymentID)
	})

	t.Run("free - with payment ID but no payments", func(t *testing.T) {
		userDetails := &UsersDetails{
			GlobalCounter:   0,
			MaxRequests:     0,
			CryptoPaymentID: 1,
		}

		ProcessUserDetails(userDetails)
		assert.Equal(t, FreeAccountType, userDetails.ProcessedAccountType)
		assert.Equal(t, false, userDetails.IsUnlimited)
		assert.Equal(t, true, userDetails.CryptoPaymentInitiated)

		assert.Equal(t, uint64(0), userDetails.MaxRequests)
		assert.Equal(t, uint64(0), userDetails.GlobalCounter)
		assert.Equal(t, uint64(1), userDetails.CryptoPaymentID)
	})

	t.Run("free - with payment ID and with payment", func(t *testing.T) {
		userDetails := &UsersDetails{
			GlobalCounter:   0,
			MaxRequests:     100,
			CryptoPaymentID: 1,
		}

		ProcessUserDetails(userDetails)
		assert.Equal(t, PremiumAccountType, userDetails.ProcessedAccountType)
		assert.Equal(t, false, userDetails.IsUnlimited)
		assert.Equal(t, true, userDetails.CryptoPaymentInitiated)

		assert.Equal(t, uint64(100), userDetails.MaxRequests)
		assert.Equal(t, uint64(0), userDetails.GlobalCounter)
		assert.Equal(t, uint64(1), userDetails.CryptoPaymentID)
	})
	t.Run("free - with payment ID and with payment and max requests reached", func(t *testing.T) {
		userDetails := &UsersDetails{
			GlobalCounter:   100,
			MaxRequests:     100,
			CryptoPaymentID: 1,
		}

		ProcessUserDetails(userDetails)
		assert.Equal(t, FreeAccountType, userDetails.ProcessedAccountType)
		assert.Equal(t, false, userDetails.IsUnlimited)
		assert.Equal(t, true, userDetails.CryptoPaymentInitiated)

		assert.Equal(t, uint64(100), userDetails.MaxRequests)
		assert.Equal(t, uint64(100), userDetails.GlobalCounter)
		assert.Equal(t, uint64(1), userDetails.CryptoPaymentID)
	})
}
