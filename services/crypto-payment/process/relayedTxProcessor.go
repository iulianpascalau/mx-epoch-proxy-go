package process

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-sdk-go/builders"
)

const requestsAddEndpoint = "addRequests"

var hashSigningTxHasher = keccak.NewKeccak()

type relayedTxProcessor struct {
	blockchainDataProvider BlockchainDataProvider
	userKeys               MultipleKeysHandler
	relayerKey             SingleKeyHandler
	gasLimit               uint64
	contractBech32Address  string
}

// NewRelayedTxProcessor creates a new instance of relayedTxProcessor
func NewRelayedTxProcessor(
	blockchainDataProvider BlockchainDataProvider,
	userKeys MultipleKeysHandler,
	relayerKey SingleKeyHandler,
	gasLimit uint64,
	contractBech32Address string,
) (*relayedTxProcessor, error) {
	if check.IfNil(blockchainDataProvider) {
		return nil, errNilBlockchainDataProvider
	}
	if check.IfNil(userKeys) {
		return nil, errNilUserKeysHandler
	}
	if check.IfNil(relayerKey) {
		return nil, errNilRelayerKeysHandler
	}
	if gasLimit == 0 {
		return nil, errZeroGasLimit
	}
	if len(contractBech32Address) == 0 {
		return nil, errEmptyContractBech32Address
	}

	return &relayedTxProcessor{
		blockchainDataProvider: blockchainDataProvider,
		userKeys:               userKeys,
		relayerKey:             relayerKey,
		gasLimit:               gasLimit,
		contractBech32Address:  contractBech32Address,
	}, nil
}

// Process implements BalanceOperator
func (processor *relayedTxProcessor) Process(ctx context.Context, id uint64, bech32Address string, value string, nonce uint64) error {
	networkConfig, err := processor.blockchainDataProvider.GetNetworkConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get network config: %w", err)
	}

	// 1. Prepare the data field
	dataFieldBuilder := builders.NewTxDataBuilder()
	dataField, err := dataFieldBuilder.Function(requestsAddEndpoint).ArgInt64(int64(id)).ToDataBytes()
	if err != nil {
		return fmt.Errorf("failed to build data field: %w", err)
	}

	// 2. Assemble the Frontend transaction
	tx := &transaction.FrontendTransaction{
		Nonce:    nonce,
		Value:    value,
		Receiver: processor.contractBech32Address,
		Sender:   bech32Address,
		GasPrice: networkConfig.MinGasPrice,
		GasLimit: processor.gasLimit,
		Data:     dataField,
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
	}

	// 3. Sign the frontend transaction with the key at the provided index
	unsignedtTxBytes, err := generateTransactionBytesToSign(tx)
	if err != nil {
		return fmt.Errorf("failed to generate unsigned tx bytes: %w", err)
	}

	userSig, err := processor.userKeys.Sign(uint32(id), unsignedtTxBytes)
	if err != nil {
		return fmt.Errorf("failed to sign the transaction with user key: %w", err)
	}
	tx.Signature = hex.EncodeToString(userSig)

	// 4. Sign the frontend transaction with the singleKeyHandler and populate relayer specific fields
	relayerSig, err := processor.relayerKey.Sign(unsignedtTxBytes)
	if err != nil {
		return fmt.Errorf("failed to sign the transaction with relayer key: %w", err)
	}
	tx.RelayerSignature = hex.EncodeToString(relayerSig)
	tx.RelayerAddr = processor.relayerKey.GetBech32Address()

	// 5. Send transaction
	hash, err := processor.blockchainDataProvider.SendTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Info("Transaction sent",
		"sender", bech32Address,
		"nonce", nonce,
		"value", value,
		"data field", string(dataField),
		"hash", hash)
	return nil
}

func generateTransactionBytesToSign(tx *transaction.FrontendTransaction) ([]byte, error) {
	txToMarshal := builders.TransactionToUnsignedTx(tx)
	unsignedMessage, err := json.Marshal(txToMarshal)
	if err != nil {
		return nil, err
	}

	shouldSignOnTxHash := txToMarshal.Version >= 2 && txToMarshal.Options&1 > 0
	if shouldSignOnTxHash {
		unsignedMessage = hashSigningTxHasher.Compute(string(unsignedMessage))
	}

	return unsignedMessage, nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (processor *relayedTxProcessor) IsInterfaceNil() bool {
	return processor == nil
}
