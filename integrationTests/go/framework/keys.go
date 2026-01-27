package framework

import (
	"testing"

	mxCrypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/stretchr/testify/require"
)

// KeysHolder holds a 2 pk-sk pairs for Mvx chain
type KeysHolder struct {
	MvxAddress *MvxAddress
	MvxSk      []byte
	ID         uint64
	PayAddress *MvxAddress
}

// KeysStore will hold all the keys used in the test
type KeysStore struct {
	testing.TB
	OwnerKeys    KeysHolder
	RelayersKeys []KeysHolder
	UserAKeys    KeysHolder
	UserBKeys    KeysHolder
	UserCKeys    KeysHolder
}

// NewKeysStore will create a KeysStore instance and generate all keys
func NewKeysStore(
	tb testing.TB,
	numShards uint32,
) *KeysStore {
	keysStore := &KeysStore{
		TB: tb,
	}

	keysStore.OwnerKeys = keysStore.generateKey(0)
	log.Info("generated owner",
		"MvX address", keysStore.OwnerKeys.MvxAddress.Bech32())

	for i := 0; i < int(numShards); i++ {
		relayer := keysStore.generateKey(byte(i))
		keysStore.RelayersKeys = append(keysStore.RelayersKeys, relayer)
		log.Info("generated relayer", "shard", i,
			"MvX address", relayer.MvxAddress.Bech32())
	}

	keysStore.UserAKeys = keysStore.generateKey(2)
	log.Info("generated user A",
		"MvX address", keysStore.UserAKeys.MvxAddress.Bech32())

	keysStore.UserBKeys = keysStore.generateKey(0)
	log.Info("generated user B",
		"MvX address", keysStore.UserBKeys.MvxAddress.Bech32())

	keysStore.UserCKeys = keysStore.generateKey(2)
	log.Info("generated user C",
		"MvX address", keysStore.UserCKeys.MvxAddress.Bech32())

	return keysStore
}

func (keyStore *KeysStore) getAllKeys() []KeysHolder {
	allKeys := make([]KeysHolder, 0, 100)
	allKeys = append(allKeys, keyStore.OwnerKeys)
	allKeys = append(allKeys, keyStore.RelayersKeys...)
	allKeys = append(allKeys, keyStore.UserAKeys)
	allKeys = append(allKeys, keyStore.UserBKeys)
	allKeys = append(allKeys, keyStore.UserCKeys)

	return allKeys
}

func (keyStore *KeysStore) generateKey(projectedShard byte) KeysHolder {
	keys := GenerateMvxPrivatePublicKey(keyStore, projectedShard)

	return keys
}

// WalletsToFundOnMultiversX will return the wallets to fund on MultiversX
func (keyStore *KeysStore) WalletsToFundOnMultiversX() []string {
	allKeys := keyStore.getAllKeys()
	walletsToFund := make([]string, 0, len(allKeys))

	for _, key := range allKeys {
		walletsToFund = append(walletsToFund, key.MvxAddress.Bech32())
	}

	return walletsToFund
}

// GenerateMvxPrivatePublicKey will generate a new keys holder instance that will hold only the MultiversX generated keys
func GenerateMvxPrivatePublicKey(tb testing.TB, projectedShard byte) KeysHolder {
	sk, pkBytes := generateSkPkInShard(tb, projectedShard)

	skBytes, err := sk.ToByteArray()
	require.Nil(tb, err)

	return KeysHolder{
		MvxSk:      skBytes,
		MvxAddress: NewMvxAddressFromBytes(tb, pkBytes),
	}
}

func generateSkPkInShard(tb testing.TB, projectedShard byte) (mxCrypto.PrivateKey, []byte) {
	var sk mxCrypto.PrivateKey
	var pk mxCrypto.PublicKey

	for {
		sk, pk = keyGenerator.GeneratePair()

		pkBytes, err := pk.ToByteArray()
		require.Nil(tb, err)

		if pkBytes[len(pkBytes)-1] == projectedShard {
			return sk, pkBytes
		}
	}
}
