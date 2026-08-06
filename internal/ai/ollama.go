package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/domain"
)

const maxOllamaResponseSize = 4 << 20

type Ollama struct {
	endpoint string
	client   *http.Client
}

func NewOllama(endpoint string, client *http.Client) (*Ollama, error) {
	validated, err := domain.ValidateOllamaEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	return &Ollama{endpoint: validated, client: client}, nil
}

func (provider *Ollama) Name() string { return "ollama" }

func (provider *Ollama) Models(ctx context.Context) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("create Ollama model request: %w", err)
	}
	response, err := provider.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("connect to Ollama: %w", err)
	}
	defer response.Body.Close()
	data, err := readLimited(response.Body, "Ollama")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, ollamaStatusError(response.Status, data)
	}
	var payload struct {
		Models []struct {
			Name  string `json:"name"`
			Model string `json:"model"`
		} `json:"models"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode Ollama models: %w", err)
	}
	models := make([]string, 0, len(payload.Models))
	seen := make(map[string]struct{}, len(payload.Models))
	for _, item := range payload.Models {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = strings.TrimSpace(item.Model)
		}
		if name == "" {
			continue
		}
		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			models = append(models, name)
		}
	}
	sort.Strings(models)
	return models, nil
}

func (provider *Ollama) Generate(ctx context.Context, model string, input Request) ([]byte, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("select an Ollama model")
	}
	prompt, err := Prompt(input)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(map[string]any{
		"model": model, "prompt": prompt, "stream": false, "format": json.RawMessage(OutputSchema()),
		"options": map[string]any{"temperature": 0},
	})
	if err != nil {
		return nil, fmt.Errorf("encode Ollama generation request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.endpoint+"/api/generate", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create Ollama generation request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := provider.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("generate with Ollama: %w", err)
	}
	defer response.Body.Close()
	data, err := readLimited(response.Body, "Ollama")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, ollamaStatusError(response.Status, data)
	}
	var result struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode Ollama generation response: %w", err)
	}
	if strings.TrimSpace(result.Error) != "" {
		return nil, fmt.Errorf("Ollama generation failed: %s", strings.TrimSpace(result.Error))
	}
	if len(result.Response) == 0 || len(result.Response) > maxOllamaResponseSize {
		return nil, fmt.Errorf("Ollama returned an empty or oversized structured response")
	}
	return []byte(result.Response), nil
}

func readLimited(reader io.Reader, provider string) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxOllamaResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", provider, err)
	}
	if len(data) > maxOllamaResponseSize {
		return nil, fmt.Errorf("%s response exceeded the 4 MiB size limit", provider)
	}
	return data, nil
}

func ollamaStatusError(status string, data []byte) error {
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(data, &payload)
	message := strings.TrimSpace(payload.Error)
	if message == "" {
		message = strings.TrimSpace(string(data))
	}
	if len(message) > 500 {
		message = message[:500]
	}
	if message == "" {
		return fmt.Errorf("Ollama returned %s", status)
	}
	return fmt.Errorf("Ollama returned %s: %s", status, message)
}
