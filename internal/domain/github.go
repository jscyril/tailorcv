package domain

import (
	"net/url"
	"strings"
)

type GitHubRepository struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"htmlUrl"`
	Homepage    string   `json:"homepage"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	Fork        bool     `json:"fork"`
	Archived    bool     `json:"archived"`
}

type GitHubImportResult struct {
	Fetched  int `json:"fetched"`
	Imported int `json:"imported"`
	Updated  int `json:"updated"`
	Skipped  int `json:"skipped"`
}

func (repository GitHubRepository) Project(existing *Project) (Project, error) {
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
		Name:           repository.Name,
		Description:    repository.Description,
		URL:            homepage,
		RepositoryURL:  repository.HTMLURL,
		Provenance:     ProvenanceGitHub,
		Verification:   VerificationUnverified,
		ResumeEligible: false,
		Skills:         skills,
	}
	if existing != nil {
		input.ID = existing.ID
		input.Role = existing.Role
		if input.URL == "" {
			input.URL = existing.URL
		}
		input.StartDate = existing.StartDate
		input.EndDate = existing.EndDate
		input.Ongoing = existing.Ongoing
		input.Verification = existing.Verification
		input.ResumeEligible = existing.ResumeEligible
		input.Skills = append(input.Skills, existing.Skills...)
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
