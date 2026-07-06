package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	errKeyDuplicate    = "ERR_KEY_DUPLICATE"
	errKeyStoreInvalid = "ERR_KEY_STORE_INVALID"
)

// App manages backend state and exposes methods to the Wails frontend.
type App struct {
	ctx context.Context
}

// KeyRecord describes API key metadata returned to the frontend.
type KeyRecord struct {
	// ID is a generated unique identifier based on the creation timestamp.
	ID string `json:"id"`
	// Provider names the API provider this key belongs to.
	Provider string `json:"provider"`
	// Name is the user-facing label for the key.
	Name string `json:"name"`
	// MaskedValue contains the redacted API key content for display.
	MaskedValue string `json:"maskedValue"`
	// CreatedAt records when the key was saved in RFC3339 format.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt records when the key metadata or encrypted value last changed in RFC3339 format.
	UpdatedAt string `json:"updatedAt"`
}

type storedKeyRecord struct {
	ID             string `json:"id"`
	Provider       string `json:"provider"`
	Name           string `json:"name"`
	EncryptedValue string `json:"encryptedValue"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SaveKey validates and stores an encrypted API key in the local key store.
func (a *App) SaveKey(provider string, name string, value string) error {
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

	if provider == "" || name == "" || value == "" {
		return errors.New("provider, name, and key content are required")
	}

	records, err := readStoredKeys()
	if err != nil {
		return err
	}

	if err := ensureUniqueKeyRecord(records, "", provider, name, value); err != nil {
		return err
	}

	protectedValue, err := protectKeyValue(value)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Format(time.RFC3339)
	records = append(records, storedKeyRecord{
		ID:             strconv.FormatInt(now.UnixNano(), 10),
		Provider:       provider,
		Name:           name,
		EncryptedValue: base64.StdEncoding.EncodeToString(protectedValue),
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
	})

	return writeStoredKeys(records)
}

// ListKeys returns all locally saved API keys with masked values only.
func (a *App) ListKeys() ([]KeyRecord, error) {
	storedRecords, err := readStoredKeys()
	if err != nil {
		return nil, err
	}

	records := make([]KeyRecord, 0, len(storedRecords))
	for _, record := range storedRecords {
		value, err := decryptStoredValue(record)
		if err != nil {
			return nil, err
		}

		records = append(records, KeyRecord{
			ID:          record.ID,
			Provider:    record.Provider,
			Name:        record.Name,
			MaskedValue: maskKeyValue(value),
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		})
	}

	return records, nil
}

// UpdateKey updates key metadata and optionally replaces the encrypted key value.
func (a *App) UpdateKey(id string, provider string, name string, value string) error {
	id = strings.TrimSpace(id)
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

	if provider == "" || name == "" {
		return errors.New("provider and name are required")
	}

	records, err := readStoredKeys()
	if err != nil {
		return err
	}

	recordIndex := -1
	var currentRecord storedKeyRecord
	for index, record := range records {
		if record.ID != id {
			continue
		}

		recordIndex = index
		currentRecord = record
		break
	}

	if recordIndex == -1 {
		return errors.New("key record not found")
	}

	providerChanged := strings.TrimSpace(currentRecord.Provider) != provider
	nameChanged := strings.TrimSpace(currentRecord.Name) != name
	valueChanged := value != ""
	if !providerChanged && !nameChanged && !valueChanged {
		return nil
	}

	effectiveValue := value
	if effectiveValue == "" {
		effectiveValue, err = decryptStoredValue(currentRecord)
		if err != nil {
			return err
		}
	}

	if err := ensureUniqueKeyRecord(records, id, provider, name, effectiveValue); err != nil {
		return err
	}

	if valueChanged {
		protectedValue, err := protectKeyValue(effectiveValue)
		if err != nil {
			return err
		}
		currentRecord.EncryptedValue = base64.StdEncoding.EncodeToString(protectedValue)
	}

	currentRecord.Provider = provider
	currentRecord.Name = name
	currentRecord.UpdatedAt = time.Now().Format(time.RFC3339)
	records[recordIndex] = currentRecord

	return writeStoredKeys(records)
}

// CopyKey decrypts the saved API key and writes the plaintext to the clipboard.
func (a *App) CopyKey(id string) error {
	id = strings.TrimSpace(id)

	records, err := readStoredKeys()
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.ID != id {
			continue
		}

		value, err := decryptStoredValue(record)
		if err != nil {
			return err
		}

		if a.ctx == nil {
			return errors.New("application context is not ready")
		}

		return runtime.ClipboardSetText(a.ctx, value)
	}

	return errors.New("key record not found")
}

// DeleteKey removes the saved API key matching the provided record ID.
func (a *App) DeleteKey(id string) error {
	id = strings.TrimSpace(id)

	records, err := readStoredKeys()
	if err != nil {
		return err
	}

	nextRecords := make([]storedKeyRecord, 0, len(records))
	found := false
	for _, record := range records {
		if record.ID == id {
			found = true
			continue
		}

		nextRecords = append(nextRecords, record)
	}

	if !found {
		return errors.New("key record not found")
	}

	return writeStoredKeys(nextRecords)
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
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(executablePath), "keys.json"), nil
}
