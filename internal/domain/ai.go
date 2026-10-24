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

// GenerateProjectReadmeBulletsInput requests editable, unverified project
// bullets from a README snapshot. Nothing is persisted until the user saves
// the project form.
type GenerateProjectReadmeBulletsInput struct {
	ProjectID string `json:"projectId"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Endpoint  string `json:"endpoint"`
}

func (input GenerateProjectReadmeBulletsInput) Validate() (GenerateProjectReadmeBulletsInput, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return GenerateProjectReadmeBulletsInput{}, fmt.Errorf("save the project before generating README bullets")
	}
	var err error
	if input.Provider, input.Model, input.Endpoint, err = validateProviderChoice(input.Provider, input.Model, input.Endpoint); err != nil {
		return GenerateProjectReadmeBulletsInput{}, err
	}
	return input, nil
}

type ProjectReadmeBulletsResult struct {
	Bullets []string `json:"bullets"`
}

func (input GenerateAITailoringInput) Validate() (GenerateAITailoringInput, error) {
	selection, err := (CreateResumeVersionInput{JobID: input.JobID, TemplateID: "ai-validation", SelectedFactIDs: input.SelectedFactIDs}).Validate()
	if err != nil {
		return GenerateAITailoringInput{}, err
	}
	input.JobID = selection.JobID
	input.SelectedFactIDs = selection.SelectedFactIDs
	if input.Provider, input.Model, input.Endpoint, err = validateProviderChoice(input.Provider, input.Model, input.Endpoint); err != nil {
		return GenerateAITailoringInput{}, err
	}
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

// validateProviderChoice normalizes the provider, model, and endpoint shared
// by every AI request.
func validateProviderChoice(provider, model, endpoint string) (string, string, string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "ollama"
	}
	if provider != "ollama" && provider != "gemini" {
		return "", "", "", fmt.Errorf("AI provider %q is not supported", provider)
	}
	model = strings.TrimSpace(model)
	if model == "" || len(model) > 200 {
		return "", "", "", fmt.Errorf("select a model for %s", provider)
	}
	if provider == "gemini" {
		return provider, strings.TrimPrefix(model, "models/"), "", nil
	}
	endpoint, err := ValidateOllamaEndpoint(endpoint)
	if err != nil {
		return "", "", "", err
	}
	return provider, model, endpoint, nil
}

// GenerateCoverLetterInput asks the configured provider for cover letter body
// paragraphs from an application's selected evidence. Nothing is saved until
// the user saves the letter.
type GenerateCoverLetterInput struct {
	ApplicationID string `json:"applicationId"`
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	Endpoint      string `json:"endpoint"`
}

func (input GenerateCoverLetterInput) Validate() (GenerateCoverLetterInput, error) {
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	if input.ApplicationID == "" {
		return GenerateCoverLetterInput{}, fmt.Errorf("application ID is required")
	}
	var err error
	if input.Provider, input.Model, input.Endpoint, err = validateProviderChoice(input.Provider, input.Model, input.Endpoint); err != nil {
		return GenerateCoverLetterInput{}, err
	}
	return input, nil
}

// FixLatexInput asks an agent to repair LaTeX that fails to compile. The
// whole source is sent to the provider.
type FixLatexInput struct {
	Source   string `json:"source"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Endpoint string `json:"endpoint"`
}

func (input FixLatexInput) Validate() (FixLatexInput, error) {
	if strings.TrimSpace(input.Source) == "" {
		return FixLatexInput{}, fmt.Errorf("there is no LaTeX source to fix")
	}
	if len(input.Source) > 256<<10 {
		return FixLatexInput{}, fmt.Errorf("LaTeX source is too large for an AI fix")
	}
	var err error
	if input.Provider, input.Model, input.Endpoint, err = validateProviderChoice(input.Provider, input.Model, input.Endpoint); err != nil {
		return FixLatexInput{}, err
	}
	return input, nil
}

// LatexFixResult is a proposed source that compiled, returned as an unsaved draft.
type LatexFixResult struct {
	Source string        `json:"source"`
	Result CompileResult `json:"result"`
}

// SuggestEvidenceInput asks an agent to choose evidence for a saved job from
// its top ranked facts.
type SuggestEvidenceInput struct {
	JobID    string `json:"jobId"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Endpoint string `json:"endpoint"`
}

func (input SuggestEvidenceInput) Validate() (SuggestEvidenceInput, error) {
	input.JobID = strings.TrimSpace(input.JobID)
	if input.JobID == "" {
		return SuggestEvidenceInput{}, fmt.Errorf("analyze a job before asking AI to choose evidence")
	}
	var err error
	if input.Provider, input.Model, input.Endpoint, err = validateProviderChoice(input.Provider, input.Model, input.Endpoint); err != nil {
		return SuggestEvidenceInput{}, err
	}
	return input, nil
}

// DraftFollowUpInput asks an agent for a follow-up email about an application.
type DraftFollowUpInput struct {
	ApplicationID string `json:"applicationId"`
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	Endpoint      string `json:"endpoint"`
}

func (input DraftFollowUpInput) Validate() (DraftFollowUpInput, error) {
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	if input.ApplicationID == "" {
		return DraftFollowUpInput{}, fmt.Errorf("choose an application to follow up on")
	}
	var err error
	if input.Provider, input.Model, input.Endpoint, err = validateProviderChoice(input.Provider, input.Model, input.Endpoint); err != nil {
		return DraftFollowUpInput{}, err
	}
	return input, nil
}

// ResumeImportAIInput asks an agent to re-read the last previewed resume file.
type ResumeImportAIInput struct {
	FileName string `json:"fileName"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Endpoint string `json:"endpoint"`
}

func (input ResumeImportAIInput) Validate() (ResumeImportAIInput, error) {
	var err error
	if input.Provider, input.Model, input.Endpoint, err = validateProviderChoice(input.Provider, input.Model, input.Endpoint); err != nil {
		return ResumeImportAIInput{}, err
	}
	return input, nil
}
