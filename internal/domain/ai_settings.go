package domain

import (
	"fmt"
	"strings"
)

type AISettings struct {
	Provider       string `json:"provider"`
	OllamaEndpoint string `json:"ollamaEndpoint"`
	OllamaModel    string `json:"ollamaModel"`
	GeminiModel    string `json:"geminiModel"`
}

func DefaultAISettings() AISettings {
	return AISettings{Provider: "ollama", OllamaEndpoint: DefaultOllamaEndpoint}
}

func (settings AISettings) Validate() (AISettings, error) {
	settings.Provider = strings.ToLower(strings.TrimSpace(settings.Provider))
	if settings.Provider == "" {
		settings.Provider = "ollama"
	}
	if settings.Provider != "ollama" && settings.Provider != "gemini" {
		return AISettings{}, fmt.Errorf("AI provider %q is not supported", settings.Provider)
	}
	endpoint, err := ValidateOllamaEndpoint(settings.OllamaEndpoint)
	if err != nil {
		return AISettings{}, err
	}
	settings.OllamaEndpoint = endpoint
	settings.OllamaModel = strings.TrimSpace(settings.OllamaModel)
	settings.GeminiModel = strings.TrimSpace(strings.TrimPrefix(settings.GeminiModel, "models/"))
	if len(settings.OllamaModel) > 200 || len(settings.GeminiModel) > 200 {
		return AISettings{}, fmt.Errorf("AI model name is too long")
	}
	return settings, nil
}

type CredentialStatus struct {
	Configured bool   `json:"configured"`
	Message    string `json:"message"`
}
