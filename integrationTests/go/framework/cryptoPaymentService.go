package framework

import (
	"context"
	"encoding/hex"
	"math/big"
	"path"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/crypto"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/storage"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/stretchr/testify/require"
)

const deployGasLimit = 20_000_000
const callGasLimit = 40_000_000
const requestsPerEGLD = 500_000
const minimumBalanceToCall = 0.01 // 0.01 EGLD

// CryptoPaymentService will hold all elements used by the crypto-payment service
type CryptoPaymentService struct {
	testing.TB
	Keys           *KeysStore
	ChainSimulator *chainSimulatorWrapper

	ContractAddress  *MvxAddress
	SQLiteWrapper    SQLiteWrapper
	BalanceProcessor BalanceProcessor
}

// NewCryptoPaymentService creates a new CryptoPaymentService instance
func NewCryptoPaymentService(tb testing.TB) *CryptoPaymentService {
	args := ArgChainSimulatorWrapper{
		TB:                           tb,
		ProxyCacherExpirationSeconds: 600,
		ProxyMaxNoncesDelta:          7,
	}
	chainSimulator := CreateChainSimulatorWrapper(args)
	chainSimulator.GenerateBlocksUntilEpochReached(context.Background(), 1)
	config, err := chainSimulator.Proxy().GetNetworkConfig(context.Background())
	require.Nil(tb, err)

	return &CryptoPaymentService{
		TB:             tb,
		Keys:           NewKeysStore(tb, config.NumShardsWithoutMeta),
		ChainSimulator: chainSimulator,
	}
}

// Setup prepares the environment
func (crs *CryptoPaymentService) Setup(ctx context.Context) {
	log.Info("minting tokens to the users")
	crs.ChainSimulator.FundWallets(ctx, crs.Keys.WalletsToFundOnMultiversX())

	log.Info("deploying contract")
	address, _, txOnNetwork := crs.ChainSimulator.DeploySC(
		ctx,
		GetContractPath("requests"),
		crs.Keys.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{
			hex.EncodeToString(big.NewInt(requestsPerEGLD).Bytes()),
		},
	)

	log.Info("deployed the requests contract", "address", address.Bech32(), "txHash", txOnNetwork.Hash)
	crs.ContractAddress = address
}

// TearDown cleans up the test environment
func (crs *CryptoPaymentService) TearDown() {
	_ = crs.SQLiteWrapper.Close()
}

// CreateService will assemble all the service processing components
func (crs *CryptoPaymentService) CreateService() {
	w := interactors.NewWallet()
	mnemonic, err := w.GenerateMnemonic()
	require.Nil(crs, err)

	log.Info("generated mnemonic", "mnemonic", mnemonic)

	multipleAddressHandler, err := crypto.NewMultipleKeysHandler(interactors.NewWallet(), string(mnemonic))
	require.Nil(crs, err)

	dbPath := path.Join(crs.TempDir(), "test.db")
	crs.SQLiteWrapper, err = storage.NewSQLiteWrapper(dbPath, multipleAddressHandler)
	require.Nil(crs, err)

	relayersHandlers := make([]process.SingleKeyHandler, 0, 100)
	for _, relayerKey := range crs.Keys.RelayersKeys {
		relayerHandler, errCreate := crypto.NewSingleKeyHandler(relayerKey.MvxSk)
		require.Nil(crs, errCreate)
		relayersHandlers = append(relayersHandlers, relayerHandler)
	}

	// Add users to DB
	var bech32Address string

	crs.Keys.UserAKeys.ID, bech32Address, err = crs.SQLiteWrapper.Add()
	require.Nil(crs, err)
	crs.Keys.UserAKeys.PayAddress = NewMvxAddressFromBech32(crs, bech32Address)
	log.Info("registered user A", "UserA address", crs.Keys.UserAKeys.MvxAddress.Bech32(),
		"payment address", bech32Address, "contract ID", crs.Keys.UserAKeys.ID)

	crs.Keys.UserBKeys.ID, bech32Address, err = crs.SQLiteWrapper.Add()
	require.Nil(crs, err)
	crs.Keys.UserBKeys.PayAddress = NewMvxAddressFromBech32(crs, bech32Address)
	log.Info("registered user B", "UserB address", crs.Keys.UserBKeys.MvxAddress.Bech32(),
		"payment address", bech32Address, "contract ID", crs.Keys.UserBKeys.ID)

	crs.Keys.UserCKeys.ID, bech32Address, err = crs.SQLiteWrapper.Add()
	require.Nil(crs, err)
	crs.Keys.UserCKeys.PayAddress = NewMvxAddressFromBech32(crs, bech32Address)
	log.Info("registered user C", "UserC address", crs.Keys.UserCKeys.MvxAddress.Bech32(),
		"payment address", bech32Address, "contract ID", crs.Keys.UserCKeys.ID)

	relayedTxProcessor, err := process.NewRelayedTxProcessor(
		crs.ChainSimulator.Proxy(),
		multipleAddressHandler,
		relayersHandlers,
		callGasLimit,
		crs.ContractAddress.Bech32(),
	)
	require.Nil(crs, err)

	crs.BalanceProcessor, err = process.NewBalanceProcessor(
		crs.SQLiteWrapper,
		crs.ChainSimulator.Proxy(),
		relayedTxProcessor,
		minimumBalanceToCall,
	)
	require.Nil(crs, err)
}
