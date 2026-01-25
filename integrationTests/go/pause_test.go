package _go

import (
	"context"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
	"github.com/stretchr/testify/require"
)

func TestPauseUnpause(t *testing.T) {
	if !framework.IsChainSimulatorIsRunning() {
		t.Skip("No chain simulator instance running found. Skipping slow test")
	}
	cryptoService := framework.NewCryptoPaymentService(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cryptoService.Setup(ctx)
	defer cryptoService.TearDown()

	cryptoService.CreateService()
	balanceProcessor := cryptoService.Components.GetBalanceProcessor()

	log.Info("======== 1. Check that the contract is not paused after deployment")
	checkIsPaused(ctx, cryptoService, false)
	log.Info("Done ✓")

	log.Info("======== 2. Pausing contract")
	_, _, txStatus := cryptoService.ChainSimulator.ScCall(
		ctx,
		cryptoService.Keys.OwnerKeys.MvxSk,
		cryptoService.ContractAddress,
		"0",
		framework.CallGasLimit,
		"pause",
		make([]string, 0),
	)
	require.Equal(t, "success", string(txStatus))
	log.Info("Done ✓")

	log.Info("======== 3. Check that the contract is paused")
	checkIsPaused(ctx, cryptoService, true)
	log.Info("Done ✓")

	log.Info("======== 4. All credit operations should not happen")
	txHash1 := cryptoService.ChainSimulator.SendTxWithoutGenerateBlocks(
		ctx,
		cryptoService.Keys.UserAKeys.MvxSk,
		cryptoService.Keys.UserAKeys.PayAddress,
		"1010000000000000000", // 1.01 EGLD
		framework.PaymentGasLimit,
		make([]byte, 0),
	)
	log.Info("sent payment tx from user A", "txHash", txHash1)
	cryptoService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, txHash1)

	log.Info("   ===== 4a. Check payment addresses to have EGLD")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "1010000000000000000") // 1.01 EGLD
	log.Info("   Done ✓")

	log.Info("   ===== 4b. Trying to process payments")
	err := balanceProcessor.ProcessAll(ctx)
	require.NotNil(t, err)
	log.Error("Error processing payments", "error", err)
	log.Info("   Done ✓")

	log.Info("   ===== 4c. Generate blocks until the payments are completed")
	cryptoService.ChainSimulator.GenerateBlocks(ctx, 12)
	log.Info("   Done ✓")

	log.Info("   ===== 4d. Check payment addresses to have the same amount of EGLD")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "1010000000000000000") // 1.01 EGLD
	log.Info("   Done ✓")

	log.Info("   ===== 4e. Credits should not change")
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserAKeys.ID, 0)
	log.Info("   Done ✓")
	log.Info("Done ✓")

	log.Info("======== 5. Unpausing contract")
	_, _, txStatus = cryptoService.ChainSimulator.ScCall(
		ctx,
		cryptoService.Keys.OwnerKeys.MvxSk,
		cryptoService.ContractAddress,
		"0",
		framework.CallGasLimit,
		"unpause",
		make([]string, 0),
	)
	require.Equal(t, "success", string(txStatus))
	log.Info("Done ✓")

	log.Info("======== 6. Check that the contract is unpaused")
	checkIsPaused(ctx, cryptoService, false)
	log.Info("Done ✓")

	log.Info("======== 7. All credit operations should happen")
	log.Info("   ===== 7a. Check payment addresses to have EGLD")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "1010000000000000000") // 1.01 EGLD
	log.Info("   Done ✓")

	log.Info("   ===== 7b. Trying to process payments")
	err = balanceProcessor.ProcessAll(ctx)
	require.Nil(t, err)
	log.Info("   Done ✓")

	log.Info("   ===== 7c. Generate blocks until the payments are completed")
	cryptoService.ChainSimulator.GenerateBlocks(ctx, 12)
	log.Info("   Done ✓")

	log.Info("   ===== 7d. Check payment addresses to be empty")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "0")
	log.Info("   Done ✓")

	log.Info("   ===== 4e. Credits should change")
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserAKeys.ID, 505000) // 500000 * 1.01
	log.Info("   Done ✓")
	log.Info("Done ✓")
}

func checkIsPaused(ctx context.Context, service *framework.CryptoPaymentService, expected bool) {
	contractHandler := service.Components.GetContractHandler()
	isPaused, err := contractHandler.IsContractPaused(ctx)
	require.Nil(service, err)

	require.Equal(service, expected, isPaused)
}
