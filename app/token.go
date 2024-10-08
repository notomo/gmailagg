package app

import (
	"fmt"
	"os"
	"path/filepath"
)

func createTokenReader(path string) (*os.File, error) {
	tokenFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file path: %w", err)
	}
	return tokenFile, nil
}

func createTokenWriter(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	tokenFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("open token file: %w", err)
	}

	return tokenFile, nil
}

func GetDefaultTokenPath() (string, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	return filepath.Join(confDir, "gmailagg", "token.json"), nil
}
