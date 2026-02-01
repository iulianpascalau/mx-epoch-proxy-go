package framework

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	contractCreditsVersionTag = "v1.0.0"
	contractCreditsURL        = "https://github.com/iulianpascalau/credits-contract-rs/releases/download/" + contractCreditsVersionTag + "/credits-contract.zip"
)

func init() {
	if !IsChainSimulatorIsRunning() {
		return
	}
	err := ensureContractCredits()
	if err != nil {
		fmt.Printf("WARNING: Failed to ensure credits contract: %v\n", err)
		// We deliberately panic here because if we can't get the contract, tests will fail anyway
		panic(err)
	}
}

func ensureContractCredits() error {
	root := traverse("integrationTests")
	extractTarget := filepath.Join(root, "contracts")
	contractDir := filepath.Join(extractTarget, "credits")
	contractPath := filepath.Join(contractDir, "credits.wasm")

	if _, err := os.Stat(contractPath); err == nil {
		// Contract exists. We could check versioning (e.g. hash) but for now existence matches intent.
		return nil
	}

	fmt.Printf("Downloading credits contract from %s...\n", contractCreditsURL)
	resp, err := http.Get(contractCreditsURL)
	if err != nil {
		return fmt.Errorf("failed to download contract: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "credits-contract-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	_ = tmpFile.Close()

	// Unzip
	r, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open zip reader: %w", err)
	}
	defer func() {
		_ = r.Close()
	}()

	err = os.MkdirAll(extractTarget, 0755)

	if err != nil {
		return fmt.Errorf("failed to create target dir: %w", err)
	}

	for _, f := range r.File {
		fpath := filepath.Join(extractTarget, f.Name)

		// ZipSlip check
		if !strings.HasPrefix(fpath, filepath.Clean(extractTarget)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		if err != nil {
			return err
		}

		var outFile *os.File
		outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			_ = outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return err
		}
	}

	fmt.Println("Successfully downloaded and extracted credits contract.")
	return nil
}
