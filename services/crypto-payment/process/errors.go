package process

import "errors"

var errNilDataProvider = errors.New("nil data provider")
var errNilBlockchainDataProvider = errors.New("nil blockchain data provider")
var errInvalidMinimumBalanceToCall = errors.New("invalid minimum balance to call the SC")
var errNilBalanceOperator = errors.New("nil balance operator")
