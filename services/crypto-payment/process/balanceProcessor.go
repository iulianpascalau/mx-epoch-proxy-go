package process

import (
	"context"
	"fmt"
	"math"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("process")

const egldDecimals = 18
const minimumBalance = float64(0.01)

type balanceProcessor struct {
	dataProvider           DataProvider
	blockchainDataProvider BlockchainDataProvider
	numRequestsPerEgldUnit int
}

// NewBalanceProcessor creates a new instance of balanceProcessor
func NewBalanceProcessor(
	dataProvider DataProvider,
	blockchainDataProvider BlockchainDataProvider,
	numRequestsPerEgldUnit int,
) (*balanceProcessor, error) {
	if dataProvider == nil {
		return nil, errNilDataProvider
	}
	if blockchainDataProvider == nil {
		return nil, errNilBlockchainDataProvider
	}
	if numRequestsPerEgldUnit <= 0 {
		return nil, errInvalidNumRequestsPerUnit
	}

	return &balanceProcessor{
		dataProvider:           dataProvider,
		blockchainDataProvider: blockchainDataProvider,
		numRequestsPerEgldUnit: numRequestsPerEgldUnit,
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

	err = processor.updateBalance(row, blockchainBalance)
	if err != nil {
		log.Warn("error updating balance", "id", row.ID, "address", row.Address, "blockchain balance", accountData.Balance, "error", err)
	}
}

func (processor *balanceProcessor) updateBalance(row *common.BalanceEntry, blockchainBalance float64) error {
	if row.CurrentBalance == blockchainBalance {
		// nothing's changed, return
		return nil
	}

	amount := float64(0)
	if blockchainBalance > row.CurrentBalance {
		amount = blockchainBalance - row.CurrentBalance
	}

	numRequests := processor.computeNumRequestsToAdd(row, amount)

	return processor.dataProvider.UpdateBalance(row.ID, blockchainBalance, row.TotalRequests+numRequests)
}

func (processor *balanceProcessor) computeNumRequestsToAdd(row *common.BalanceEntry, amount float64) int {
	if amount < minimumBalance {
		log.Debug("amount is too small", "id", row.ID, "address", row.Address, "amount", amount)
		return 0
	}

	return int(math.Ceil(float64(processor.numRequestsPerEgldUnit) * amount))
}

// IsInterfaceNil returns true if the value under the interface is nil
func (processor *balanceProcessor) IsInterfaceNil() bool {
	return processor == nil
}
