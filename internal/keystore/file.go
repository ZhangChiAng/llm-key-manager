package keystore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	errKeyStoreInvalid = "ERR_KEY_STORE_INVALID"
	keyStoreDirectory  = "LLM Key Manager"
	keyStoreFileName   = "keys.json"
)

type storedKeyRecord struct {
	ID             string `json:"id"`
	Provider       string `json:"provider"`
	Name           string `json:"name"`
	EncryptedValue string `json:"encryptedValue"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

func load() ([]storedKeyRecord, error) {
	path, err := filePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []storedKeyRecord{}, nil
	}
	if err != nil {
		return nil, err
	}

	var records []storedKeyRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, errors.New(errKeyStoreInvalid)
	}
	if records == nil {
		return []storedKeyRecord{}, nil
	}
	if err := validate(records); err != nil {
		return nil, err
	}

	return records, nil
}

func save(records []storedKeyRecord) error {
	path, err := filePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomically(path, data)
}

func writeFileAtomically(path string, data []byte) error {
	directory := filepath.Dir(path)
	tempFile, err := os.CreateTemp(directory, ".keys-*.tmp")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	removeTempFile := true
	defer func() {
		if removeTempFile {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	removeTempFile = false

	return nil
}

func validate(records []storedKeyRecord) error {
	for _, record := range records {
		if strings.TrimSpace(record.ID) == "" ||
			strings.TrimSpace(record.Provider) == "" ||
			strings.TrimSpace(record.Name) == "" ||
			strings.TrimSpace(record.EncryptedValue) == "" ||
			strings.TrimSpace(record.CreatedAt) == "" ||
			strings.TrimSpace(record.UpdatedAt) == "" {
			return errors.New(errKeyStoreInvalid)
		}
	}

	return nil
}

func filePath() (string, error) {
	baseDir, err := keyStoreBaseDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(baseDir, keyStoreDirectory, keyStoreFileName), nil
}
