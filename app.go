package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// App manages backend state and exposes methods to the Wails frontend.
type App struct {
	ctx context.Context
}

// KeyRecord describes a plaintext API key saved in the local key store.
type KeyRecord struct {
	// ID is a generated unique identifier based on the creation timestamp.
	ID string `json:"id"`
	// Provider names the API provider this key belongs to.
	Provider string `json:"provider"`
	// Name is the user-facing label for the key.
	Name string `json:"name"`
	// Value contains the plaintext API key content.
	Value string `json:"value"`
	// CreatedAt records when the key was saved in RFC3339 format.
	CreatedAt string `json:"createdAt"`
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

// SaveKey validates and stores a plaintext API key in the local key store.
func (a *App) SaveKey(provider string, name string, value string) error {
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)

	if provider == "" || name == "" || value == "" {
		return errors.New("provider, name, and key content are required")
	}

	records, err := a.ListKeys()
	if err != nil {
		return err
	}

	now := time.Now()
	records = append(records, KeyRecord{
		ID:        strconv.FormatInt(now.UnixNano(), 10),
		Provider:  provider,
		Name:      name,
		Value:     value,
		CreatedAt: now.Format(time.RFC3339),
	})

	path, err := keyStorePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0666)
}

// ListKeys returns all locally saved API keys from the local key store.
func (a *App) ListKeys() ([]KeyRecord, error) {
	path, err := keyStorePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []KeyRecord{}, nil
	}
	if err != nil {
		return nil, err
	}

	var records []KeyRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	if records == nil {
		return []KeyRecord{}, nil
	}

	return records, nil
}

func keyStorePath() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(executablePath), "keys.json"), nil
}
