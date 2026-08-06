package domain

import (
	"fmt"
	"net/url"
	"strings"
)

const DefaultOllamaEndpoint = "http://127.0.0.1:11434"

type AIProposal struct {
	TargetFactID      string   `json:"targetFactId"`
	SupportingFactIDs []string `json:"supportingFactIds"`
	Text              string   `json:"text"`
}

type AIRun struct {
	ID               string       `json:"id"`
	JobID            string       `json:"jobId"`
	Provider         string       `json:"provider"`
	Model            string       `json:"model"`
	PromptVersion    string       `json:"promptVersion"`
	SchemaVersion    string       `json:"schemaVersion"`
	SelectedFactIDs  []string     `json:"selectedFactIds"`
	ValidationPassed bool         `json:"validationPassed"`
	FailureCategory  string       `json:"failureCategory"`
	ValidationErrors []string     `json:"validationErrors"`
	Proposals        []AIProposal `json:"proposals"`
	ResumeVersionID  string       `json:"resumeVersionId"`
	CreatedAt        string       `json:"createdAt"`
	AcceptedAt       string       `json:"acceptedAt"`
}

type GenerateAITailoringInput struct {
	JobID           string   `json:"jobId"`
	SelectedFactIDs []string `json:"selectedFactIds"`
	Provider        string   `json:"provider"`
	Model           string   `json:"model"`
	Endpoint        string   `json:"endpoint"`
}

func (input GenerateAITailoringInput) Validate() (GenerateAITailoringInput, error) {
	selection, err := (CreateResumeVersionInput{JobID: input.JobID, TemplateID: "ai-validation", SelectedFactIDs: input.SelectedFactIDs}).Validate()
	if err != nil {
		return GenerateAITailoringInput{}, err
	}
	input.JobID = selection.JobID
	input.SelectedFactIDs = selection.SelectedFactIDs
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	if input.Provider == "" {
		input.Provider = "ollama"
	}
	if input.Provider != "ollama" {
		return GenerateAITailoringInput{}, fmt.Errorf("AI provider %q is not supported", input.Provider)
	}
	input.Model = strings.TrimSpace(input.Model)
	if input.Model == "" || len(input.Model) > 200 {
		return GenerateAITailoringInput{}, fmt.Errorf("select an Ollama model")
	}
	endpoint, err := ValidateOllamaEndpoint(input.Endpoint)
	if err != nil {
		return GenerateAITailoringInput{}, err
	}
	input.Endpoint = endpoint
	return input, nil
}

type AIProviderStatus struct {
	Provider  string   `json:"provider"`
	Endpoint  string   `json:"endpoint"`
	Available bool     `json:"available"`
	Models    []string `json:"models"`
	Message   string   `json:"message"`
}

type AcceptAITailoringInput struct {
	RunID      string       `json:"runId"`
	TemplateID string       `json:"templateId"`
	Proposals  []AIProposal `json:"proposals"`
}

func (input AcceptAITailoringInput) Validate() (AcceptAITailoringInput, error) {
	input.RunID = strings.TrimSpace(input.RunID)
	input.TemplateID = strings.TrimSpace(input.TemplateID)
	if input.RunID == "" {
		return AcceptAITailoringInput{}, fmt.Errorf("generate and review AI proposals before accepting them")
	}
	if input.TemplateID == "" {
		return AcceptAITailoringInput{}, fmt.Errorf("select a resume template")
	}
	if len(input.Proposals) == 0 {
		return AcceptAITailoringInput{}, fmt.Errorf("accept at least one proposed bullet")
	}
	return input, nil
}

func ValidateOllamaEndpoint(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = DefaultOllamaEndpoint
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("Ollama endpoint must be a valid http or https URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("Ollama endpoint cannot contain credentials, a query, or a fragment")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return strings.TrimRight(parsed.String(), "/"), nil
}
