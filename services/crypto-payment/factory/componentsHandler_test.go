package factory

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

func TestNewComponentsHandler(t *testing.T) {
	t.Parallel()

	testConfig := config.Config{
		Port:                            0,
		WalletURL:                       "https://wallet",
		ExplorerURL:                     "https://explorer",
		ProxyURL:                        "proxy URL",
		ContractAddress:                 "er1test",
		CallSCGasLimit:                  100,
		SCSettingsCacheInMillis:         1,
		MinimumBalanceToProcess:         0.01,
		TimeToProcessAddressesInSeconds: 1,
		ServiceApiKey:                   "service-api-key",
	}
	relayersKeys := [][]byte{
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 32),
	}

	proxy := &testsCommon.BlockchainDataProviderStub{
		GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
			return &data.NetworkConfig{
				NumShardsWithoutMeta: 1,
			}, nil
		},
	}

	t.Run("nil proxy should error and close & start cron jobs should not panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("panic: %v", r))
			}
		}()

		dbPath := path.Join(t.TempDir(), "data.db")
		components, err := NewComponentsHandler(
			"mnemonics",
			dbPath,
			nil,
			testConfig,
			relayersKeys,
		)

		assert.Nil(t, components)
		assert.Equal(t, err, errNilBlockchainDataProvider)

		components.StartCronJobs(context.Background())
		components.Close()
	})
	t.Run("should work", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("panic: %v", r))
			}
		}()

		dbPath := path.Join(t.TempDir(), "data.db")
		components, err := NewComponentsHandler(
			"mnemonics",
			dbPath,
			proxy,
			testConfig,
			relayersKeys,
		)

		assert.NotNil(t, components)
		assert.Nil(t, err)

		components.StartCronJobs(context.Background())
		components.Close()
	})
}

func TestNewComponentsHandler_Getters(t *testing.T) {
	t.Parallel()

	testConfig := config.Config{
		Port:                            0,
		WalletURL:                       "https://wallet",
		ExplorerURL:                     "https://explorer",
		ProxyURL:                        "proxy URL",
		ContractAddress:                 "er1test",
		CallSCGasLimit:                  100,
		SCSettingsCacheInMillis:         1,
		MinimumBalanceToProcess:         0.01,
		TimeToProcessAddressesInSeconds: 1,
		ServiceApiKey:                   "service-api-key",
	}
	relayersKeys := [][]byte{
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 32),
	}

	proxy := &testsCommon.BlockchainDataProviderStub{
		GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
			return &data.NetworkConfig{
				NumShardsWithoutMeta: 1,
			}, nil
		},
	}

	dbPath := path.Join(t.TempDir(), "data.db")
	components, _ := NewComponentsHandler(
		"mnemonics",
		dbPath,
		proxy,
		testConfig,
		relayersKeys,
	)

	assert.False(t, check.IfNil(components.GetSQLiteWrapper()))
	assert.Equal(t, "*storage.sqliteWrapper", fmt.Sprintf("%T", components.GetSQLiteWrapper()))

	assert.False(t, check.IfNil(components.GetBalanceProcessor()))
	assert.Equal(t, "*process.balanceProcessor", fmt.Sprintf("%T", components.GetBalanceProcessor()))

	assert.False(t, check.IfNil(components.GetContractHandler()))
	assert.Equal(t, "*process.contractQueryHandler", fmt.Sprintf("%T", components.GetContractHandler()))
}
