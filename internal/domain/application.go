package domain

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

const maxSelectedFacts = 50

type ApplicationStatus string

const (
	ApplicationStatusDraft        ApplicationStatus = "draft"
	ApplicationStatusSubmitted    ApplicationStatus = "submitted"
	ApplicationStatusInterviewing ApplicationStatus = "interviewing"
	ApplicationStatusOffer        ApplicationStatus = "offer"
	ApplicationStatusRejected     ApplicationStatus = "rejected"
	ApplicationStatusArchived     ApplicationStatus = "archived"
)

func (status ApplicationStatus) valid() bool {
	switch status {
	case ApplicationStatusDraft, ApplicationStatusSubmitted, ApplicationStatusInterviewing, ApplicationStatusOffer, ApplicationStatusRejected, ApplicationStatusArchived:
		return true
	}
	return false
}

type Application struct {
	ID              string            `json:"id"`
	JobID           string            `json:"jobId"`
	Name            string            `json:"name"`
	Status          ApplicationStatus `json:"status"`
	AppliedOn       string            `json:"appliedOn,omitempty"`
	FollowUpOn      string            `json:"followUpOn,omitempty"`
	PostingURL      string            `json:"postingUrl,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	CoverLetter     string            `json:"coverLetter,omitempty"`
	SelectedFactIDs []string          `json:"selectedFactIds"`
	Versions        []ResumeVersion   `json:"versions"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

const maxResumeNameLength = 120

// NormalizeResumeName collapses whitespace and enforces the stored length limit.
func NormalizeResumeName(name string) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return "", fmt.Errorf("resume name is required")
	}
	if len(name) > maxResumeNameLength {
		return "", fmt.Errorf("resume name must be %d characters or fewer", maxResumeNameLength)
	}
	return name, nil
}

// CreateResumeInput saves edited or template-rendered source as version 1 of a
// new resume that is not tied to a job.
type CreateResumeInput struct {
	Name        string `json:"name"`
	TemplateID  string `json:"templateId"`
	LatexSource string `json:"latexSource"`
}

func (input CreateResumeInput) Validate() (CreateResumeInput, error) {
	name, err := NormalizeResumeName(input.Name)
	if err != nil {
		return CreateResumeInput{}, err
	}
	input.Name = name
	input.TemplateID = strings.TrimSpace(input.TemplateID)
	if input.TemplateID == "" {
		return CreateResumeInput{}, fmt.Errorf("select a resume template")
	}
	if err := validateResumeSource(input.LatexSource); err != nil {
		return CreateResumeInput{}, err
	}
	return input, nil
}

type RenameResumeInput struct {
	ApplicationID string `json:"applicationId"`
	Name          string `json:"name"`
}

func validateResumeSource(source string) error {
	if len(source) == 0 || len(source) > MaxTemplateSourceBytes {
		return fmt.Errorf("LaTeX source is empty or exceeds the 1 MiB size limit")
	}
	if strings.IndexByte(source, 0) >= 0 {
		return fmt.Errorf("LaTeX source contains a null byte")
	}
	return nil
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
	if !input.Status.valid() {
		return UpdateApplicationStatusInput{}, fmt.Errorf("application status is not valid")
	}
	return input, nil
}

const maxApplicationNotesLength = 10000

const maxCoverLetterLength = 20000

type SaveCoverLetterInput struct {
	ApplicationID string `json:"applicationId"`
	Text          string `json:"text"`
}

func (input SaveCoverLetterInput) Validate() (SaveCoverLetterInput, error) {
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	input.Text = strings.TrimSpace(strings.ReplaceAll(input.Text, "\r\n", "\n"))
	if input.ApplicationID == "" {
		return SaveCoverLetterInput{}, fmt.Errorf("application ID is required")
	}
	if len(input.Text) > maxCoverLetterLength || strings.IndexByte(input.Text, 0) >= 0 {
		return SaveCoverLetterInput{}, fmt.Errorf("cover letter must be %d characters or fewer", maxCoverLetterLength)
	}
	return input, nil
}

// ComposeCoverLetter wraps body paragraphs in a greeting, an opening that
// names the role, and a closing signed with the profile name. The body is the
// only part that carries claims, so it must come from evidence.
func ComposeCoverLetter(profile Profile, job Job, body []string) string {
	greeting := "Dear hiring team,"
	if job.Company != "" {
		greeting = "Dear " + job.Company + " hiring team,"
	}
	opening := "I am writing to apply for "
	if job.Role != "" {
		opening += "the " + job.Role + " role"
	} else {
		opening += "this role"
	}
	if job.Company != "" {
		opening += " at " + job.Company
	}
	opening += "."
	parts := []string{greeting, opening}
	for _, paragraph := range body {
		if paragraph = strings.TrimSpace(paragraph); paragraph != "" {
			parts = append(parts, paragraph)
		}
	}
	parts = append(parts, "Thank you for your time and consideration. I would welcome the chance to discuss how my experience fits your team.", strings.TrimSpace("Sincerely,\n"+profile.Name))
	return strings.Join(parts, "\n\n")
}

// UpdateApplicationDetailsInput holds the tracking fields of an application.
// Dates use YYYY-MM-DD and may be empty.
type UpdateApplicationDetailsInput struct {
	ApplicationID string `json:"applicationId"`
	AppliedOn     string `json:"appliedOn"`
	FollowUpOn    string `json:"followUpOn"`
	PostingURL    string `json:"postingUrl"`
	Notes         string `json:"notes"`
}

func (input UpdateApplicationDetailsInput) Validate() (UpdateApplicationDetailsInput, error) {
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	input.AppliedOn = strings.TrimSpace(input.AppliedOn)
	input.FollowUpOn = strings.TrimSpace(input.FollowUpOn)
	input.PostingURL = strings.TrimSpace(input.PostingURL)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.ApplicationID == "" {
		return UpdateApplicationDetailsInput{}, fmt.Errorf("application ID is required")
	}
	for field, value := range map[string]string{"applied date": input.AppliedOn, "follow-up date": input.FollowUpOn} {
		if value == "" {
			continue
		}
		if _, err := time.Parse(time.DateOnly, value); err != nil {
			return UpdateApplicationDetailsInput{}, fmt.Errorf("%s must use YYYY-MM-DD format", field)
		}
	}
	if err := validateHTTPURL("job posting link", input.PostingURL); err != nil {
		return UpdateApplicationDetailsInput{}, err
	}
	if len(input.Notes) > maxApplicationNotesLength || strings.IndexByte(input.Notes, 0) >= 0 {
		return UpdateApplicationDetailsInput{}, fmt.Errorf("notes must be %d characters or fewer", maxApplicationNotesLength)
	}
	return input, nil
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
	if err := validateResumeSource(input.LatexSource); err != nil {
		return SaveResumeVersionEditInput{}, err
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
