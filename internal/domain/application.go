package domain

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const maxSelectedFacts = 50

type ApplicationStatus string

const (
	ApplicationStatusDraft     ApplicationStatus = "draft"
	ApplicationStatusSubmitted ApplicationStatus = "submitted"
	ApplicationStatusArchived  ApplicationStatus = "archived"
)

type Application struct {
	ID              string            `json:"id"`
	JobID           string            `json:"jobId"`
	Status          ApplicationStatus `json:"status"`
	SelectedFactIDs []string          `json:"selectedFactIds"`
	Versions        []ResumeVersion   `json:"versions"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

type UpdateApplicationStatusInput struct {
	ApplicationID string            `json:"applicationId"`
	Status        ApplicationStatus `json:"status"`
}

func (input UpdateApplicationStatusInput) Validate() (UpdateApplicationStatusInput, error) {
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	if input.ApplicationID == "" {
		return UpdateApplicationStatusInput{}, fmt.Errorf("application ID is required")
	}
	switch input.Status {
	case ApplicationStatusDraft, ApplicationStatusSubmitted, ApplicationStatusArchived:
		return input, nil
	default:
		return UpdateApplicationStatusInput{}, fmt.Errorf("application status is not valid")
	}
}

type ResumeVersion struct {
	ID                     string              `json:"id"`
	ApplicationID          string              `json:"applicationId"`
	VersionNumber          int                 `json:"versionNumber"`
	JobDescriptionSnapshot string              `json:"jobDescriptionSnapshot"`
	SelectedFactIDs        []string            `json:"selectedFactIds"`
	LatexSource            string              `json:"latexSource"`
	TemplateID             string              `json:"templateId"`
	RankingExplanations    []EvidenceMatch     `json:"rankingExplanations"`
	ContentHash            string              `json:"contentHash"`
	CompileSuccess         bool                `json:"compileSuccess"`
	CompileEngine          string              `json:"compileEngine"`
	CompileDurationMS      int64               `json:"compileDurationMs"`
	CompileDiagnostics     []CompileDiagnostic `json:"compileDiagnostics"`
	CompiledAt             string              `json:"compiledAt"`
	PDFPath                string              `json:"-"`
	PDFAvailable           bool                `json:"pdfAvailable"`
	CreatedAt              string              `json:"createdAt"`
}

type CreateResumeVersionInput struct {
	JobID           string   `json:"jobId"`
	SelectedFactIDs []string `json:"selectedFactIds"`
	TemplateID      string   `json:"templateId"`
}

type ApplicationResumeResult struct {
	Application Application   `json:"application"`
	Version     ResumeVersion `json:"version"`
}

type SaveResumeVersionEditInput struct {
	ApplicationID string `json:"applicationId"`
	BaseVersionID string `json:"baseVersionId"`
	LatexSource   string `json:"latexSource"`
}

func (input SaveResumeVersionEditInput) Validate() (SaveResumeVersionEditInput, error) {
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	input.BaseVersionID = strings.TrimSpace(input.BaseVersionID)
	if input.ApplicationID == "" || input.BaseVersionID == "" {
		return SaveResumeVersionEditInput{}, fmt.Errorf("open a saved resume version before saving an edit")
	}
	if len(input.LatexSource) == 0 || len(input.LatexSource) > MaxTemplateSourceBytes {
		return SaveResumeVersionEditInput{}, fmt.Errorf("LaTeX source is empty or exceeds the 1 MiB size limit")
	}
	if strings.IndexByte(input.LatexSource, 0) >= 0 {
		return SaveResumeVersionEditInput{}, fmt.Errorf("LaTeX source contains a null byte")
	}
	return input, nil
}

func ResumeContentHash(source string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(source)))
}

func (input CreateResumeVersionInput) Validate() (CreateResumeVersionInput, error) {
	input.JobID = strings.TrimSpace(input.JobID)
	input.TemplateID = strings.TrimSpace(input.TemplateID)
	if input.JobID == "" {
		return CreateResumeVersionInput{}, fmt.Errorf("save and analyze a job before creating a resume version")
	}
	if input.TemplateID == "" {
		return CreateResumeVersionInput{}, fmt.Errorf("select a resume template")
	}
	if len(input.SelectedFactIDs) == 0 {
		return CreateResumeVersionInput{}, fmt.Errorf("select at least one evidence fact")
	}
	if len(input.SelectedFactIDs) > maxSelectedFacts {
		return CreateResumeVersionInput{}, fmt.Errorf("select at most %d evidence facts", maxSelectedFacts)
	}
	selected := make([]string, 0, len(input.SelectedFactIDs))
	seen := make(map[string]struct{}, len(input.SelectedFactIDs))
	for _, source := range input.SelectedFactIDs {
		id := strings.TrimSpace(source)
		if id == "" {
			return CreateResumeVersionInput{}, fmt.Errorf("selected evidence IDs cannot be empty")
		}
		if _, exists := seen[id]; exists {
			return CreateResumeVersionInput{}, fmt.Errorf("selected evidence IDs must be unique")
		}
		seen[id] = struct{}{}
		selected = append(selected, id)
	}
	input.SelectedFactIDs = selected
	return input, nil
}
