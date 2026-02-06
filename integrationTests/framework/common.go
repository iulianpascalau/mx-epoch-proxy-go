package framework

import (
	"path/filepath"
	"runtime"
	"testing"

	cryptoPaymentsFramework "github.com/iulianpascalau/mx-crypto-payments-go/integrationTests/framework"
	"github.com/stretchr/testify/require"
)

// GetProxyRootPath returns the absolute path to the proxy service root path
func GetProxyRootPath(templateFile string) string {
	currentDir := traverse("integrationTests")

	return filepath.Join(currentDir, "services", "proxy", templateFile)
}

// GetContractPath returns the absolute path to the wasm file
func GetContractPath(contractName string) string {
	currentDir := traverse("integrationTests")

	return filepath.Join(currentDir, "contracts", contractName, contractName+".wasm")
}

// EnsureTestContracts test if the contracts are present in the project, if not, download them
func EnsureTestContracts(tb testing.TB) {
	root := traverse("integrationTests")
	extractTarget := filepath.Join(root, "contracts")

	err := cryptoPaymentsFramework.EnsureContractCredits(cryptoPaymentsFramework.ContractCreditsURL, extractTarget)
	require.NoError(tb, err)
}

func traverse(upToDir string) string {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Traverse up until we find the "services" directory
	for {
		if filepath.Base(currentDir) == upToDir {
			// Found 'integrationTests', go one level up to project root
			currentDir = filepath.Join(currentDir, "../")
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached filesystem root without finding 'services'
			break
		}
		currentDir = parent
	}

	return currentDir
}
