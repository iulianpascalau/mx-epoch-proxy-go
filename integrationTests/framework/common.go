package framework

import (
	"path/filepath"
	"runtime"
)

// PaymentGasLimit is the gas limit for the payment transaction
const PaymentGasLimit = 50000

// GetContractPath returns the absolute path to the wasm file
func GetContractPath(contractName string) string {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Traverse up until we find the "services" directory
	for {
		if filepath.Base(currentDir) == "integrationTests" {
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

	return filepath.Join(currentDir, "requests-contract", "output", contractName, contractName+".wasm")
}
