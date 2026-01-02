package common

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
	IsAdmin        bool
}
