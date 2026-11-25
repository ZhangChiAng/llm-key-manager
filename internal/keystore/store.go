// Package keystore persists API keys in keys.json, encrypted with the platform protector.
package keystore

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"
)

const errKeyDuplicate = "ERR_KEY_DUPLICATE"

var errRecordNotFound = errors.New("key record not found")

// Store reads and writes keys.json under the platform data directory. Each operation
// resolves the path and reads the file anew, preserving path errors and external edits.
type Store struct{}

// Record is saved metadata with a still-encrypted value, decrypted only on demand.
type Record struct {
	ID, Provider, Name, CreatedAt, UpdatedAt string
	encryptedValue                           string
}

func recordFromStored(record storedKeyRecord) Record {
	return Record{
		ID: record.ID, Provider: record.Provider, Name: record.Name,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
		encryptedValue: record.EncryptedValue,
	}
}

// Reveal decrypts the saved value without modifying storage.
func (r Record) Reveal() (string, error) {
	return decryptStoredValue(storedKeyRecord{EncryptedValue: r.encryptedValue})
}

// Masked decrypts and redacts the value using byte-based prefix and suffix lengths.
func (r Record) Masked() (string, error) {
	value, err := r.Reveal()
	if err != nil {
		return "", err
	}
	return maskKeyValue(value), nil
}

// List returns records in file order without decrypting their values.
func (Store) List() ([]Record, error) {
	storedRecords, err := load()
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0, len(storedRecords))
	for _, record := range storedRecords {
		records = append(records, recordFromStored(record))
	}
	return records, nil
}

// Get returns the first matching record without decrypting its value.
func (Store) Get(id string) (Record, error) {
	records, err := load()
	if err != nil {
		return Record{}, err
	}
	index := indexOf(records, id)
	if index == -1 {
		return Record{}, errRecordNotFound
	}
	return recordFromStored(records[index]), nil
}

// Add trims frontend-validated input and saves an encrypted key unless it is a duplicate.
func (Store) Add(provider string, name string, value string) error {
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

	records, err := load()
	if err != nil {
		return err
	}

	if err := ensureUniqueKeyRecord(records, "", provider, name, value); err != nil {
		return err
	}

	encryptedValue, err := seal(value)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Format(time.RFC3339)
	records = append(records, storedKeyRecord{
		ID:             strconv.FormatInt(now.UnixNano(), 10),
		Provider:       provider,
		Name:           name,
		EncryptedValue: encryptedValue,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
	})

	return save(records)
}

// Update changes the first matching record. Empty values preserve the ciphertext;
// unchanged metadata with an empty value returns without decryption or writing.
func (Store) Update(id string, provider string, name string, value string) error {
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

	records, err := load()
	if err != nil {
		return err
	}

	recordIndex := indexOf(records, id)
	if recordIndex == -1 {
		return errRecordNotFound
	}
	currentRecord := records[recordIndex]

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
		encryptedValue, err := seal(effectiveValue)
		if err != nil {
			return err
		}
		currentRecord.EncryptedValue = encryptedValue
	}

	currentRecord.Provider = provider
	currentRecord.Name = name
	currentRecord.UpdatedAt = time.Now().Format(time.RFC3339)
	records[recordIndex] = currentRecord

	return save(records)
}

// Delete removes every record matching the ID, without decrypting any values.
func (Store) Delete(id string) error {
	records, err := load()
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
		return errRecordNotFound
	}

	return save(nextRecords)
}

func indexOf(records []storedKeyRecord, id string) int {
	for index, record := range records {
		if record.ID == id {
			return index
		}
	}
	return -1
}

func seal(value string) (string, error) {
	protectedValue, err := protectKeyValue(value)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(protectedValue), nil
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
