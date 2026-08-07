package domain

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

const (
	maxSkills        = 80
	maxSkillLength   = 60
	maxSummaryLength = 2400
)

var githubUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

type Profile struct {
	Name           string        `json:"name"`
	Headline       string        `json:"headline"`
	Email          string        `json:"email"`
	Phone          string        `json:"phone"`
	Location       string        `json:"location"`
	Website        string        `json:"website"`
	GitHubUsername string        `json:"githubUsername"`
	LinkedInURL    string        `json:"linkedInUrl"`
	Summary        string        `json:"summary"`
	Skills         []string      `json:"skills"`
	ContactLinks   []ContactLink `json:"contactLinks"`
	UpdatedAt      string        `json:"updatedAt"`
}

type ProfileInput struct {
	Name           string             `json:"name"`
	Headline       string             `json:"headline"`
	Email          string             `json:"email"`
	Phone          string             `json:"phone"`
	Location       string             `json:"location"`
	Website        string             `json:"website"`
	GitHubUsername string             `json:"githubUsername"`
	LinkedInURL    string             `json:"linkedInUrl"`
	Summary        string             `json:"summary"`
	Skills         []string           `json:"skills"`
	ContactLinks   []ContactLinkInput `json:"contactLinks"`
}

type ContactLink struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	URL       string `json:"url"`
	Position  int    `json:"position"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
type ContactLinkInput struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

func (input ProfileInput) Validate() (Profile, error) {
	profile := Profile{
		Name:           strings.TrimSpace(input.Name),
		Headline:       strings.TrimSpace(input.Headline),
		Email:          strings.TrimSpace(input.Email),
		Phone:          strings.TrimSpace(input.Phone),
		Location:       strings.TrimSpace(input.Location),
		Website:        strings.TrimSpace(input.Website),
		GitHubUsername: strings.TrimSpace(input.GitHubUsername),
		LinkedInURL:    strings.TrimSpace(input.LinkedInURL),
		Summary:        strings.TrimSpace(input.Summary),
		Skills:         normalizeSkills(input.Skills),
		ContactLinks:   make([]ContactLink, 0, len(input.ContactLinks)),
	}

	if len(profile.Name) > 120 {
		return Profile{}, fmt.Errorf("name must be 120 characters or fewer")
	}
	if len(profile.Headline) > 180 {
		return Profile{}, fmt.Errorf("headline must be 180 characters or fewer")
	}
	if len(profile.Summary) > maxSummaryLength {
		return Profile{}, fmt.Errorf("summary must be %d characters or fewer", maxSummaryLength)
	}
	if profile.Email != "" {
		address, err := mail.ParseAddress(profile.Email)
		if err != nil || !strings.EqualFold(address.Address, profile.Email) {
			return Profile{}, fmt.Errorf("email must be a valid email address")
		}
	}
	if err := validateHTTPURL("website", profile.Website); err != nil {
		return Profile{}, err
	}
	if err := validateHTTPURL("LinkedIn URL", profile.LinkedInURL); err != nil {
		return Profile{}, err
	}
	if profile.GitHubUsername != "" && !githubUsernamePattern.MatchString(profile.GitHubUsername) {
		return Profile{}, fmt.Errorf("GitHub username is not valid")
	}
	if len(profile.Skills) > maxSkills {
		return Profile{}, fmt.Errorf("a profile can contain at most %d skills", maxSkills)
	}
	for _, skill := range profile.Skills {
		if len(skill) > maxSkillLength {
			return Profile{}, fmt.Errorf("skill %q must be %d characters or fewer", skill, maxSkillLength)
		}
	}
	if len(input.ContactLinks) > 20 {
		return Profile{}, fmt.Errorf("a profile can contain at most 20 additional links")
	}
	seenLinks := make(map[string]struct{}, len(input.ContactLinks))
	for position, source := range input.ContactLinks {
		link := ContactLink{ID: strings.TrimSpace(source.ID), Label: strings.Join(strings.Fields(source.Label), " "), URL: strings.TrimSpace(source.URL), Position: position}
		if link.Label == "" || len(link.Label) > 80 {
			return Profile{}, fmt.Errorf("contact link label is required and must be 80 characters or fewer")
		}
		if err := validateHTTPURL("contact link", link.URL); err != nil {
			return Profile{}, err
		}
		key := strings.ToLower(link.Label + "\x00" + link.URL)
		if _, exists := seenLinks[key]; exists {
			return Profile{}, fmt.Errorf("contact links must be unique")
		}
		seenLinks[key] = struct{}{}
		profile.ContactLinks = append(profile.ContactLinks, link)
	}

	return profile, nil
}

func normalizeSkills(skills []string) []string {
	result := make([]string, 0, len(skills))
	seen := make(map[string]struct{}, len(skills))
	for _, skill := range skills {
		skill = strings.Join(strings.Fields(skill), " ")
		if skill == "" {
			continue
		}
		key := strings.ToLower(skill)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, skill)
	}
	return result
}

func validateHTTPURL(field, value string) error {
	if value == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be a valid http or https URL", field)
	}
	return nil
}
