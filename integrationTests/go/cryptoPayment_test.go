package _go

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const functionGetRequests = "getRequests"

var log = logger.GetOrCreate("integrationTests")

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

func TestCallingSCWhenBalanceIsAvailableInSync(t *testing.T) {
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

	log.Info("======== 1. All users initiate some payments")
	txHash1 := cryptoService.ChainSimulator.SendTxWithoutGenerateBlocks(
		ctx,
		cryptoService.Keys.UserAKeys.MvxSk,
		cryptoService.Keys.UserAKeys.PayAddress,
		"1010000000000000000", // 1.01 EGLD
		framework.PaymentGasLimit,
		make([]byte, 0),
	)
	log.Info("sent payment tx from user A", "txHash", txHash1)

	txHash2 := cryptoService.ChainSimulator.SendTxWithoutGenerateBlocks(
		ctx,
		cryptoService.Keys.UserBKeys.MvxSk,
		cryptoService.Keys.UserBKeys.PayAddress,
		"1999999900000000000", // 1.9999999 EGLD
		framework.PaymentGasLimit,
		make([]byte, 0),
	)
	log.Info("sent payment tx from user B", "txHash", txHash2)

	txHash3 := cryptoService.ChainSimulator.SendTxWithoutGenerateBlocks(
		ctx,
		cryptoService.Keys.UserCKeys.MvxSk,
		cryptoService.Keys.UserCKeys.PayAddress,
		"3000000000000000000", // 3 EGLD
		framework.PaymentGasLimit,
		make([]byte, 0),
	)
	log.Info("sent payment tx from user C", "txHash", txHash3)

	log.Info("======== 2. The payments are not completed yet, all contract credits should be 0")
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserAKeys.ID, 0)
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserBKeys.ID, 0)
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserCKeys.ID, 0)
	log.Info("Done ✓")

	log.Info("======== 3. Generate blocks until the payments are completed")
	cryptoService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, txHash1)
	cryptoService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, txHash2)
	cryptoService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, txHash3)
	log.Info("Done ✓")

	log.Info("======== 4. Check payment addresses to have EGLD")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "1010000000000000000") // 1.01 EGLD
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserBKeys.PayAddress, "1999999900000000000") // 1.9999999 EGLD
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserCKeys.PayAddress, "3000000000000000000") // 3 EGLD
	log.Info("Done ✓")

	log.Info("======== 5. The SC call is not done, all contract credits should be 0")
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserAKeys.ID, 0)
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserBKeys.ID, 0)
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserCKeys.ID, 0)
	log.Info("Done ✓")

	log.Info("======== 6. The balance processor checks & process all addresses")
	err := balanceProcessor.ProcessAll(ctx)
	require.Nil(t, err)

	log.Info("======== 7. Generate blocks until the payments are completed")
	cryptoService.ChainSimulator.GenerateBlocks(ctx, 12)
	log.Info("Done ✓")

	log.Info("======== 8. Check payment addresses to have 0 EGLD")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "0")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserBKeys.PayAddress, "0")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserCKeys.PayAddress, "0")
	log.Info("Done ✓")

	log.Info("======== 9. The payments are completed yet, check credits")
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserAKeys.ID, 505000)  // 500000 * 1.01
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserBKeys.ID, 999999)  // 500000 * 1.9999999
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserCKeys.ID, 1500000) // 500000 * 3
	log.Info("Done ✓")

	log.Info("======== 10. Another round of payments")
	txHash1 = cryptoService.ChainSimulator.SendTxWithoutGenerateBlocks(
		ctx,
		cryptoService.Keys.UserAKeys.MvxSk,
		cryptoService.Keys.UserAKeys.PayAddress,
		"950000000000000000", // 0.95 EGLD
		framework.PaymentGasLimit,
		make([]byte, 0),
	)
	log.Info("sent payment tx from user A", "txHash", txHash1)

	txHash2 = cryptoService.ChainSimulator.SendTxWithoutGenerateBlocks(
		ctx,
		cryptoService.Keys.UserBKeys.MvxSk,
		cryptoService.Keys.UserBKeys.PayAddress,
		"009999999000000000", // 0.009999999 EGLD
		framework.PaymentGasLimit,
		make([]byte, 0),
	)
	log.Info("sent payment tx from user B", "txHash", txHash2)

	txHash3 = cryptoService.ChainSimulator.SendTxWithoutGenerateBlocks(
		ctx,
		cryptoService.Keys.UserCKeys.MvxSk,
		cryptoService.Keys.UserCKeys.PayAddress,
		"10000000000000000", // 0.01 EGLD
		framework.PaymentGasLimit,
		make([]byte, 0),
	)
	log.Info("sent payment tx from user C", "txHash", txHash3)

	log.Info("======== 11. Generate blocks until the payments are completed")
	cryptoService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, txHash1)
	cryptoService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, txHash2)
	cryptoService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, txHash3)
	log.Info("Done ✓")

	log.Info("======== 12. Check payment addresses to have EGLD")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "950000000000000000") // 0.95 EGLD
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserBKeys.PayAddress, "9999999000000000")   // 0.009999999 EGLD
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserCKeys.PayAddress, "10000000000000000")  // 0.01 EGLD
	log.Info("Done ✓")

	log.Info("======== 13. The balance processor checks & process all addresses")
	err = balanceProcessor.ProcessAll(ctx)
	require.Nil(t, err)

	log.Info("======== 14. Generate blocks until the payments are completed")
	cryptoService.ChainSimulator.GenerateBlocks(ctx, 12)
	log.Info("Done ✓")

	log.Info("======== 15. Check payment addresses to have 0 EGLD except user B's address")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserAKeys.PayAddress, "0")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserBKeys.PayAddress, "9999999000000000")
	checkEGLD(ctx, cryptoService, cryptoService.Keys.UserCKeys.PayAddress, "0")
	log.Info("Done ✓")

	log.Info("======== 16. The payments are completed yet, check credits")
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserAKeys.ID, 980000)  // 500000 * 1.01 + 500000 * 0.95
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserBKeys.ID, 999999)  // 500000 * 1.9999999 + 0
	checkCredits(ctx, cryptoService, cryptoService.Keys.UserCKeys.ID, 1505000) // 500000 * 3 + 500000 * 0.01
	log.Info("Done ✓")
}

func checkIsPaused(ctx context.Context, service *framework.CryptoPaymentService, expected bool) {
	contractHandler := service.Components.GetContractHandler()
	isPaused, err := contractHandler.IsContractPaused(ctx)
	require.Nil(service, err)

	require.Equal(service, expected, isPaused)
}

func checkCredits(ctx context.Context, service *framework.CryptoPaymentService, id uint64, expectedCredits uint64) {
	result := service.ChainSimulator.ExecuteVMQuery(
		ctx,
		service.ContractAddress,
		functionGetRequests,
		[]string{
			hex.EncodeToString(big.NewInt(int64(id)).Bytes()),
		},
	)
	require.NotNil(service, result)
	require.Equal(service, 1, len(result))

	valueAsBytes := result[0]
	credits := big.NewInt(0).SetBytes(valueAsBytes)
	require.Equal(service, expectedCredits, credits.Uint64())
}

func checkEGLD(ctx context.Context, service *framework.CryptoPaymentService, address *framework.MvxAddress, expectedEGLD string) {
	account, err := service.ChainSimulator.Proxy().GetAccount(ctx, address)
	require.Nil(service, err)

	require.Equal(service, expectedEGLD, account.Balance)
}
