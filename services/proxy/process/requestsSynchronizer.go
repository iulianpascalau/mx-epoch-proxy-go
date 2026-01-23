package process

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// RequestsSynchronizer synchronizes user requests limits with the crypto-payment service
type requestsSynchronizer struct {
	store        UsersSyncerStore
	cryptoClient CryptoDataFetcher
}

// NewRequestsSynchronizer creates a new instance of requestsSynchronizer
func NewRequestsSynchronizer(
	store UsersSyncerStore,
	cryptoClient CryptoDataFetcher,
) (*requestsSynchronizer, error) {
	if check.IfNil(store) {
		return nil, errNilStorer
	}
	if check.IfNil(cryptoClient) {
		return nil, errNilCryptoClient
	}

	return &requestsSynchronizer{
		store:        store,
		cryptoClient: cryptoClient,
	}, nil
}

// Process will process all available DB users to have their maximum requests value updated
func (rs *requestsSynchronizer) Process() {
	users, err := rs.store.GetAllUsers()
	if err != nil {
		log.Warn("failed to get all users", "error", err)
	}

	for _, user := range users {
		if user.CryptoPaymentID == 0 {
			continue
		}

		accountInfo, errGet := rs.cryptoClient.GetAccount(user.CryptoPaymentID)
		if errGet != nil {
			log.Warn("failed to get account info from crypto service", "user", user.Username, "paymentID", user.CryptoPaymentID, "error", errGet)
			continue
		}

		// The DB should be updated if the contract's value exceeds what we have in the DB
		if accountInfo.NumberOfRequests > user.MaxRequests {
			log.Info("updating user max requests from contract", "user", user.Username, "old", user.MaxRequests, "new", accountInfo.NumberOfRequests)

			errUpdate := rs.store.UpdateMaxRequests(user.Username, accountInfo.NumberOfRequests)
			if errUpdate != nil {
				log.Error("failed to update user max requests", "user", user.Username, "error", errUpdate)
			}
		}
	}
}

// IsInterfaceNil returns true if the value under the interface is nil
func (rs *requestsSynchronizer) IsInterfaceNil() bool {
	return rs == nil
}
