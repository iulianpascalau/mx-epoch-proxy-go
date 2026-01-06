package integrationTests

import (
	"fmt"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/storage"
	"github.com/stretchr/testify/require"
)

var wrapper api.KeyAccessProvider
var keys []string

func BenchmarkSqlite(b *testing.B) {
	b.Skip("long run, should be run manually")

	b.StopTimer()
	if wrapper == nil {
		wrapper = setupBenchmark(b)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		key := getKey(i)
		b.StartTimer()

		_, _, _ = wrapper.IsKeyAllowed(key)
	}
}

func setupBenchmark(tb testing.TB) api.KeyAccessProvider {
	dbPath := tb.TempDir() + "/sqlite.db"

	keys = make([]string, 0, 2000)

	keyAccessProvider, err := storage.NewSQLiteWrapper(dbPath)
	require.NoError(tb, err)

	for i := 0; i < 100; i++ {
		user := fmt.Sprintf("user free %d", i)

		err = keyAccessProvider.AddUser(user, "pass", i%4 == 0, 0, string(common.FreeAccountType), true, "")
		require.NoError(tb, err)

		for j := 0; j < 10; j++ {
			key := common.GenerateKey()
			err = keyAccessProvider.AddKey(user, key)
			require.NoError(tb, err)
			keys = append(keys, key)
		}
	}

	for i := 0; i < 100; i++ {
		user := fmt.Sprintf("user premium %d", i)

		maxRequests := uint64(10 + i)
		err = keyAccessProvider.AddUser(user, "pass", i%4 == 0, maxRequests, string(common.PremiumAccountType), true, "")
		require.NoError(tb, err)

		for j := 0; j < 10; j++ {
			key := common.GenerateKey()
			err = keyAccessProvider.AddKey(user, key)
			require.NoError(tb, err)
			keys = append(keys, key)
		}
	}

	return keyAccessProvider
}

func getKey(i int) string {
	return keys[i%len(keys)]
}
