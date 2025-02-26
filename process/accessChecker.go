package process

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
)

const headerApiKey = "X-Api-Key"
const uriSeparator = "/"

var allowedVersions = []string{"v1"}

type accessChecker struct {
	keys []config.AccessKeyConfig
}

// NewAccessChecker creates a new instance of type access checker
func NewAccessChecker(accessKeys []config.AccessKeyConfig) (*accessChecker, error) {
	processedAccessKeys, err := checkKeys(accessKeys)
	if err != nil {
		return nil, err
	}

	return &accessChecker{
		keys: processedAccessKeys,
	}, nil
}

func checkKeys(accessKeys []config.AccessKeyConfig) ([]config.AccessKeyConfig, error) {
	uniqueAliases := make(map[string]struct{})

	processedAccessKeys := make([]config.AccessKeyConfig, 0, len(accessKeys))
	for i, key := range accessKeys {
		key.Key = strings.TrimSpace(key.Key)
		key.Alias = strings.TrimSpace(key.Alias)

		if strings.ToLower(key.Alias) == strings.ToLower(common.AllAliases) {
			continue
		}

		if len(key.Key) == 0 {
			return nil, fmt.Errorf("%w for key at position %d", errEmptyKey, i)
		}
		if len(key.Alias) == 0 {
			return nil, fmt.Errorf("%w for alias at position %d", errEmptyAlias, i)
		}

		lowerCaseAlias := strings.ToLower(key.Alias)
		_, found := uniqueAliases[lowerCaseAlias]
		if found {
			return nil, fmt.Errorf("%w for alias %s", errDuplicatedAccessKeyAlias, lowerCaseAlias)
		}

		uniqueAliases[lowerCaseAlias] = struct{}{}
		processedAccessKeys = append(processedAccessKeys, config.AccessKeyConfig{
			Key:   strings.ToLower(key.Key),
			Alias: key.Alias,
		})
	}

	if len(processedAccessKeys) == 0 {
		log.Warn("no access keys provided, will process all requests")
	}

	return processedAccessKeys, nil
}

// ShouldProcessRequest returns true if the request is allowed to be processed
func (checker *accessChecker) ShouldProcessRequest(header http.Header, requestURI string) (string, string, error) {
	accessKey, processedRequestURI := processRequestURI(requestURI)
	alias := checker.getAllowedAlias(accessKey)
	if len(alias) > 0 {
		// authorized, useless to check the header
		return processedRequestURI, alias, nil
	}

	accessKey = parseHeaderForAccessKey(header)
	alias = checker.getAllowedAlias(accessKey)
	if len(alias) > 0 {
		return processedRequestURI, alias, nil
	}

	return "", "", errUnauthorized
}

func processRequestURI(inputRequestURI string) (*config.AccessKeyConfig, string) {
	splt := strings.Split(inputRequestURI, uriSeparator)
	if len(splt) < 4 {
		// token was not provided in the URL
		return nil, inputRequestURI
	}
	if !checkVersion(splt[1]) {
		// token was not provided in the URL
		return nil, inputRequestURI
	}

	return &config.AccessKeyConfig{
			Key: strings.ToLower(splt[2]),
		},
		uriSeparator + strings.Join(splt[3:], uriSeparator)
}

func checkVersion(version string) bool {
	for _, vers := range allowedVersions {
		if vers == strings.ToLower(version) {
			return true
		}
	}

	return false
}

func (checker *accessChecker) getAllowedAlias(accessKey *config.AccessKeyConfig) string {
	if len(checker.keys) == 0 {
		return common.AllAliases
	}

	if accessKey == nil {
		return ""
	}
	for _, key := range checker.keys {
		if key.Key == accessKey.Key {
			return key.Alias
		}
	}

	return ""
}

func parseHeaderForAccessKey(header http.Header) *config.AccessKeyConfig {
	val := header.Get(headerApiKey)
	if len(val) == 0 {
		return nil
	}

	return &config.AccessKeyConfig{
		Key: strings.ToLower(val),
	}
}

// GetAllAliases return all aliases stored
func (checker *accessChecker) GetAllAliases() []string {
	aliases := make([]string, 0, len(checker.keys))
	for _, accessKey := range checker.keys {
		aliases = append(aliases, accessKey.Alias)
	}

	return aliases
}

// IsInterfaceNil returns true if the value under the interface is nil
func (checker *accessChecker) IsInterfaceNil() bool {
	return checker == nil
}
