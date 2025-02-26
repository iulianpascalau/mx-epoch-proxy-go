package main

import (
	"crypto/rand"

	logger "github.com/multiversx/mx-chain-logger-go"
)

const apiKeySize = 16

var log = logger.GetOrCreate("proxy")

func main() {
	buff := make([]byte, apiKeySize)
	_, _ = rand.Read(buff)

	log.Info("Generated API key", "key", buff)
}
