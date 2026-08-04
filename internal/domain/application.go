package domain

import (
	"fmt"
	"strings"
)

const maxSelectedFacts = 50

type Application struct {
	ID              string          `json:"id"`
	JobID           string          `json:"jobId"`
	Status          string          `json:"status"`
	SelectedFactIDs []string        `json:"selectedFactIds"`
	Versions        []ResumeVersion `json:"versions"`
	CreatedAt       string          `json:"createdAt"`
	UpdatedAt       string          `json:"updatedAt"`
}

type ResumeVersion struct {
	ID                     string   `json:"id"`
	ApplicationID          string   `json:"applicationId"`
	VersionNumber          int      `json:"versionNumber"`
	JobDescriptionSnapshot string   `json:"jobDescriptionSnapshot"`
	SelectedFactIDs        []string `json:"selectedFactIds"`
	LatexSource            string   `json:"latexSource"`
	TemplateID             string   `json:"templateId"`
	CreatedAt              string   `json:"createdAt"`
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
