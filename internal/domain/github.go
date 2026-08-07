package domain

import (
	"net/url"
	"strings"
)

type GitHubRepository struct {
	ID                int64                `json:"id"`
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	HTMLURL           string               `json:"htmlUrl"`
	Homepage          string               `json:"homepage"`
	Language          string               `json:"language"`
	Languages         []RepositoryLanguage `json:"languages"`
	LanguagesComplete bool                 `json:"languagesComplete"`
	Topics            []string             `json:"topics"`
	Fork              bool                 `json:"fork"`
	Archived          bool                 `json:"archived"`
	Visibility        string               `json:"visibility"`
	UpdatedAt         string               `json:"updatedAt"`
	Readme            string               `json:"readme"`
	ReadmeComplete    bool                 `json:"readmeComplete"`
}

type GitHubImportResult struct {
	Fetched           int `json:"fetched"`
	Imported          int `json:"imported"`
	Updated           int `json:"updated"`
	Skipped           int `json:"skipped"`
	LanguageFallbacks int `json:"languageFallbacks"`
	ReadmeFallbacks   int `json:"readmeFallbacks"`
}

func (repository GitHubRepository) Project(existing *Project) (Project, error) {
	detectedLanguages := repository.Languages
	if len(detectedLanguages) == 0 && strings.TrimSpace(repository.Language) != "" {
		detectedLanguages = []RepositoryLanguage{{Name: repository.Language}}
	}
	skills := make([]string, 0, len(repository.Topics)+1)
	if strings.TrimSpace(repository.Language) != "" {
		skills = append(skills, repository.Language)
	}
	skills = append(skills, repository.Topics...)
	homepage := strings.TrimSpace(repository.Homepage)
	if !isHTTPURL(homepage) {
		homepage = ""
	}
	input := ProjectInput{
		Name:                 repository.Name,
		Description:          repository.Description,
		URL:                  homepage,
		RepositoryURL:        repository.HTMLURL,
		RepositoryID:         repository.ID,
		RepositoryReadme:     repository.Readme,
		RepositoryVisibility: repository.Visibility,
		RepositoryUpdatedAt:  repository.UpdatedAt,
		Provenance:           ProvenanceGitHub,
		Verification:         VerificationUnverified,
		ResumeEligible:       false,
		Skills:               skills,
		DetectedLanguages:    detectedLanguages,
	}
	if existing != nil {
		input.ID = existing.ID
		input.Role = existing.Role
		if !repository.ReadmeComplete {
			input.RepositoryReadme = existing.RepositoryReadme
		}
		if input.URL == "" {
			input.URL = existing.URL
		}
		input.StartDate = existing.StartDate
		input.EndDate = existing.EndDate
		input.Ongoing = existing.Ongoing
		input.Verification = existing.Verification
		input.ResumeEligible = existing.ResumeEligible && existing.Verification == VerificationVerified
		input.Skills = existing.Skills
		input.Bullets = make([]EvidenceBulletInput, len(existing.Bullets))
		for index, bullet := range existing.Bullets {
			input.Bullets[index] = EvidenceBulletInput{ID: bullet.ID, Text: bullet.Text, Provenance: bullet.Provenance, SourceURL: bullet.SourceURL, Verification: bullet.Verification}
		}
	}
	return input.Validate()
}

func isHTTPURL(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}
