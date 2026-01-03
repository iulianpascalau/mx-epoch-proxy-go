package common

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
	MaxRequests    uint64
	GlobalCounter  uint64
	Username       string
	HashedPassword string
	AccountType    AccountType
	IsAdmin        bool
}
