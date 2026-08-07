package domain

import (
	"fmt"
	"strings"
)

type Certification struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Issuer        string            `json:"issuer"`
	IssueDate     string            `json:"issueDate"`
	ExpiryDate    string            `json:"expiryDate"`
	CredentialID  string            `json:"credentialId"`
	CredentialURL string            `json:"credentialUrl"`
	Description   string            `json:"description"`
	Provenance    Provenance        `json:"provenance"`
	Verification  VerificationState `json:"verification"`
	Position      int               `json:"position"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
}

type CertificationInput struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Issuer        string            `json:"issuer"`
	IssueDate     string            `json:"issueDate"`
	ExpiryDate    string            `json:"expiryDate"`
	CredentialID  string            `json:"credentialId"`
	CredentialURL string            `json:"credentialUrl"`
	Description   string            `json:"description"`
	Provenance    Provenance        `json:"provenance"`
	Verification  VerificationState `json:"verification"`
}

func (input CertificationInput) Validate() (Certification, error) {
	certification := Certification{ID: strings.TrimSpace(input.ID), Name: strings.Join(strings.Fields(input.Name), " "), Issuer: strings.Join(strings.Fields(input.Issuer), " "), IssueDate: strings.TrimSpace(input.IssueDate), ExpiryDate: strings.TrimSpace(input.ExpiryDate), CredentialID: strings.TrimSpace(input.CredentialID), CredentialURL: strings.TrimSpace(input.CredentialURL), Description: strings.TrimSpace(input.Description), Provenance: input.Provenance, Verification: input.Verification}
	if certification.Name == "" || len(certification.Name) > 180 {
		return Certification{}, fmt.Errorf("certification name is required and must be 180 characters or fewer")
	}
	if certification.Issuer == "" || len(certification.Issuer) > 180 {
		return Certification{}, fmt.Errorf("certification issuer is required and must be 180 characters or fewer")
	}
	if err := validateMonth("issue date", certification.IssueDate, false); err != nil {
		return Certification{}, err
	}
	if err := validateMonth("expiry date", certification.ExpiryDate, false); err != nil {
		return Certification{}, err
	}
	if certification.IssueDate != "" && certification.ExpiryDate != "" && certification.ExpiryDate < certification.IssueDate {
		return Certification{}, fmt.Errorf("expiry date cannot be before issue date")
	}
	if len(certification.CredentialID) > 200 || len(certification.Description) > maxEvidenceLength {
		return Certification{}, fmt.Errorf("certification details exceed the size limit")
	}
	if err := validateEvidenceURL(certification.CredentialURL); err != nil {
		return Certification{}, fmt.Errorf("credential URL: %w", err)
	}
	certification.Provenance, certification.Verification = normalizeEvidenceState(certification.Provenance, certification.Verification)
	if !validEvidenceState(certification.Provenance, certification.Verification) {
		return Certification{}, fmt.Errorf("certification evidence state is not valid")
	}
	return certification, nil
}

type Achievement struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Date         string            `json:"date"`
	SourceURL    string            `json:"sourceUrl"`
	Provenance   Provenance        `json:"provenance"`
	Verification VerificationState `json:"verification"`
	Position     int               `json:"position"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

type AchievementInput struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Date         string            `json:"date"`
	SourceURL    string            `json:"sourceUrl"`
	Provenance   Provenance        `json:"provenance"`
	Verification VerificationState `json:"verification"`
}

func (input AchievementInput) Validate() (Achievement, error) {
	achievement := Achievement{ID: strings.TrimSpace(input.ID), Title: strings.Join(strings.Fields(input.Title), " "), Description: strings.TrimSpace(input.Description), Date: strings.TrimSpace(input.Date), SourceURL: strings.TrimSpace(input.SourceURL), Provenance: input.Provenance, Verification: input.Verification}
	if achievement.Title == "" || len(achievement.Title) > 180 {
		return Achievement{}, fmt.Errorf("achievement title is required and must be 180 characters or fewer")
	}
	if achievement.Description == "" || len(achievement.Description) > maxEvidenceLength {
		return Achievement{}, fmt.Errorf("achievement description is required and must be 1200 characters or fewer")
	}
	if err := validateMonth("achievement date", achievement.Date, false); err != nil {
		return Achievement{}, err
	}
	if err := validateEvidenceURL(achievement.SourceURL); err != nil {
		return Achievement{}, err
	}
	achievement.Provenance, achievement.Verification = normalizeEvidenceState(achievement.Provenance, achievement.Verification)
	if !validEvidenceState(achievement.Provenance, achievement.Verification) {
		return Achievement{}, fmt.Errorf("achievement evidence state is not valid")
	}
	return achievement, nil
}

func normalizeEvidenceState(provenance Provenance, verification VerificationState) (Provenance, VerificationState) {
	if provenance == "" {
		provenance = ProvenanceManual
	}
	if verification == "" {
		verification = VerificationUnverified
	}
	return provenance, verification
}

func validEvidenceState(provenance Provenance, verification VerificationState) bool {
	return (provenance == ProvenanceManual || provenance == ProvenanceGitHub || provenance == ProvenanceImported) && (verification == VerificationUnverified || verification == VerificationVerified)
}
