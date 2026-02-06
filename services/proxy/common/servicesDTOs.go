package common

// CryptoPaymentConfig response from crypto-payment service
type CryptoPaymentConfig struct {
	IsContractPaused bool    `json:"isContractPaused"`
	CreditsPerEGLD   uint64  `json:"creditsPerEGLD"`
	WalletURL        string  `json:"walletURL"`
	ExplorerURL      string  `json:"explorerURL"`
	ContractAddress  string  `json:"contractAddress"`
	MinimumBalance   float64 `json:"minimumBalance"`
}

// CreateAddressResponse response from create-address endpoint
type CreateAddressResponse struct {
	PaymentID uint64 `json:"paymentID"`
}

// AccountInfo response from account endpoint
type AccountInfo struct {
	PaymentID uint64 `json:"paymentID"`
	Address   string `json:"address"`
	Credits   uint64 `json:"credits"`
}
