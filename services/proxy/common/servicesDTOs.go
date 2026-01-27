package common

// CryptoPaymentConfig response from crypto-payment service
type CryptoPaymentConfig struct {
	IsContractPaused bool   `json:"isContractPaused"`
	RequestsPerEGLD  uint64 `json:"requestsPerEGLD"`
	WalletURL        string `json:"walletURL"`
	ExplorerURL      string `json:"explorerURL"`
	ContractAddress  string `json:"contractAddress"`
}

// CreateAddressResponse response from create-address endpoint
type CreateAddressResponse struct {
	PaymentID uint64 `json:"paymentID"`
}

// AccountInfo response from account endpoint
type AccountInfo struct {
	PaymentID        uint64 `json:"paymentID"`
	Address          string `json:"address"`
	NumberOfRequests uint64 `json:"numberOfRequests"`
}
