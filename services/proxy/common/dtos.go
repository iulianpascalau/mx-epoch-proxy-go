package common

import "github.com/golang-jwt/jwt/v5"

// AccountType describes the account type
type AccountType string

// FreeAccountType defines the free account type, usually throttled
const FreeAccountType AccountType = "free"

// PremiumAccountType defines the premium account type, un-throttled
const PremiumAccountType AccountType = "premium"

// AccessKeyDetails holds details about an access key
type AccessKeyDetails struct {
	MaxRequests    uint64
	GlobalCounter  uint64
	KeyCounter     uint64
	Username       string
	HashedPassword string
	IsAdmin        bool
}

// UsersDetails holds details about a user
type UsersDetails struct {
	MaxRequests            uint64      `json:"MaxRequests"`
	GlobalCounter          uint64      `json:"GlobalCounter"`
	Username               string      `json:"Username"`
	HashedPassword         string      `json:"HashedPassword"`
	IsPremium              bool        `json:"IsPremium"`
	ProcessedAccountType   AccountType `json:"AccountType"`
	CryptoPaymentInitiated bool        `json:"CryptoPaymentInitiated"`
	IsUnlimited            bool        `json:"IsUnlimited"`
	IsActive               bool        `json:"IsActive"`
	IsAdmin                bool        `json:"IsAdmin"`
	CryptoPaymentID        uint64      `json:"PaymentID"`
}

// Claims struct holds the JWT claims
type Claims struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}
