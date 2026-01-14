package process

import "errors"

var errNilDataProvider = errors.New("nil data provider")
var errNilBlockchainDataProvider = errors.New("nil blockchain data provider")
var errInvalidMinimumBalanceToCall = errors.New("invalid minimum balance to call the SC")
var errNilBalanceOperator = errors.New("nil balance operator")
var errNilUserKeysHandler = errors.New("nil user keys handler")
var errNilRelayerKeysHandler = errors.New("nil relayer keys handler")
var errZeroGasLimit = errors.New("gas limit must be greater than 0")
var errEmptyContractBech32Address = errors.New("empty contract bech32 address")
