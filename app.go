package main

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SaveKey stores an encrypted API key after the frontend validates required fields.
func (a *App) SaveKey(provider string, name string, value string) error {
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

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

// UpdateKey updates frontend-validated metadata and optionally replaces the encrypted key value.
// An empty value preserves the existing encrypted key.
func (a *App) UpdateKey(id string, provider string, name string, value string) error {
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

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

		return runtime.ClipboardSetText(a.ctx, value)
	}

	return errors.New("key record not found")
}

// DeleteKey removes the saved API key matching the provided record ID.
func (a *App) DeleteKey(id string) error {
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
