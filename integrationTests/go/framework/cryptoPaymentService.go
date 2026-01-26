package framework

import (
	"context"
	"encoding/hex"
	"math/big"
	"path"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/factory"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/stretchr/testify/require"
)

const deployGasLimit = 20_000_000
const callGasLimit = 40_000_000
const minimumBalanceToCall = 0.01 // 0.01 EGLD

// CryptoPaymentService will hold all elements used by the crypto-payment service
type CryptoPaymentService struct {
	testing.TB
	Keys           *KeysStore
	ChainSimulator *chainSimulatorWrapper

	ContractAddress *MvxAddress
	Components      CryptoPaymentComponentsHandler
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
	cfg, err := chainSimulator.Proxy().GetNetworkConfig(context.Background())
	require.Nil(tb, err)

	return &CryptoPaymentService{
		TB:             tb,
		Keys:           NewKeysStore(tb, cfg.NumShardsWithoutMeta),
		ChainSimulator: chainSimulator,
	}
}

// Setup prepares the environment
func (crs *CryptoPaymentService) Setup(ctx context.Context, numRequestsPerEGLD int64) {
	log.Info("minting tokens to the users")
	crs.ChainSimulator.FundWallets(ctx, crs.Keys.WalletsToFundOnMultiversX())

	log.Info("deploying contract")
	address, _, txOnNetwork := crs.ChainSimulator.DeploySC(
		ctx,
		GetContractPath("requests"),
		crs.Keys.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{
			hex.EncodeToString(big.NewInt(numRequestsPerEGLD).Bytes()),
		},
	)

	log.Info("deployed the requests contract", "address", address.Bech32(), "txHash", txOnNetwork.Hash)
	crs.ContractAddress = address
}

// TearDown cleans up the test environment
func (crs *CryptoPaymentService) TearDown() {
	crs.Components.Close()
}

// CreateService will assemble all the service processing components
func (crs *CryptoPaymentService) CreateService() {
	w := interactors.NewWallet()
	mnemonic, err := w.GenerateMnemonic()
	require.Nil(crs, err)

	log.Info("generated mnemonic", "mnemonic", mnemonic)

	cfg := config.Config{
		Port:                            0,
		WalletURL:                       "https://wallet",
		ExplorerURL:                     "https://explorer",
		ProxyURL:                        "",
		ContractAddress:                 crs.ContractAddress.Bech32(),
		CallSCGasLimit:                  callGasLimit,
		SCSettingsCacheInMillis:         1,
		MinimumBalanceToProcess:         minimumBalanceToCall,
		TimeToProcessAddressesInSeconds: 5,
		ServiceApiKey:                   "service-api-key",
	}

	dbPath := path.Join(crs.TempDir(), "test.db")

	relayersKeys := make([][]byte, 0, len(crs.Keys.RelayersKeys))
	for _, relayerKey := range crs.Keys.RelayersKeys {
		relayersKeys = append(relayersKeys, relayerKey.MvxSk)
	}

	crs.Components, err = factory.NewComponentsHandler(
		string(mnemonic),
		dbPath,
		crs.ChainSimulator.Proxy(),
		cfg,
		relayersKeys,
	)
	require.Nil(crs, err)

	// Add users to DB
	var bech32Address string

	sqliteWrapper := crs.Components.GetSQLiteWrapper()
	crs.Keys.UserAKeys.ID, err = sqliteWrapper.Add()
	require.Nil(crs, err)
	entryA, err := sqliteWrapper.Get(crs.Keys.UserAKeys.ID)
	require.Nil(crs, err)
	bech32Address = entryA.Address
	crs.Keys.UserAKeys.PayAddress = NewMvxAddressFromBech32(crs, bech32Address)
	log.Info("registered user A", "UserA address", crs.Keys.UserAKeys.MvxAddress.Bech32(),
		"payment address", bech32Address, "contract ID", crs.Keys.UserAKeys.ID)

	crs.Keys.UserBKeys.ID, err = sqliteWrapper.Add()
	require.Nil(crs, err)
	entryB, err := sqliteWrapper.Get(crs.Keys.UserBKeys.ID)
	require.Nil(crs, err)
	bech32Address = entryB.Address
	crs.Keys.UserBKeys.PayAddress = NewMvxAddressFromBech32(crs, bech32Address)
	log.Info("registered user B", "UserB address", crs.Keys.UserBKeys.MvxAddress.Bech32(),
		"payment address", bech32Address, "contract ID", crs.Keys.UserBKeys.ID)

	crs.Keys.UserCKeys.ID, err = sqliteWrapper.Add()
	require.Nil(crs, err)
	entryC, err := sqliteWrapper.Get(crs.Keys.UserCKeys.ID)
	require.Nil(crs, err)
	bech32Address = entryC.Address
	crs.Keys.UserCKeys.PayAddress = NewMvxAddressFromBech32(crs, bech32Address)
	log.Info("registered user C", "UserC address", crs.Keys.UserCKeys.MvxAddress.Bech32(),
		"payment address", bech32Address, "contract ID", crs.Keys.UserCKeys.ID)
}
