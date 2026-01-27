package integrationTests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/storage"
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

	b.StopTimer()
	_ = wrapper.Close()
	b.StartTimer()
}

func setupBenchmark(tb testing.TB) api.KeyAccessProvider {
	dbPath := tb.TempDir() + "/sqlite.db"

	keys = make([]string, 0, 2000)

	counters, _ := storage.NewCountersCache(time.Minute)
	keyAccessProvider, err := storage.NewSQLiteWrapper(dbPath, counters)
	require.NoError(tb, err)

	for i := 0; i < 100; i++ {
		user := fmt.Sprintf("user free %d", i)

		err = keyAccessProvider.AddUser(user, "pass", i%4 == 0, 0, false, true, "")
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
		err = keyAccessProvider.AddUser(user, "pass", i%4 == 0, maxRequests, true, true, "")
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
