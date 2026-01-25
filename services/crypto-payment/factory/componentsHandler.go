package factory

import (
	"context"
	"fmt"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/crypto"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/storage"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/interactors"
)

var log = logger.GetOrCreate("factory")

type componentsHandler struct {
	config              config.Config
	wallet              crypto.Wallet
	multipleKeysHandler process.MultipleKeysHandler
	sqliteWrapper       SQLiteWrapper
	proxy               process.BlockchainDataProvider
	timeCacher          process.Cacher
	contractHandler     process.ContractHandler
	configProvider      api.ConfigProvider
	accountHandler      api.AccountHandler
	apiHandler          APIHandler
	httpServer          HTTPServer
	relayersHandlers    []process.SingleKeyHandler
	balanceOperator     process.BalanceOperator
	balanceProcessor    BalanceProcessor
}

// NewComponentsHandler creates a new instance of the components handler holding all high-level components
func NewComponentsHandler(
	mnemonics string,
	sqlitePath string,
	proxy process.BlockchainDataProvider,
	cfg config.Config,
	relayersKeys [][]byte,
) (*componentsHandler, error) {
	if check.IfNil(proxy) {
		return nil, errNilBlockchainDataProvider
	}

	ch := &componentsHandler{
		config: cfg,
		proxy:  proxy,
	}
	var err error
	defer func() {
		if err != nil {
			ch.Close()
		}
	}()

	ch.wallet = interactors.NewWallet()
	ch.multipleKeysHandler, err = crypto.NewMultipleKeysHandler(ch.wallet, mnemonics)
	if err != nil {
		return nil, err
	}

	ch.sqliteWrapper, err = storage.NewSQLiteWrapper(sqlitePath, ch.multipleKeysHandler)
	if err != nil {
		return nil, err
	}

	ch.timeCacher = storage.NewTimeCacher(time.Duration(cfg.SCSettingsCacheInMillis) * time.Millisecond)
	ch.contractHandler, err = process.NewContractQueryHandler(
		ch.proxy,
		cfg.ContractAddress,
		ch.timeCacher,
	)
	if err != nil {
		return nil, err
	}

	ch.configProvider, err = process.NewConfigHandler(
		cfg.WalletURL,
		cfg.ExplorerURL,
		ch.contractHandler,
	)
	if err != nil {
		return nil, err
	}

	ch.accountHandler, err = process.NewAccountHandler(ch.contractHandler, ch.sqliteWrapper)
	if err != nil {
		return nil, err
	}

	ch.apiHandler, err = api.NewHandler(ch.sqliteWrapper, ch.configProvider, ch.accountHandler)
	if err != nil {
		return nil, err
	}

	ch.httpServer = api.NewHTTPServer(ch.apiHandler, int(cfg.Port), cfg.ServiceApiKey)
	err = ch.httpServer.Start()
	if err != nil {
		return nil, err
	}

	ch.relayersHandlers = make([]process.SingleKeyHandler, 0, len(relayersKeys))
	for _, relayerKey := range relayersKeys {
		relayerHandler, errCreate := crypto.NewSingleKeyHandler(relayerKey)
		if errCreate != nil {
			return nil, errCreate
		}
		ch.relayersHandlers = append(ch.relayersHandlers, relayerHandler)
	}

	ch.balanceOperator, err = process.NewRelayedTxProcessor(
		proxy,
		ch.multipleKeysHandler,
		ch.relayersHandlers,
		cfg.CallSCGasLimit,
		cfg.ContractAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("%w while initializing the relayedTxProcessor", err)
	}

	ch.balanceProcessor, err = process.NewBalanceProcessor(
		ch.sqliteWrapper,
		ch.proxy,
		ch.balanceOperator,
		ch.contractHandler,
		cfg.MinimumBalanceToProcess,
	)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// StartCronJobs starts all defined cron jobs
func (ch *componentsHandler) StartCronJobs(ctx context.Context) {
	if ch == nil {
		return
	}

	common.CronJobStarter(ctx, func() {
		errRun := ch.balanceProcessor.ProcessAll(ctx)
		log.LogIfError(errRun)
	}, time.Duration(ch.config.TimeToProcessAddressesInSeconds)*time.Second)
}

// GetSQLiteWrapper returns the SQLiteWrapper instance
func (ch *componentsHandler) GetSQLiteWrapper() SQLiteWrapper {
	return ch.sqliteWrapper
}

// GetBalanceProcessor returns the BalanceProcessor instance
func (ch *componentsHandler) GetBalanceProcessor() BalanceProcessor {
	return ch.balanceProcessor
}

// GetContractHandler returns the ContractHandler instance
func (ch *componentsHandler) GetContractHandler() process.ContractHandler {
	return ch.contractHandler
}

// Close closes all the components held by the handler
func (ch *componentsHandler) Close() {
	if ch == nil {
		return
	}

	if ch.sqliteWrapper != nil {
		err := ch.sqliteWrapper.Close()
		log.LogIfError(err)
	}

	if ch.timeCacher != nil {
		ch.timeCacher.Close()
	}

	if ch.httpServer != nil {
		_ = ch.httpServer.Close()
	}

	if ch.balanceOperator != nil {
		_ = ch.balanceOperator.Close()
	}
}
