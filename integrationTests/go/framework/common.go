package framework

import (
	"path/filepath"
	"runtime"
)

// PaymentGasLimit is the gas limit for the payment transaction
const PaymentGasLimit = 50000

// CallGasLimit is the gas limit for a SC call
const CallGasLimit = 3000000

// GetContractPath returns the absolute path to the wasm file
func GetContractPath(contractName string) string {
	currentDir := traverse("integrationTests")

	return filepath.Join(currentDir, "requests-contract", "output", contractName, contractName+".wasm")
}

// GetProxyRootPath returns the absolute path to the proxy service root path
func GetProxyRootPath(templateFile string) string {
	currentDir := traverse("integrationTests")

	return filepath.Join(currentDir, "services", "proxy", templateFile)
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
