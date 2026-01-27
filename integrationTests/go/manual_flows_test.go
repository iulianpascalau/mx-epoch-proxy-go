package _go

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var manualFlowsLog = logger.GetOrCreate("integrationTests")

func TestManualFlowsIsolation(t *testing.T) {
	if !framework.IsChainSimulatorIsRunning() {
		t.Skip("No chain simulator instance running found. Skipping slow test")
	}

	// 1. Setup Environment
	proxyService := framework.NewProxyService(t)
	cryptoPaymentService := framework.NewCryptoPaymentService(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proxyService.Setup(ctx)
	// High rate to avoid throttle issues during setup
	cryptoPaymentService.Setup(ctx, 1000)

	defer proxyService.TearDown()
	defer cryptoPaymentService.TearDown()

	cryptoPaymentService.CreateService()
	cryptoPaymentService.Components.StartCronJobs(ctx)
	cryptoPaymentServiceURL := "http://" + cryptoPaymentService.Components.GetHTTPServer().GetAddress()
	proxyService.CreateService(cryptoPaymentServiceURL)
	proxyService.Components.StartCronJobs(ctx)

	// Allow services to stabilize
	time.Sleep(time.Second)

	// =========================================================================
	// FLOW-002: Duplicate Address Request Prevention
	// =========================================================================
	t.Run("FLOW-002_DuplicateAddressPrevention", func(t *testing.T) {
		email := "user1@example.com"
		session := framework.NewTestSession(proxyService.Components, email, "password123456")
		session.Register(t)
		session.Activate(t, proxyService.EmailSender)
		session.Login(t)

		// 1. First request should succeed
		manualFlowsLog.Info("Step 1: Requesting first payment address...")
		session.ObtainDepositAddress(t)
		firstAddr := session.GetDepositAddress()
		require.NotEmpty(t, firstAddr)

		// 2. Second request should FAIL (400 Bad Request)
		manualFlowsLog.Info("Step 2: Requesting duplicate payment address (should fail)...")
		resp, err := session.InvokeCryptoPaymentCreateAddress(t)
		require.Nil(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		// Verify 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Second address creation request should be rejected")
	})

	// =========================================================================
	// FLOW-003: Concurrent Address Creation (Race Condition)
	// =========================================================================
	t.Run("FLOW-003_ConcurrentAddressRequest", func(t *testing.T) {
		email := "user2@example.com"
		session := framework.NewTestSession(proxyService.Components, email, "password123456")
		session.Register(t)
		session.Activate(t, proxyService.EmailSender)
		session.Login(t)
		manualFlowsLog.Info("Step 1: Launching concurrent address creation requests...")

		var wg sync.WaitGroup
		concurrency := 10
		responses := make([]int, concurrency)

		// Launch 10 requests simultaneously
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				// Note: InvokeCryptoPaymentCreateAddress uses t for NewRequest error checking
				// This is generally safe as NewRequest shouldn't fail with valid constant inputs
				resp, err := session.InvokeCryptoPaymentCreateAddress(t)
				if err == nil {
					responses[idx] = resp.StatusCode
					_ = resp.Body.Close()
				} else {
					manualFlowsLog.Error("Request failed", "idx", idx, "error", err)
					responses[idx] = 0
				}
			}(i)
		}
		wg.Wait()

		// Analyze results
		successCount := 0
		failCount := 0
		for _, code := range responses {
			if code == http.StatusOK {
				successCount++
			} else if code == http.StatusBadRequest || code == http.StatusConflict {
				failCount++
			}
		}

		manualFlowsLog.Info("Concurrent results", "success", successCount, "conflicts", failCount, "stats", responses)

		// We expect ONE success and (N-1) failures due to lock/duplicate check
		assert.Equal(t, 1, successCount, "Only exactly one request should succeed")
		assert.Equal(t, concurrency-1, failCount, "All other requests should fail with 400 or lock error")

		// Verify DB state consistency (User should have valid ID)
		// Check via Account endpoint
		resp, err := session.InvokeCryptoPaymentAccount(t)
		require.Nil(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should be able to retrieve account info after concurrent creation")
	})

	// =========================================================================
	// FLOW-SYNC-001 (Edge Case): Sync Logic Protection
	// Verify that DB credits are NOT lowered if Contract has less
	// =========================================================================
	t.Run("FLOW-SYNC-001_SyncDowngradeProtection", func(t *testing.T) {
		email := "user3@example.com"
		session := framework.NewTestSession(proxyService.Components, email, "password123456")
		session.Register(t)
		session.Activate(t, proxyService.EmailSender)
		session.Login(t)

		// 1. Get initial address (creates DB entry with 0 credits)
		session.ObtainDepositAddress(t)

		// 2. ARTIFICIALLY inject high credits into the Proxy DB
		// We need to bypass the API and touch the DB directly.
		keyAccess := proxyService.Components.GetSQLiteWrapper()
		user, err := keyAccess.GetUser(email)
		require.Nil(t, err)

		// Grant 5,000,000 credits manually (simulate Admin Bonus)
		currentMax := uint64(5_000_000)
		// Fix: Pass empty string for password to avoid re-hashing or length validation issues
		err = keyAccess.UpdateUser(email, "", user.IsAdmin, currentMax, user.IsPremium)
		require.Nil(t, err)

		manualFlowsLog.Info("Step 1: Manually set DB limit to 5,000,000. Contract has 0.")

		// 3. Force the Sync Job (RequestsSynchronizer) to run (or wait for it)
		time.Sleep(2 * time.Second)

		// 4. Verify DB Value hasn't been overwritten by the (0) value from contract
		updatedUser, err := keyAccess.GetUser(email)
		require.Nil(t, err)

		assert.Equal(t, currentMax, updatedUser.MaxRequests, "Database sync should NEVER lower the user's credit balance")
	})
}
