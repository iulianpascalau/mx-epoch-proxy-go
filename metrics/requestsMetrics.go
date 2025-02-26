package metrics

import (
	"context"
	"fmt"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const totalMarker = "_total"

type requestsMetrics struct {
	storer Storer
}

// NewRequestMetrics will create a new instance of type requests metrics
func NewRequestMetrics(storer Storer) (*requestsMetrics, error) {
	if check.IfNil(storer) {
		return nil, errNilStorer
	}

	instance := &requestsMetrics{
		storer: storer,
	}

	return instance, nil
}

// ProcessedResponse will increment the number of requests processed on an alias
func (metrics *requestsMetrics) ProcessedResponse(alias string) {
	_ = metrics.storer.Increment(context.Background(), alias+totalMarker)

	if alias != common.AllAliases {
		_ = metrics.storer.Increment(context.Background(), common.AllAliases+totalMarker)
	}
}

// GetAllKeyValues will fetch all (key, value) pairs from the storer
func (metrics *requestsMetrics) GetAllKeyValues() []string {
	keys, err := metrics.storer.GetAllKeys(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("%s while getting all the keys", err.Error())}
	}

	data := make([]string, 0, len(keys))
	for _, key := range keys {
		val, _, errGet := metrics.storer.Get(context.Background(), key)
		if errGet != nil {
			val = fmt.Sprintf("%s while getting data", errGet.Error())
		}

		data = append(data, fmt.Sprintf("    %s: %s", key, val))
	}

	return data
}

// IsInterfaceNil returns true if the value under the interface is nil
func (metrics *requestsMetrics) IsInterfaceNil() bool {
	return metrics == nil
}
