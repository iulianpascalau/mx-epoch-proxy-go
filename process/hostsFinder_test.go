package process

import (
	"math"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/stretchr/testify/assert"
)

func createTestConfigs() []config.GatewayConfig {
	return []config.GatewayConfig{
		{
			URL:        "URL1",
			EpochStart: "100",
			EpochEnd:   "laTeST",
			NonceStart: "10000",
			NonceEnd:   "LAtesT",
		},
		{
			URL:        "URL2",
			EpochStart: "0",
			EpochEnd:   "49",
			NonceStart: "0",
			NonceEnd:   "4999",
		},
		{
			URL:        "URL3",
			EpochStart: "50",
			EpochEnd:   "99",
			NonceStart: "5000",
			NonceEnd:   "9999",
		},
	}
}

func TestNewHostsFinder(t *testing.T) {
	t.Parallel()

	t.Run("no gateways defined, should error", func(t *testing.T) {
		t.Parallel()

		finder, err := NewHostsFinder(make([]config.GatewayConfig, 0))
		assert.Equal(t, errNoGatewayDefined, err)
		assert.Nil(t, finder)
	})
	t.Run("epoch start is not a valid number, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg[1].EpochStart = "NaN"
		finder, err := NewHostsFinder(cfg)
		assert.Contains(t, err.Error(), "epoch start at index 1")
		assert.Nil(t, finder)
	})
	t.Run("nonce start is not a valid number, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg[1].NonceStart = "NaN"
		finder, err := NewHostsFinder(cfg)
		assert.Contains(t, err.Error(), "nonce start at index 1")
		assert.Nil(t, finder)
	})
	t.Run("2 gateways with latest data, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg = append(cfg, cfg[0])
		finder, err := NewHostsFinder(cfg)
		assert.ErrorIs(t, err, errMoreThanOneLatestDataGatewayFound)
		assert.Nil(t, finder)
	})
	t.Run("epoch end is not a valid number, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg[1].EpochEnd = "NaN"
		finder, err := NewHostsFinder(cfg)
		assert.Contains(t, err.Error(), "epoch end at index 1")
		assert.Nil(t, finder)
	})
	t.Run("nonce end is not a valid number, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg[1].NonceEnd = "NaN"
		finder, err := NewHostsFinder(cfg)
		assert.Contains(t, err.Error(), "nonce end at index 1")
		assert.Nil(t, finder)
	})
	t.Run("epoch end is lower than epoch start, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg[1].EpochStart = "50"
		finder, err := NewHostsFinder(cfg)
		assert.ErrorIs(t, err, errBadGatewayInterval)
		assert.Contains(t, err.Error(), "when checking epoch end & epoch start at index 1")
		assert.Nil(t, finder)
	})
	t.Run("nonce end is lower than nonce start, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg[1].NonceStart = "5000"
		finder, err := NewHostsFinder(cfg)
		assert.ErrorIs(t, err, errBadGatewayInterval)
		assert.Contains(t, err.Error(), "when checking nonce end & nonce start at index 1")
		assert.Nil(t, finder)
	})
	t.Run("missing interval, should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg[1].NonceEnd = "4998"
		finder, err := NewHostsFinder(cfg)
		assert.ErrorIs(t, err, errUnexpectedIntervalStart)
		assert.Contains(t, err.Error(), "EpochStart: 50, current epoch: 49, NonceStart: 5000, current nonce: 4998")
		assert.Nil(t, finder)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		finder, err := NewHostsFinder(createTestConfigs())
		assert.Nil(t, err)
		assert.NotNil(t, finder)

		expectedGateways := []gatewayConfig{
			{
				GatewayConfig: config.GatewayConfig{
					URL:        "URL2",
					EpochStart: "0",
					EpochEnd:   "49",
					NonceStart: "0",
					NonceEnd:   "4999",
				},
				epochStart:    0,
				epochEnd:      49,
				nonceStart:    0,
				nonceEnd:      4999,
				hasLatestData: false,
			},
			{
				GatewayConfig: config.GatewayConfig{
					URL:        "URL3",
					EpochStart: "50",
					EpochEnd:   "99",
					NonceStart: "5000",
					NonceEnd:   "9999",
				},
				epochStart:    50,
				epochEnd:      99,
				nonceStart:    5000,
				nonceEnd:      9999,
				hasLatestData: false,
			},
			{
				GatewayConfig: config.GatewayConfig{
					URL:        "URL1",
					EpochStart: "100",
					EpochEnd:   "latest",
					NonceStart: "10000",
					NonceEnd:   "latest",
				},
				epochStart:    100,
				epochEnd:      math.MaxUint64,
				nonceStart:    10000,
				nonceEnd:      math.MaxUint64,
				hasLatestData: true,
			},
		}
		assert.NotNil(t, finder.latestDataGateway)

		assert.Equal(t, expectedGateways, finder.gateways)
		assert.Equal(t, expectedGateways[2], *finder.latestDataGateway)
	})
}

