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

func TestCreateAccountSettings(t *testing.T) {
	t.Parallel()

	t.Run("premium - unlimited", func(t *testing.T) {
		settings := CreateAccountSettings(string(PremiumAccountType), 100, 0, 0)
		assert.Equal(t, PremiumAccountType, settings.Type)
		assert.Equal(t, uint64(0), settings.MaxRequests)
		assert.Equal(t, uint64(100), settings.RequestCount)
		assert.Equal(t, false, settings.CryptoPaymentInitiated)
		assert.Equal(t, true, settings.IsUnlimited)
	})

	t.Run("premium - limited", func(t *testing.T) {
		settings := CreateAccountSettings(string(PremiumAccountType), 100, 200, 1)
		assert.Equal(t, PremiumAccountType, settings.Type)
		assert.Equal(t, uint64(200), settings.MaxRequests)
		assert.Equal(t, uint64(100), settings.RequestCount)
		assert.Equal(t, true, settings.CryptoPaymentInitiated)
		assert.Equal(t, false, settings.IsUnlimited)
	})

	t.Run("premium - limited and max reached", func(t *testing.T) {
		settings := CreateAccountSettings(string(PremiumAccountType), 200, 200, 1)
		assert.Equal(t, FreeAccountType, settings.Type)
		assert.Equal(t, uint64(200), settings.MaxRequests)
		assert.Equal(t, uint64(200), settings.RequestCount)
		assert.Equal(t, true, settings.CryptoPaymentInitiated)
		assert.Equal(t, false, settings.IsUnlimited)
	})

	t.Run("free - no payment id", func(t *testing.T) {
		settings := CreateAccountSettings(string(FreeAccountType), 0, 0, 0)
		assert.Equal(t, FreeAccountType, settings.Type)
		assert.Equal(t, uint64(0), settings.MaxRequests)
		assert.Equal(t, uint64(0), settings.RequestCount)
		assert.Equal(t, false, settings.CryptoPaymentInitiated)
		assert.Equal(t, false, settings.IsUnlimited)
	})

	t.Run("free - with payment ID but no payments", func(t *testing.T) {
		settings := CreateAccountSettings(string(FreeAccountType), 0, 0, 1)
		assert.Equal(t, FreeAccountType, settings.Type)
		assert.Equal(t, uint64(0), settings.MaxRequests)
		assert.Equal(t, uint64(0), settings.RequestCount)
		assert.Equal(t, true, settings.CryptoPaymentInitiated)
		assert.Equal(t, false, settings.IsUnlimited)
	})

	t.Run("free - with payment ID and with payment", func(t *testing.T) {
		settings := CreateAccountSettings(string(FreeAccountType), 0, 100, 1)
		assert.Equal(t, PremiumAccountType, settings.Type)
		assert.Equal(t, uint64(100), settings.MaxRequests)
		assert.Equal(t, uint64(0), settings.RequestCount)
		assert.Equal(t, true, settings.CryptoPaymentInitiated)
		assert.Equal(t, false, settings.IsUnlimited)
	})
}
