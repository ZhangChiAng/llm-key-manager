package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	errKeyDuplicate    = "ERR_KEY_DUPLICATE"
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

func readStoredKeys() ([]storedKeyRecord, error) {
	path, err := keyStorePath()
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
	if err := validateStoredKeys(records); err != nil {
		return nil, err
	}

	return records, nil
}

func writeStoredKeys(records []storedKeyRecord) error {
	path, err := keyStorePath()
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

func validateStoredKeys(records []storedKeyRecord) error {
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

func ensureUniqueKeyRecord(records []storedKeyRecord, excludedID string, provider string, name string, value string) error {
	for _, record := range records {
		if record.ID == excludedID || strings.TrimSpace(record.Provider) != provider || strings.TrimSpace(record.Name) != name {
			continue
		}

		storedValue, err := decryptStoredValue(record)
		if err != nil {
			return err
		}
		if storedValue == value {
			return errors.New(errKeyDuplicate)
		}
	}

	return nil
}

func decryptStoredValue(record storedKeyRecord) (string, error) {
	protectedValue, err := base64.StdEncoding.DecodeString(record.EncryptedValue)
	if err != nil {
		return "", err
	}

	return unprotectKeyValue(protectedValue)
}

func maskKeyValue(value string) string {
	const (
		prefixLength  = 10
		suffixLength  = 4
		fullMask      = "*****************"
		separatorMask = "***"
	)

	if len(value) <= prefixLength+suffixLength {
		return fullMask
	}

	return value[:prefixLength] + separatorMask + value[len(value)-suffixLength:]
}

func keyStorePath() (string, error) {
	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData == "" {
		return "", errors.New("LOCALAPPDATA is not set")
	}

	return filepath.Join(localAppData, keyStoreDirectory, keyStoreFileName), nil
}
