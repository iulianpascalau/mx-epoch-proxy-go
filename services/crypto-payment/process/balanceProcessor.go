package process

import (
	"context"
	"fmt"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("process")

const egldDecimals = 18

type balanceProcessor struct {
	dataProvider           DataProvider
	blockchainDataProvider BlockchainDataProvider
	balanceOperator        BalanceOperator
	minimumBalanceToCall   float64
}

// NewBalanceProcessor creates a new instance of balanceProcessor
func NewBalanceProcessor(
	dataProvider DataProvider,
	blockchainDataProvider BlockchainDataProvider,
	balanceOperator BalanceOperator,
	minimumBalanceToCall float64,
) (*balanceProcessor, error) {
	if check.IfNil(dataProvider) {
		return nil, errNilDataProvider
	}
	if check.IfNil(blockchainDataProvider) {
		return nil, errNilBlockchainDataProvider
	}
	if check.IfNil(balanceOperator) {
		return nil, errNilBalanceOperator
	}
	if minimumBalanceToCall <= 0 {
		return nil, errInvalidMinimumBalanceToCall
	}

	return &balanceProcessor{
		dataProvider:           dataProvider,
		blockchainDataProvider: blockchainDataProvider,
		balanceOperator:        balanceOperator,
		minimumBalanceToCall:   minimumBalanceToCall,
	}, nil
}

// Process will update the inner data provider state based on the accounts balances changes
func (processor *balanceProcessor) Process(ctx context.Context) error {
	allRows, err := processor.dataProvider.GetAll()
	if err != nil {
		return fmt.Errorf("%w when getting all records", err)
	}

	for _, row := range allRows {
		processor.processRecord(ctx, row)
	}
	return nil
}

func (processor *balanceProcessor) processRecord(ctx context.Context, row *common.BalanceEntry) {
	select {
	case <-ctx.Done():
		log.Debug("context done", "id", row.ID, "address", row.Address)
		return
	default:
	}

	addressHandler, err := data.NewAddressFromBech32String(row.Address)
	if err != nil {
		log.Trace("error converting address to AddressHandler instance", "id", row.ID, "address", row.Address, "error", err)
		return
	}

	accountData, err := processor.blockchainDataProvider.GetAccount(ctx, addressHandler)
	if err != nil {
		log.Debug("error fetching account data", "id", row.ID, "address", row.Address, "error", err)
		return
	}

	blockchainBalance, err := accountData.GetBalance(egldDecimals)
	if err != nil {
		log.Debug("error getting the balance", "id", row.ID, "address", row.Address, "blockchain balance", accountData.Balance, "error", err)
		return
	}

	if blockchainBalance < processor.minimumBalanceToCall {
		log.Trace("balance is too low", "id", row.ID, "address", row.Address, "blockchain balance", accountData.Balance)
		return
	}

	err = processor.balanceOperator.Process(ctx, row.ID, row.Address, accountData.Balance, accountData.Nonce)
	if err != nil {
		log.Error("error processing balance",
			"id", row.ID, "address", row.Address, "balance", accountData.Balance,
			"nonce", accountData.Nonce, "error", err)
	}
}

// IsInterfaceNil returns true if the value under the interface is nil
func (processor *balanceProcessor) IsInterfaceNil() bool {
	return processor == nil
}
