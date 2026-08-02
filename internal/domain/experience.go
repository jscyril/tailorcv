package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	maxExperienceBullets = 30
	maxEvidenceLength    = 1200
)

var monthPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

type Provenance string

const (
	ProvenanceManual   Provenance = "manual"
	ProvenanceGitHub   Provenance = "github"
	ProvenanceImported Provenance = "imported"
)

type VerificationState string

const (
	VerificationUnverified VerificationState = "unverified"
	VerificationVerified   VerificationState = "verified"
)

type EvidenceBullet struct {
	ID           string            `json:"id"`
	Text         string            `json:"text"`
	Provenance   Provenance        `json:"provenance"`
	SourceURL    string            `json:"sourceUrl"`
	Verification VerificationState `json:"verification"`
	Position     int               `json:"position"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

type EvidenceBulletInput struct {
	ID           string            `json:"id"`
	Text         string            `json:"text"`
	Provenance   Provenance        `json:"provenance"`
	SourceURL    string            `json:"sourceUrl"`
	Verification VerificationState `json:"verification"`
}

type Experience struct {
	ID        string           `json:"id"`
	Company   string           `json:"company"`
	Title     string           `json:"title"`
	Location  string           `json:"location"`
	StartDate string           `json:"startDate"`
	EndDate   string           `json:"endDate"`
	Current   bool             `json:"current"`
	Position  int              `json:"position"`
	Bullets   []EvidenceBullet `json:"bullets"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
}

type ExperienceInput struct {
	ID        string                `json:"id"`
	Company   string                `json:"company"`
	Title     string                `json:"title"`
	Location  string                `json:"location"`
	StartDate string                `json:"startDate"`
	EndDate   string                `json:"endDate"`
	Current   bool                  `json:"current"`
	Bullets   []EvidenceBulletInput `json:"bullets"`
}

func (input ExperienceInput) Validate() (Experience, error) {
	experience := Experience{
		ID:        strings.TrimSpace(input.ID),
		Company:   strings.Join(strings.Fields(input.Company), " "),
		Title:     strings.Join(strings.Fields(input.Title), " "),
		Location:  strings.Join(strings.Fields(input.Location), " "),
		StartDate: strings.TrimSpace(input.StartDate),
		EndDate:   strings.TrimSpace(input.EndDate),
		Current:   input.Current,
		Bullets:   make([]EvidenceBullet, 0, len(input.Bullets)),
	}

	if experience.Company == "" {
		return Experience{}, fmt.Errorf("company is required")
	}
	if len(experience.Company) > 180 {
		return Experience{}, fmt.Errorf("company must be 180 characters or fewer")
	}
	if experience.Title == "" {
		return Experience{}, fmt.Errorf("job title is required")
	}
	if len(experience.Title) > 180 {
		return Experience{}, fmt.Errorf("job title must be 180 characters or fewer")
	}
	if len(experience.Location) > 180 {
		return Experience{}, fmt.Errorf("location must be 180 characters or fewer")
	}
	if err := validateMonth("start date", experience.StartDate, true); err != nil {
		return Experience{}, err
	}
	if experience.Current {
		experience.EndDate = ""
	} else if err := validateMonth("end date", experience.EndDate, false); err != nil {
		return Experience{}, err
	}
	if experience.EndDate != "" && experience.EndDate < experience.StartDate {
		return Experience{}, fmt.Errorf("end date cannot be before start date")
	}
	if len(input.Bullets) > maxExperienceBullets {
		return Experience{}, fmt.Errorf("an experience can contain at most %d evidence bullets", maxExperienceBullets)
	}

	seenIDs := make(map[string]struct{}, len(input.Bullets))
	for position, bulletInput := range input.Bullets {
		bullet, err := bulletInput.validate(position)
		if err != nil {
			return Experience{}, fmt.Errorf("evidence bullet %d: %w", position+1, err)
		}
		if bullet.ID != "" {
			if _, exists := seenIDs[bullet.ID]; exists {
				return Experience{}, fmt.Errorf("evidence bullet IDs must be unique")
			}
			seenIDs[bullet.ID] = struct{}{}
		}
		experience.Bullets = append(experience.Bullets, bullet)
	}
	return experience, nil
}

func (input EvidenceBulletInput) validate(position int) (EvidenceBullet, error) {
	bullet := EvidenceBullet{
		ID:           strings.TrimSpace(input.ID),
		Text:         strings.Join(strings.Fields(input.Text), " "),
		Provenance:   input.Provenance,
		SourceURL:    strings.TrimSpace(input.SourceURL),
		Verification: input.Verification,
		Position:     position,
	}
	if bullet.Text == "" {
		return EvidenceBullet{}, fmt.Errorf("text is required")
	}
	if len(bullet.Text) > maxEvidenceLength {
		return EvidenceBullet{}, fmt.Errorf("text must be %d characters or fewer", maxEvidenceLength)
	}
	if bullet.Provenance == "" {
		bullet.Provenance = ProvenanceManual
	}
	if bullet.Provenance != ProvenanceManual && bullet.Provenance != ProvenanceGitHub && bullet.Provenance != ProvenanceImported {
		return EvidenceBullet{}, fmt.Errorf("provenance is not valid")
	}
	if bullet.Verification == "" {
		bullet.Verification = VerificationUnverified
	}
	if bullet.Verification != VerificationUnverified && bullet.Verification != VerificationVerified {
		return EvidenceBullet{}, fmt.Errorf("verification state is not valid")
	}
	if err := validateEvidenceURL(bullet.SourceURL); err != nil {
		return EvidenceBullet{}, err
	}
	return bullet, nil
}

func validateMonth(field, value string, required bool) error {
	if value == "" && !required {
		return nil
	}
	if !monthPattern.MatchString(value) {
		return fmt.Errorf("%s must use YYYY-MM format", field)
	}
	return nil
}

func validateEvidenceURL(value string) error {
	if value == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("source URL must be a valid http or https URL")
	}
	return nil
}