func TestHostsFinder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *hostsFinder
	assert.True(t, instance.IsInterfaceNil())

	instance = &hostsFinder{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestHostsFinder_FindHost(t *testing.T) {
	t.Parallel()

	t.Run("nil url values map should error", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())
		cfg, err := finder.FindHost(nil)
		assert.Empty(t, cfg.URL)
		assert.ErrorIs(t, err, errCanNotDetermineSuitableHost)
		assert.Contains(t, err.Error(), errNilMap.Error())
	})
	t.Run("nil url values map should error", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())
		cfg, err := finder.FindHost(nil)
		assert.Empty(t, cfg.URL)
		assert.ErrorIs(t, err, errCanNotDetermineSuitableHost)
		assert.Contains(t, err.Error(), errNilMap.Error())
	})
	t.Run("no nonce or epoch provided should return the latest URL", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())
		cfg, err := finder.FindHost(make(map[string][]string))
		assert.Nil(t, err)
		assert.Equal(t, "URL1", cfg.URL)
	})
	t.Run("no nonce or epoch provided with no latest data should error", func(t *testing.T) {
		t.Parallel()

		configs := createTestConfigs()
		configs[0].EpochEnd = "900"
		configs[0].NonceEnd = "90000"
		finder, _ := NewHostsFinder(configs)
		cfg, err := finder.FindHost(make(map[string][]string))
		assert.Empty(t, cfg.URL)
		assert.ErrorIs(t, err, errNoLatestDataGatewayDefined)
	})
	t.Run("no value for the nonce key should error", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())
		urlValues := map[string][]string{
			UrlParameterBlockNonce: nil,
		}
		cfg, err := finder.FindHost(urlValues)
		assert.Empty(t, cfg.URL)
		assert.ErrorIs(t, err, errMissingValue)
		assert.Contains(t, err.Error(), "for key blockNonce")
	})
	t.Run("not a number on index 0 for the nonce key should error", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())
		urlValues := map[string][]string{
			UrlParameterBlockNonce: {"NaN"},
		}
		cfg, err := finder.FindHost(urlValues)
		assert.Empty(t, cfg.URL)
		assert.Contains(t, err.Error(), "strconv.Atoi: parsing")
	})
	t.Run("no value for the epoch key should error", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())
		urlValues := map[string][]string{
			UrlParameterHintEpoch: nil,
		}
		cfg, err := finder.FindHost(urlValues)
		assert.Empty(t, cfg.URL)
		assert.ErrorIs(t, err, errMissingValue)
		assert.Contains(t, err.Error(), "for key hintEpoch")
	})
	t.Run("not a number on index 0 for the epoch key should error", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())
		urlValues := map[string][]string{
			UrlParameterHintEpoch: {"NaN"},
		}
		cfg, err := finder.FindHost(urlValues)
		assert.Empty(t, cfg.URL)
		assert.Contains(t, err.Error(), "strconv.Atoi: parsing")
	})
	t.Run("no nonce & epoch found, should error", func(t *testing.T) {
		t.Parallel()

		configs := createTestConfigs()
		configs[0].NonceEnd = "90000"
		configs[0].EpochEnd = "900"

		finder, _ := NewHostsFinder(configs)

		t.Run("with providing out of bound nonce", func(t *testing.T) {
			t.Parallel()

			urlValues := map[string][]string{
				UrlParameterBlockNonce: {"90001"},
			}
			cfg, err := finder.FindHost(urlValues)
			assert.Empty(t, cfg.URL)
			assert.ErrorIs(t, err, errNoGatewayDefined)
			assert.Contains(t, err.Error(), "for nonce 90001")

			urlValues[UrlParameterBlockNonce] = []string{"9000001"}
			cfg, err = finder.FindHost(urlValues)
			assert.Empty(t, cfg.URL)
			assert.ErrorIs(t, err, errNoGatewayDefined)
			assert.Contains(t, err.Error(), "for nonce 9000001")
		})
		t.Run("with providing out of bound epoch", func(t *testing.T) {
			t.Parallel()

			urlValues := map[string][]string{
				UrlParameterHintEpoch: {"901"},
			}
			cfg, err := finder.FindHost(urlValues)
			assert.Empty(t, cfg.URL)
			assert.ErrorIs(t, err, errNoGatewayDefined)
			assert.Contains(t, err.Error(), "for epoch 901")

			urlValues[UrlParameterHintEpoch] = []string{"90001"}
			cfg, err = finder.FindHost(urlValues)
			assert.Empty(t, cfg.URL)
			assert.ErrorIs(t, err, errNoGatewayDefined)
			assert.Contains(t, err.Error(), "for epoch 90001")
		})

	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		finder, _ := NewHostsFinder(createTestConfigs())

		t.Run("with providing nonce", func(t *testing.T) {
			t.Parallel()

			urlValues := map[string][]string{
				UrlParameterBlockNonce: {"0"},
			}
			cfg, err := finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterBlockNonce] = []string{"1"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterBlockNonce] = []string{"2500"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterBlockNonce] = []string{"4999"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterBlockNonce] = []string{"5000"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterBlockNonce] = []string{"5001"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterBlockNonce] = []string{"7500"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterBlockNonce] = []string{"9999"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterBlockNonce] = []string{"10000"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")

			urlValues[UrlParameterBlockNonce] = []string{"10001"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")

			urlValues[UrlParameterBlockNonce] = []string{"100000"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")

			urlValues[UrlParameterBlockNonce] = []string{"10000000"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")
		})
		t.Run("with providing epoch", func(t *testing.T) {
			t.Parallel()

			urlValues := map[string][]string{
				UrlParameterHintEpoch: {"0"},
			}
			cfg, err := finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterHintEpoch] = []string{"1"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterHintEpoch] = []string{"25"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterHintEpoch] = []string{"49"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL2")

			urlValues[UrlParameterHintEpoch] = []string{"50"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterHintEpoch] = []string{"51"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterHintEpoch] = []string{"75"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterHintEpoch] = []string{"99"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL3")

			urlValues[UrlParameterHintEpoch] = []string{"100"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")

			urlValues[UrlParameterHintEpoch] = []string{"101"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")

			urlValues[UrlParameterHintEpoch] = []string{"1000"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")

			urlValues[UrlParameterHintEpoch] = []string{"100000"}
			cfg, err = finder.FindHost(urlValues)
			assert.Nil(t, err)
			assert.Equal(t, cfg.URL, "URL1")
		})
	})
}

func TestHostsFinder_LoadedGateways(t *testing.T) {
	t.Parallel()

	cfg := createTestConfigs()
	finder, _ := NewHostsFinder(cfg)
	expectedResult := []config.GatewayConfig{
		cfg[1],
		cfg[2],
		cfg[0],
	}
	expectedResult[2].NonceEnd = "latest"
	expectedResult[2].EpochEnd = "latest"

	assert.Equal(t, expectedResult, finder.LoadedGateways())
}
