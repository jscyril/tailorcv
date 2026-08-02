package domain

import (
	"fmt"
	"strings"
)

const maxProjectDescriptionLength = 2400

type Project struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Role           string            `json:"role"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	RepositoryURL  string            `json:"repositoryUrl"`
	StartDate      string            `json:"startDate"`
	EndDate        string            `json:"endDate"`
	Ongoing        bool              `json:"ongoing"`
	Provenance     Provenance        `json:"provenance"`
	Verification   VerificationState `json:"verification"`
	ResumeEligible bool              `json:"resumeEligible"`
	Position       int               `json:"position"`
	Skills         []string          `json:"skills"`
	Bullets        []EvidenceBullet  `json:"bullets"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

type ProjectInput struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Role           string                `json:"role"`
	Description    string                `json:"description"`
	URL            string                `json:"url"`
	RepositoryURL  string                `json:"repositoryUrl"`
	StartDate      string                `json:"startDate"`
	EndDate        string                `json:"endDate"`
	Ongoing        bool                  `json:"ongoing"`
	Provenance     Provenance            `json:"provenance"`
	Verification   VerificationState     `json:"verification"`
	ResumeEligible bool                  `json:"resumeEligible"`
	Skills         []string              `json:"skills"`
	Bullets        []EvidenceBulletInput `json:"bullets"`
}

func (input ProjectInput) Validate() (Project, error) {
	project := Project{
		ID:             strings.TrimSpace(input.ID),
		Name:           strings.Join(strings.Fields(input.Name), " "),
		Role:           strings.Join(strings.Fields(input.Role), " "),
		Description:    strings.TrimSpace(input.Description),
		URL:            strings.TrimSpace(input.URL),
		RepositoryURL:  strings.TrimSpace(input.RepositoryURL),
		StartDate:      strings.TrimSpace(input.StartDate),
		EndDate:        strings.TrimSpace(input.EndDate),
		Ongoing:        input.Ongoing,
		Provenance:     input.Provenance,
		Verification:   input.Verification,
		ResumeEligible: input.ResumeEligible,
		Skills:         normalizeSkills(input.Skills),
		Bullets:        make([]EvidenceBullet, 0, len(input.Bullets)),
	}

	if project.Name == "" {
		return Project{}, fmt.Errorf("project name is required")
	}
	if len(project.Name) > 180 {
		return Project{}, fmt.Errorf("project name must be 180 characters or fewer")
	}
	if len(project.Role) > 180 {
		return Project{}, fmt.Errorf("project role must be 180 characters or fewer")
	}
	if len(project.Description) > maxProjectDescriptionLength {
		return Project{}, fmt.Errorf("project description must be %d characters or fewer", maxProjectDescriptionLength)
	}
	if err := validateEvidenceURL(project.URL); err != nil {
		return Project{}, fmt.Errorf("project URL: %w", err)
	}
	if err := validateEvidenceURL(project.RepositoryURL); err != nil {
		return Project{}, fmt.Errorf("repository URL: %w", err)
	}
	if err := validateMonth("start date", project.StartDate, false); err != nil {
		return Project{}, err
	}
	if project.Ongoing {
		project.EndDate = ""
	} else if err := validateMonth("end date", project.EndDate, false); err != nil {
		return Project{}, err
	}
	if project.StartDate != "" && project.EndDate != "" && project.EndDate < project.StartDate {
		return Project{}, fmt.Errorf("end date cannot be before start date")
	}
	if project.Provenance == "" {
		project.Provenance = ProvenanceManual
	}
	if project.Provenance != ProvenanceManual && project.Provenance != ProvenanceGitHub && project.Provenance != ProvenanceImported {
		return Project{}, fmt.Errorf("project provenance is not valid")
	}
	if project.Verification == "" {
		project.Verification = VerificationUnverified
	}
	if project.Verification != VerificationUnverified && project.Verification != VerificationVerified {
		return Project{}, fmt.Errorf("project verification state is not valid")
	}
	if len(project.Skills) > maxSkills {
		return Project{}, fmt.Errorf("a project can contain at most %d skills", maxSkills)
	}
	for _, skill := range project.Skills {
		if len(skill) > maxSkillLength {
			return Project{}, fmt.Errorf("skill %q must be %d characters or fewer", skill, maxSkillLength)
		}
	}
	if len(input.Bullets) > maxEvidenceBullets {
		return Project{}, fmt.Errorf("a project can contain at most %d evidence bullets", maxEvidenceBullets)
	}

	seenIDs := make(map[string]struct{}, len(input.Bullets))
	for position, bulletInput := range input.Bullets {
		bullet, err := bulletInput.validate(position)
		if err != nil {
			return Project{}, fmt.Errorf("evidence bullet %d: %w", position+1, err)
		}
		if bullet.ID != "" {
			if _, exists := seenIDs[bullet.ID]; exists {
				return Project{}, fmt.Errorf("evidence bullet IDs must be unique")
			}
			seenIDs[bullet.ID] = struct{}{}
		}
		project.Bullets = append(project.Bullets, bullet)
	}
	return project, nil
}
