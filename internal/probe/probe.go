// Package probe checks API keys against fixed provider presets with one billable request.
package probe

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Status is a safe, transient probe conclusion, never a raw provider response.
type Status string

const (
	// Available means the preset model generated a response at probe time.
	Available Status = "available"
	// AuthenticationFailed means the provider rejected the credentials.
	AuthenticationFailed Status = "authentication_failed"
	// InsufficientBalance means the provider reported exhausted balance or quota.
	InsufficientBalance Status = "insufficient_balance"
	// PermissionDenied means the provider denied access to the requested resource.
	PermissionDenied Status = "permission_denied"
	// RateLimited means the provider reported a rate or concurrency limit.
	RateLimited Status = "rate_limited"
	// Failed covers transport, response-format, and otherwise unclassified failures.
	Failed Status = "failed"
	// Unsupported means no preset matches; no request is made.
	Unsupported Status = "unsupported"
)

type providerPreset struct {
	name            string
	endpoint        string
	model           string
	disableThinking bool
	businessCodes   map[string]Status
}

// Keep names stable so existing saved records continue to match these presets.
// The README lists the same providers and models for users.
var providerPresets = []providerPreset{
	// API reference: https://api-docs.deepseek.com/api/create-chat-completion/
	{
		name: "DeepSeek", endpoint: "https://api.deepseek.com/chat/completions",
		model: "deepseek-flash", disableThinking: true,
	},
	// API reference: https://platform.minimax.cn/docs/api-reference/text-openai-api
	// Business codes: https://platform.minimax.cn/docs/api-reference/errorcode
	{
		name: "MiniMax", endpoint: "https://api.minimax.cn/v1/chat/completions",
		model: "MiniMax-M3", disableThinking: true,
		businessCodes: map[string]Status{
			"1004": AuthenticationFailed,
			"2049": AuthenticationFailed,
			"1008": InsufficientBalance,
			"1002": RateLimited,
			"1041": RateLimited,
			"2045": RateLimited,
			"2056": RateLimited,
		},
	},
	// API reference: https://ofox.ai/docs/api
	{
		name: "OfoxAI", endpoint: "https://api.ofox.io/v1/chat/completions",
		model: "qwen/qwen-flash", disableThinking: false,
	},
}

// ProviderNames returns the fixed provider names in preset order.
func ProviderNames() []string {
	names := make([]string, 0, len(providerPresets))
	for _, preset := range providerPresets {
		names = append(names, preset.name)
	}
	return names
}

// Supports reports whether the trimmed provider name exactly matches a preset.
func Supports(provider string) bool {
	for _, preset := range providerPresets {
		if strings.TrimSpace(provider) == preset.name {
			return true
		}
	}
	return false
}

// Check makes one billable request for supported providers and returns only a safe
// conclusion. Unsupported providers return Unsupported without making a request.
func Check(ctx context.Context, provider string, key string) Status {
	for _, preset := range providerPresets {
		if strings.TrimSpace(provider) == preset.name {
			return probeCompletion(ctx, preset, key)
		}
	}
	return Unsupported
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

// probeCompletion sends one minimal completion request and classifies the response.
// The first matching rule wins, in this order: a transport failure, oversized body, or
// redirect is Failed; then the provider's base_resp business code; then error.code and
// error.type; then the HTTP status; and only then whether a choice holds generated text.
func probeCompletion(parent context.Context, preset providerPreset, value string) Status {
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
		return Failed
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, preset.endpoint, bytes.NewReader(body))
	if err != nil {
		return Failed
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
		return Failed
	}
	defer func() { _ = response.Body.Close() }()
	// Bound untrusted responses; raw content is never logged or returned to Wails.
	const maxResponseBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(data) > maxResponseBytes {
		return Failed
	}
	var decoded probeResponse
	decodeErr := json.Unmarshal(data, &decoded)
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		return Failed
	}
	if decodeErr == nil {
		if preset.businessCodes != nil && decoded.BaseResp != nil && decoded.BaseResp.StatusCode != 0 {
			if status, ok := preset.businessCodes[strconv.Itoa(decoded.BaseResp.StatusCode)]; ok {
				return status
			}
			return Failed
		}
		if decoded.Error != nil {
			for _, code := range []string{strings.Trim(string(decoded.Error.Code), "\""), decoded.Error.Type} {
				if status, ok := preset.businessCodes[code]; ok {
					return status
				}
				switch strings.ToLower(code) {
				case "invalid_api_key", "authentication_error", "authentication_failed":
					return AuthenticationFailed
				case "insufficient_balance", "insufficient_quota":
					return InsufficientBalance
				case "permission_denied", "permission_error":
					return PermissionDenied
				case "rate_limit_exceeded", "rate_limit_error":
					return RateLimited
				}
			}
		}
	}
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return AuthenticationFailed
	case http.StatusPaymentRequired:
		return InsufficientBalance
	case http.StatusForbidden:
		return PermissionDenied
	case http.StatusTooManyRequests:
		return RateLimited
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || decodeErr != nil || decoded.Error != nil {
		return Failed
	}
	for _, choice := range decoded.Choices {
		if choice.Message == nil || choice.Message.Role != "assistant" {
			continue
		}
		if choice.FinishReason == "length" ||
			(choice.FinishReason == "stop" && choice.Message.Content != nil && *choice.Message.Content != "") {
			return Available
		}
	}
	return Failed
}
