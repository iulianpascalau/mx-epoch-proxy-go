package process

import "errors"

var errNilDataProvider = errors.New("nil data provider")
var errNilBlockchainDataProvider = errors.New("nil blockchain data provider")
var errInvalidNumRequestsPerUnit = errors.New("invalid num requests per egld unit")
