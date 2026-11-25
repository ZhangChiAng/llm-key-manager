package main

import (
	"context"

	"llm-key-manager/internal/keystore"
	"llm-key-manager/internal/probe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App manages backend state and exposes methods to the Wails frontend.
type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	store  keystore.Store
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

// KeyProbeResult contains only a safe, transient conclusion.
// Available means the preset model generated a response at the time of the probe.
type KeyProbeResult struct {
	Status string `json:"status"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
}

// shutdown cancels any outstanding probes when the desktop application exits.
func (a *App) shutdown(_ context.Context) {
	a.cancel()
}

// SaveKey stores an encrypted API key after the frontend validates required fields.
func (a *App) SaveKey(provider string, name string, value string) error {
	return a.store.Add(provider, name, value)
}

// ListKeys returns all locally saved API keys with masked values only.
func (a *App) ListKeys() ([]KeyRecord, error) {
	storedRecords, err := a.store.List()
	if err != nil {
		return nil, err
	}

	records := make([]KeyRecord, 0, len(storedRecords))
	for _, record := range storedRecords {
		value, err := record.Masked()
		if err != nil {
			return nil, err
		}

		records = append(records, KeyRecord{
			ID:          record.ID,
			Provider:    record.Provider,
			Name:        record.Name,
			MaskedValue: value,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		})
	}

	return records, nil
}

// UpdateKey updates frontend-validated metadata and optionally replaces the encrypted key value.
// An empty value preserves the existing encrypted key.
func (a *App) UpdateKey(id string, provider string, name string, value string) error {
	return a.store.Update(id, provider, name, value)
}

// CopyKey decrypts the saved API key and writes the plaintext to the clipboard.
func (a *App) CopyKey(id string) error {
	record, err := a.store.Get(id)
	if err != nil {
		return err
	}
	value, err := record.Reveal()
	if err != nil {
		return err
	}
	return runtime.ClipboardSetText(a.ctx, value)
}

// DeleteKey removes all saved API keys matching the provided record ID.
func (a *App) DeleteKey(id string) error {
	return a.store.Delete(id)
}

// ListProviderPresets supplies the fixed provider names used by suggestions and probes.
func (a *App) ListProviderPresets() []string {
	return probe.ProviderNames()
}

// ProbeKey decrypts a saved key only in the backend and makes one billable request.
// It never persists results or returns response bodies, headers, or network errors.
func (a *App) ProbeKey(id string) (KeyProbeResult, error) {
	result := KeyProbeResult{Status: string(probe.Failed)}
	record, err := a.store.Get(id)
	if err != nil {
		return result, err
	}
	// Match metadata first: unsupported providers must not attempt decryption.
	if !probe.Supports(record.Provider) {
		result.Status = string(probe.Unsupported)
		return result, nil
	}
	value, err := record.Reveal()
	if err != nil {
		return result, err
	}
	result.Status = string(probe.Check(a.ctx, record.Provider, value))
	return result, nil
}
