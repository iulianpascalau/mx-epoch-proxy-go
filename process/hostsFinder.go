package process

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
)

const latestMarker = "latest"

type gatewayConfig struct {
	config.GatewayConfig

	epochStart    uint64
	epochEnd      uint64
	nonceStart    uint64
	nonceEnd      uint64
	hasLatestData bool
}

type hostsFinder struct {
	gateways          []gatewayConfig
	latestDataGateway *gatewayConfig
}

// NewHostsFinder will create a new hosts finder instance
func NewHostsFinder(configGateways []config.GatewayConfig) (*hostsFinder, error) {
	gateways, latestDataGateway, err := convertAndCheckGateways(configGateways)
	if err != nil {
		return nil, err
	}

	return &hostsFinder{
		gateways:          gateways,
		latestDataGateway: latestDataGateway,
	}, nil
}

func convertAndCheckGateways(configGateways []config.GatewayConfig) ([]gatewayConfig, *gatewayConfig, error) {
	if len(configGateways) == 0 {
		return nil, nil, errNoGatewayDefined
	}

	gatewayConfigs, latestDataConfig, err := convertGateways(configGateways)
	if err != nil {
		return nil, nil, err
	}

	sort.SliceStable(gatewayConfigs, func(i, j int) bool {
		return gatewayConfigs[i].epochStart < gatewayConfigs[j].epochStart
	})

	err = checkStartAndEndIntervals(gatewayConfigs)
	if err != nil {
		return nil, nil, err
	}

	return gatewayConfigs, latestDataConfig, nil
}

func convertGateways(configGateways []config.GatewayConfig) ([]gatewayConfig, *gatewayConfig, error) {
	gatewayConfigs := make([]gatewayConfig, len(configGateways))
	var latestDataConfig *gatewayConfig
	for i, cfg := range configGateways {
		gatewayConfigs[i].GatewayConfig = cfg

		val, err := strconv.Atoi(cfg.EpochStart)
		if err != nil {
			return nil, nil, fmt.Errorf("%w for epoch start at index %d with URL %s", err, i, cfg.URL)
		}
		gatewayConfigs[i].epochStart = uint64(val)

		val, err = strconv.Atoi(cfg.NonceStart)
		if err != nil {
			return nil, nil, fmt.Errorf("%w for nonce start at index %d with URL %s", err, i, cfg.URL)
		}
		gatewayConfigs[i].nonceStart = uint64(val)

		gatewayConfigs[i].EpochEnd = strings.ToLower(gatewayConfigs[i].EpochEnd)
		gatewayConfigs[i].NonceEnd = strings.ToLower(gatewayConfigs[i].NonceEnd)

		if gatewayConfigs[i].EpochEnd == latestMarker && gatewayConfigs[i].NonceEnd == latestMarker {
			gatewayConfigs[i].nonceEnd = math.MaxUint64
			gatewayConfigs[i].epochEnd = math.MaxUint64
			gatewayConfigs[i].hasLatestData = true

			if latestDataConfig != nil {
				return nil, nil, errMoreThanOneLatestDataGatewayFound
			}
			copiedConfig := gatewayConfigs[i] // we need to copy the value in a new instance. Otherwise, the following slice.Sort call might change the order and also change this value
			latestDataConfig = &copiedConfig

			continue
		}

		val, err = strconv.Atoi(cfg.EpochEnd)
		if err != nil {
			return nil, nil, fmt.Errorf("%w for epoch end at index %d with URL %s", err, i, cfg.URL)
		}
		gatewayConfigs[i].epochEnd = uint64(val)

		val, err = strconv.Atoi(cfg.NonceEnd)
		if err != nil {
			return nil, nil, fmt.Errorf("%w for nonce end at index %d with URL %s", err, i, cfg.URL)
		}
		gatewayConfigs[i].nonceEnd = uint64(val)

		if gatewayConfigs[i].nonceEnd < gatewayConfigs[i].nonceStart {
			return nil, nil, fmt.Errorf("%w when checking nonce end & nonce start at index %d with URL %s", errBadGatewayInterval, i, cfg.URL)
		}
		if gatewayConfigs[i].epochEnd < gatewayConfigs[i].epochStart {
			return nil, nil, fmt.Errorf("%w when checking epoch end & epoch start at index %d with URL %s", errBadGatewayInterval, i, cfg.URL)
		}
	}

	return gatewayConfigs, latestDataConfig, nil
}

func checkStartAndEndIntervals(gatewayConfigs []gatewayConfig) error {
	currentEpoch := -1
	currentNonce := -1
	for _, cfg := range gatewayConfigs {
		isEpochCorrect := int(cfg.epochStart)-1 == currentEpoch
		isNonceCorrect := int(cfg.nonceStart)-1 == currentNonce
		if isEpochCorrect && isNonceCorrect {
			currentEpoch = int(cfg.epochEnd)
			currentNonce = int(cfg.nonceEnd)
			continue
		}

		return fmt.Errorf("%w, EpochStart: %d, current epoch: %d, NonceStart: %d, current nonce: %d, URL: %s",
			errUnexpectedIntervalStart, cfg.epochStart, currentEpoch, cfg.nonceStart, currentNonce, cfg.URL)
	}

	return nil
}

// FindHost tries to find a matching host based on the URL values. Errors if it can not find a suitable host
func (finder *hostsFinder) FindHost(urlValues map[string][]string) (string, error) {
	if urlValues == nil {
		return "", fmt.Errorf("%w: %s", errCanNotDetermineSuitableHost, errNilMap.Error())
	}

	nonce, nonceFound, err := finder.parseToUint64(urlValues, UrlParameterBlockNonce)
	if err != nil {
		return "", err
	}

	epoch, epochFound, err := finder.parseToUint64(urlValues, UrlParameterHintEpoch)
	if err != nil {
		return "", err
	}

	if !nonceFound && !epochFound {
		if finder.latestDataGateway == nil {
			return "", errNoLatestDataGatewayDefined
		}

		return finder.latestDataGateway.URL, nil
	}

	if nonceFound {
		for _, cfg := range finder.gateways {
			if cfg.nonceStart <= nonce && nonce <= cfg.nonceEnd {
				return cfg.URL, nil
			}
		}

		return "", fmt.Errorf("%w for nonce %d", errNoGatewayDefined, nonce)
	}

	for _, cfg := range finder.gateways {
		if cfg.epochStart <= epoch && epoch <= cfg.epochEnd {
			return cfg.URL, nil
		}
	}

	return "", fmt.Errorf("%w for epoch %d", errNoGatewayDefined, epoch)
}

func (finder *hostsFinder) parseToUint64(urlValues map[string][]string, key string) (uint64, bool, error) {
	stringValues, exists := urlValues[key]
	if !exists {
		return 0, false, nil
	}

	if len(stringValues) == 0 {
		return 0, true, fmt.Errorf("%w for key %s", errMissingValue, key)
	}

	val, err := strconv.Atoi(stringValues[0])

	return uint64(val), true, err
}

// LoadedGateways returns the loaded config in the order that will be used
func (finder *hostsFinder) LoadedGateways() []config.GatewayConfig {
	results := make([]config.GatewayConfig, 0, len(finder.gateways))
	for _, cfg := range finder.gateways {
		results = append(results, cfg.GatewayConfig)
	}

	return results
}

// IsInterfaceNil returns true if the value under the interface is nil
func (finder *hostsFinder) IsInterfaceNil() bool {
	return finder == nil
}
