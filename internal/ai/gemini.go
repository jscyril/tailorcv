package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const defaultGeminiEndpoint = "https://generativelanguage.googleapis.com/v1beta"

// Gemini accepts a documented subset of JSON Schema. Contract validation
// below still enforces text lengths and unique citations after generation.
var geminiOutputSchema = json.RawMessage(`{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "schemaVersion": {"type": "string", "enum": ["tailorcv.ai.tailoring.v1"]},
    "proposals": {
      "type": "array",
      "maxItems": 20,
      "items": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "targetFactId": {"type": "string"},
          "supportingFactIds": {"type": "array", "minItems": 1, "maxItems": 8, "items": {"type": "string"}},
          "text": {"type": "string"}
        },
        "required": ["targetFactId", "supportingFactIds", "text"]
      }
    }
  },
  "required": ["schemaVersion", "proposals"]
}`)

type Gemini struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func NewGemini(apiKey string, client *http.Client) (*Gemini, error) {
	return newGemini(apiKey, defaultGeminiEndpoint, client)
}

func newGemini(apiKey, endpoint string, client *http.Client) (*Gemini, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is not configured")
	}
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	return &Gemini{apiKey: apiKey, endpoint: strings.TrimRight(endpoint, "/"), client: client}, nil
}

func (*Gemini) Name() string { return "gemini" }

func (provider *Gemini) Models(ctx context.Context) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.endpoint+"/models?pageSize=1000", nil)
	if err != nil {
		return nil, fmt.Errorf("create Gemini model request: %w", err)
	}
	provider.authorize(request)
	response, err := provider.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("connect to Gemini: %w", err)
	}
	defer response.Body.Close()
	data, err := readLimited(response.Body, "Gemini")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, geminiStatusError(response.Status, data, provider.apiKey)
	}
	var payload struct {
		Models []struct {
			Name             string   `json:"name"`
			SupportedMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode Gemini models: %w", err)
	}
	models := make([]string, 0, len(payload.Models))
	for _, model := range payload.Models {
		if !contains(model.SupportedMethods, "generateContent") {
			continue
		}
		name := strings.TrimPrefix(strings.TrimSpace(model.Name), "models/")
		if name != "" {
			models = append(models, name)
		}
	}
	sort.Strings(models)
	return models, nil
}

func (provider *Gemini) Generate(ctx context.Context, model string, input Request) ([]byte, error) {
	model = strings.TrimSpace(strings.TrimPrefix(model, "models/"))
	if model == "" {
		return nil, fmt.Errorf("select a Gemini model")
	}
	prompt, err := Prompt(input)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(map[string]any{
		"contents": []any{map[string]any{"parts": []any{map[string]string{"text": prompt}}}},
		"generationConfig": map[string]any{
			"temperature": 0, "responseMimeType": "application/json",
			"responseJsonSchema": json.RawMessage(geminiOutputSchema),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("encode Gemini generation request: %w", err)
	}
	path := provider.endpoint + "/models/" + url.PathEscape(model) + ":generateContent"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create Gemini generation request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	provider.authorize(request)
	response, err := provider.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("generate with Gemini: %w", err)
	}
	defer response.Body.Close()
	data, err := readLimited(response.Body, "Gemini")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, geminiStatusError(response.Status, data, provider.apiKey)
	}
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode Gemini generation response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini returned no generated content")
	}
	text := strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text)
	if text == "" || len(text) > maxOllamaResponseSize {
		return nil, fmt.Errorf("Gemini returned an empty or oversized structured response")
	}
	return []byte(text), nil
}

func (provider *Gemini) authorize(request *http.Request) {
	request.Header.Set("x-goog-api-key", provider.apiKey)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func geminiStatusError(status string, data []byte, apiKey string) error {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(data, &payload)
	message := strings.TrimSpace(payload.Error.Message)
	if apiKey != "" {
		message = strings.ReplaceAll(message, apiKey, "[redacted]")
	}
	if len(message) > 500 {
		message = message[:500]
	}
	if message == "" {
		return fmt.Errorf("Gemini returned %s", status)
	}
	return fmt.Errorf("Gemini returned %s: %s", status, message)
}
