package process

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/interactors/nonceHandlerV2"
)

const requestsAddEndpoint = "addRequests"
const intervalToResendTxs = time.Minute

var hashSigningTxHasher = keccak.NewKeccak()

type relayedTxProcessor struct {
	blockchainDataProvider BlockchainDataProvider
	userKeys               MultipleKeysHandler
	relayersKeys           map[uint32]SingleKeyHandler
	gasLimit               uint64
	contractBech32Address  string
	nonceTxHandler         NonceTransactionsHandler
}

// NewRelayedTxProcessor creates a new instance of relayedTxProcessor
func NewRelayedTxProcessor(
	blockchainDataProvider BlockchainDataProvider,
	userKeys MultipleKeysHandler,
	relayersKeys []SingleKeyHandler,
	gasLimit uint64,
	contractBech32Address string,
) (*relayedTxProcessor, error) {
	if check.IfNil(blockchainDataProvider) {
		return nil, errNilBlockchainDataProvider
	}
	if check.IfNil(userKeys) {
		return nil, errNilUserKeysHandler
	}

	relayersMap, err := makeRelayersMap(relayersKeys, blockchainDataProvider)
	if err != nil {
		return nil, err
	}

	if gasLimit == 0 {
		return nil, errZeroGasLimit
	}
	if len(contractBech32Address) == 0 {
		return nil, errEmptyContractBech32Address
	}

	argsNonceTxHandler := nonceHandlerV2.ArgsNonceTransactionsHandlerV2{
		Proxy:            blockchainDataProvider,
		IntervalToResend: intervalToResendTxs,
	}

	nonceTxHandler, err := nonceHandlerV2.NewNonceTransactionHandlerV2(argsNonceTxHandler)
	if err != nil {
		return nil, err
	}

	return &relayedTxProcessor{
		blockchainDataProvider: blockchainDataProvider,
		userKeys:               userKeys,
		relayersKeys:           relayersMap,
		gasLimit:               gasLimit,
		contractBech32Address:  contractBech32Address,
		nonceTxHandler:         nonceTxHandler,
	}, nil
}

func makeRelayersMap(relayersKeys []SingleKeyHandler, blockchainDataProvider BlockchainDataProvider) (map[uint32]SingleKeyHandler, error) {
	if relayersKeys == nil {
		return nil, errNilRelayersKeysMap
	}

	networkConfig, err := blockchainDataProvider.GetNetworkConfig(context.Background())
	if err != nil {
		return nil, err
	}

	shardCoordinator, err := blockchain.NewShardCoordinator(networkConfig.NumShardsWithoutMeta, 0)
	if err != nil {
		return nil, err
	}
	relayersMap := make(map[uint32]SingleKeyHandler, len(relayersKeys))
	for _, relayerKey := range relayersKeys {
		shardID, errCompute := shardCoordinator.ComputeShardId(relayerKey.GetAddress())
		if errCompute != nil {
			return nil, errCompute
		}
		relayersMap[shardID] = relayerKey
	}

	for shardID := uint32(0); shardID < networkConfig.NumShardsWithoutMeta; shardID++ {
		if check.IfNil(relayersMap[shardID]) {
			return nil, fmt.Errorf("relayer key for shard %d is nil", shardID)
		}
	}

	return relayersMap, nil
}

// Process implements BalanceOperator
func (processor *relayedTxProcessor) Process(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
	networkConfig, err := processor.blockchainDataProvider.GetNetworkConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get network config: %w", err)
	}

	if check.IfNil(sender) {
		return errNilSender
	}
	senderBech32Address, err := sender.AddressAsBech32String()
	if err != nil {
		return fmt.Errorf("failed to convert sender address to bech32 string: %w", err)
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
		Sender:   senderBech32Address,
		GasLimit: processor.gasLimit,
		Data:     dataField,
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
	}
	err = processor.nonceTxHandler.ApplyNonceAndGasPrice(ctx, sender, tx)
	if err != nil {
		return fmt.Errorf("failed to apply nonce and gas price: %w", err)
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

	// 4. select the correct relayer (same shard with the sender)
	relayerKey, err := processor.selectRelayer(networkConfig, sender)
	if err != nil {
		return fmt.Errorf("failed to select a valid: %w", err)
	}

	// 5. Sign the frontend transaction with the singleKeyHandler and populate relayer specific fields
	relayerSig, err := relayerKey.Sign(unsignedtTxBytes)
	if err != nil {
		return fmt.Errorf("failed to sign the transaction with relayer key: %w", err)
	}
	tx.RelayerSignature = hex.EncodeToString(relayerSig)
	tx.RelayerAddr = relayerKey.GetBech32Address()

	// 6. Send transaction
	hash, err := processor.nonceTxHandler.SendTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Info("Transaction sent",
		"sender", senderBech32Address,
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

func (processor *relayedTxProcessor) selectRelayer(networkConfig *data.NetworkConfig, sender core.AddressHandler) (SingleKeyHandler, error) {
	shardCoordinator, err := blockchain.NewShardCoordinator(networkConfig.NumShardsWithoutMeta, 0)
	if err != nil {
		return nil, err
	}

	shardID, err := shardCoordinator.ComputeShardId(sender)
	if err != nil {
		return nil, err
	}

	relayerKey := processor.relayersKeys[shardID]
	if check.IfNil(relayerKey) {
		return nil, fmt.Errorf("no relayer key found for shard %d", shardID)
	}

	return relayerKey, nil
}

// Close closes any subcomponents it uses
func (processor *relayedTxProcessor) Close() error {
	return processor.nonceTxHandler.Close()
}

// IsInterfaceNil returns true if the value under the interface is nil
func (processor *relayedTxProcessor) IsInterfaceNil() bool {
	return processor == nil
}
