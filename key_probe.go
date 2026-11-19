package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type providerPreset struct {
	name            string
	endpoint        string
	model           string
	disableThinking bool
}

// Keep names stable so existing saved records continue to match these presets.
var providerPresets = []providerPreset{
	{"DeepSeek", "https://api.deepseek.com/chat/completions", "deepseek-flash", true},
	{"MiniMax", "https://api.minimax.cn/v1/chat/completions", "MiniMax-M3", true},
	{"OfoxAI", "https://api.ofox.io/v1/chat/completions", "qwen/qwen-flash", false},
}

// KeyProbeResult contains only the record ID and a safe, transient conclusion.
// Available means the preset model generated a response at the time of the probe.
type KeyProbeResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ListProviderPresets supplies the fixed provider names used by suggestions and probes.
func (a *App) ListProviderPresets() []string {
	names := make([]string, 0, len(providerPresets))
	for _, preset := range providerPresets {
		names = append(names, preset.name)
	}
	return names
}

// ProbeKey decrypts a saved key only in the backend and makes one billable request.
// It never persists results or returns response bodies, headers, or network errors.
func (a *App) ProbeKey(id string) (KeyProbeResult, error) {
	result := KeyProbeResult{ID: id, Status: "failed"}
	records, err := readStoredKeys()
	if err != nil {
		return result, err
	}
	for _, record := range records {
		if record.ID != id {
			continue
		}
		for _, preset := range providerPresets {
			if strings.TrimSpace(record.Provider) != preset.name {
				continue
			}
			value, err := decryptStoredValue(record)
			if err != nil {
				return result, err
			}
			result.Status = probeCompletion(a.ctx, preset, value)
			return result, nil
		}
		result.Status = "unsupported"
		return result, nil
	}
	return result, errors.New("key record not found")
}

type probeResponse struct {
	Choices []struct {
		Message *struct {
			Role    string  `json:"role"`
			Content *string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Code json.RawMessage `json:"code"`
		Type string          `json:"type"`
	} `json:"error"`
	BaseResp *struct {
		StatusCode int `json:"status_code"`
	} `json:"base_resp"`
}

func probeCompletion(parent context.Context, preset providerPreset, value string) string {
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	payload := map[string]any{
		"model":      preset.model,
		"messages":   []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":     false,
		"max_tokens": 1,
	}
	if preset.disableThinking {
		payload["thinking"] = map[string]string{"type": "disabled"}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "failed"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, preset.endpoint, bytes.NewReader(body))
	if err != nil {
		return "failed"
	}
	req.Header.Set("Authorization", "Bearer "+value)
	req.Header.Set("Content-Type", "application/json")
	// POST is not idempotent: the standard transport does not retry this request.
	// Refuse redirects to keep credentials and API usage at the configured endpoint.
	client := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	response, err := client.Do(req)
	if err != nil {
		return "failed"
	}
	defer func() { _ = response.Body.Close() }()
	// Bound untrusted responses; raw content is never logged or returned to Wails.
	const maxResponseBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(data) > maxResponseBytes {
		return "failed"
	}
	var decoded probeResponse
	decodeErr := json.Unmarshal(data, &decoded)
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		return "failed"
	}
	if decodeErr == nil {
		if preset.name == "MiniMax" && decoded.BaseResp != nil && decoded.BaseResp.StatusCode != 0 {
			return minimaxProbeStatus(strconv.Itoa(decoded.BaseResp.StatusCode))
		}
		if decoded.Error != nil {
			for _, code := range []string{strings.Trim(string(decoded.Error.Code), "\""), decoded.Error.Type} {
				if preset.name == "MiniMax" {
					if status := minimaxProbeStatus(code); status != "failed" {
						return status
					}
				}
				switch strings.ToLower(code) {
				case "invalid_api_key", "authentication_error", "authentication_failed":
					return "authentication_failed"
				case "insufficient_balance", "insufficient_quota":
					return "insufficient_balance"
				case "permission_denied", "permission_error":
					return "permission_denied"
				case "rate_limit_exceeded", "rate_limit_error":
					return "rate_limited"
				}
			}
		}
	}
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return "authentication_failed"
	case http.StatusPaymentRequired:
		return "insufficient_balance"
	case http.StatusForbidden:
		return "permission_denied"
	case http.StatusTooManyRequests:
		return "rate_limited"
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || decodeErr != nil || decoded.Error != nil {
		return "failed"
	}
	for _, choice := range decoded.Choices {
		if choice.Message == nil || choice.Message.Role != "assistant" {
			continue
		}
		if choice.FinishReason == "length" ||
			(choice.FinishReason == "stop" && choice.Message.Content != nil && *choice.Message.Content != "") {
			return "available"
		}
	}
	return "failed"
}

// MiniMax may return business errors in base_resp or the compatible error code.
func minimaxProbeStatus(code string) string {
	switch code {
	case "1004", "2049":
		return "authentication_failed"
	case "1008":
		return "insufficient_balance"
	case "1002", "1041", "2045", "2056":
		return "rate_limited"
	default:
		return "failed"
	}
}
