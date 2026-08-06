package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const ProfileBackupSchemaVersion = 3

type ProfileBackup struct {
	SchemaVersion      int              `json:"schemaVersion"`
	ExportedAt         string           `json:"exportedAt"`
	Profile            Profile          `json:"profile"`
	Experiences        []Experience     `json:"experiences"`
	Projects           []Project        `json:"projects"`
	Educations         []Education      `json:"educations"`
	Jobs               []Job            `json:"jobs,omitempty"`
	Applications       []Application    `json:"applications,omitempty"`
	Templates          []ResumeTemplate `json:"templates,omitempty"`
	SelectedTemplateID string           `json:"selectedTemplateId,omitempty"`
	AIRuns             []AIRun          `json:"aiRuns,omitempty"`
	AISettings         AISettings       `json:"aiSettings"`
}

type BackupResult struct {
	Path               string `json:"path"`
	Cancelled          bool   `json:"cancelled"`
	ExperienceCount    int    `json:"experienceCount"`
	ProjectCount       int    `json:"projectCount"`
	EducationCount     int    `json:"educationCount"`
	JobCount           int    `json:"jobCount"`
	ApplicationCount   int    `json:"applicationCount"`
	ResumeVersionCount int    `json:"resumeVersionCount"`
	TemplateCount      int    `json:"templateCount"`
	AIRunCount         int    `json:"aiRunCount"`
}

func DecodeProfileBackup(data []byte) (ProfileBackup, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var backup ProfileBackup
	if err := decoder.Decode(&backup); err != nil {
		return ProfileBackup{}, fmt.Errorf("decode backup: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return ProfileBackup{}, fmt.Errorf("decode backup: multiple JSON values are not allowed")
		}
		return ProfileBackup{}, fmt.Errorf("decode backup: %w", err)
	}
	return backup.Validate()
}

func (backup ProfileBackup) Validate() (ProfileBackup, error) {
	if backup.SchemaVersion < 1 || backup.SchemaVersion > ProfileBackupSchemaVersion {
		return ProfileBackup{}, fmt.Errorf("backup schema version %d is not supported", backup.SchemaVersion)
	}
	if _, err := time.Parse(time.RFC3339, backup.ExportedAt); err != nil {
		return ProfileBackup{}, fmt.Errorf("backup export time is not valid")
	}

	profile, err := (ProfileInput{
		Name:           backup.Profile.Name,
		Headline:       backup.Profile.Headline,
		Email:          backup.Profile.Email,
		Phone:          backup.Profile.Phone,
		Location:       backup.Profile.Location,
		Website:        backup.Profile.Website,
		GitHubUsername: backup.Profile.GitHubUsername,
		LinkedInURL:    backup.Profile.LinkedInURL,
		Summary:        backup.Profile.Summary,
		Skills:         backup.Profile.Skills,
	}).Validate()
	if err != nil {
		return ProfileBackup{}, fmt.Errorf("profile: %w", err)
	}
	profile.UpdatedAt = backup.Profile.UpdatedAt
	if err := validateBackupTimestamp("profile update time", profile.UpdatedAt, false); err != nil {
		return ProfileBackup{}, err
	}
	backup.Profile = profile

	experiences := make([]Experience, 0, len(backup.Experiences))
	for index, source := range backup.Experiences {
		if source.Position < 0 {
			return ProfileBackup{}, fmt.Errorf("experience %d: position cannot be negative", index+1)
		}
		if err := validateBackupTimestamp(fmt.Sprintf("experience %d creation time", index+1), source.CreatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		if err := validateBackupTimestamp(fmt.Sprintf("experience %d update time", index+1), source.UpdatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		bulletInputs := make([]EvidenceBulletInput, len(source.Bullets))
		for bulletIndex, bullet := range source.Bullets {
			if bullet.Position < 0 {
				return ProfileBackup{}, fmt.Errorf("experience %d evidence %d: position cannot be negative", index+1, bulletIndex+1)
			}
			if err := validateBackupTimestamp(fmt.Sprintf("experience %d evidence %d creation time", index+1, bulletIndex+1), bullet.CreatedAt, true); err != nil {
				return ProfileBackup{}, err
			}
			if err := validateBackupTimestamp(fmt.Sprintf("experience %d evidence %d update time", index+1, bulletIndex+1), bullet.UpdatedAt, true); err != nil {
				return ProfileBackup{}, err
			}
			bulletInputs[bulletIndex] = EvidenceBulletInput{ID: bullet.ID, Text: bullet.Text, Provenance: bullet.Provenance, SourceURL: bullet.SourceURL, Verification: bullet.Verification}
		}
		validated, err := (ExperienceInput{ID: source.ID, Company: source.Company, Title: source.Title, Location: source.Location, StartDate: source.StartDate, EndDate: source.EndDate, Current: source.Current, Bullets: bulletInputs}).Validate()
		if err != nil {
			return ProfileBackup{}, fmt.Errorf("experience %d: %w", index+1, err)
		}
		validated.Position, validated.CreatedAt, validated.UpdatedAt = source.Position, source.CreatedAt, source.UpdatedAt
		for bulletIndex := range validated.Bullets {
			validated.Bullets[bulletIndex].Position = source.Bullets[bulletIndex].Position
			validated.Bullets[bulletIndex].CreatedAt = source.Bullets[bulletIndex].CreatedAt
			validated.Bullets[bulletIndex].UpdatedAt = source.Bullets[bulletIndex].UpdatedAt
		}
		experiences = append(experiences, validated)
	}
	backup.Experiences = experiences

	projects := make([]Project, 0, len(backup.Projects))
	for index, source := range backup.Projects {
		if source.Position < 0 {
			return ProfileBackup{}, fmt.Errorf("project %d: position cannot be negative", index+1)
		}
		if err := validateBackupTimestamp(fmt.Sprintf("project %d creation time", index+1), source.CreatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		if err := validateBackupTimestamp(fmt.Sprintf("project %d update time", index+1), source.UpdatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		bulletInputs := make([]EvidenceBulletInput, len(source.Bullets))
		for bulletIndex, bullet := range source.Bullets {
			if bullet.Position < 0 {
				return ProfileBackup{}, fmt.Errorf("project %d evidence %d: position cannot be negative", index+1, bulletIndex+1)
			}
			if err := validateBackupTimestamp(fmt.Sprintf("project %d evidence %d creation time", index+1, bulletIndex+1), bullet.CreatedAt, true); err != nil {
				return ProfileBackup{}, err
			}
			if err := validateBackupTimestamp(fmt.Sprintf("project %d evidence %d update time", index+1, bulletIndex+1), bullet.UpdatedAt, true); err != nil {
				return ProfileBackup{}, err
			}
			bulletInputs[bulletIndex] = EvidenceBulletInput{ID: bullet.ID, Text: bullet.Text, Provenance: bullet.Provenance, SourceURL: bullet.SourceURL, Verification: bullet.Verification}
		}
		validated, err := (ProjectInput{ID: source.ID, Name: source.Name, Role: source.Role, Description: source.Description, URL: source.URL, RepositoryURL: source.RepositoryURL, StartDate: source.StartDate, EndDate: source.EndDate, Ongoing: source.Ongoing, Provenance: source.Provenance, Verification: source.Verification, ResumeEligible: source.ResumeEligible, Skills: source.Skills, DetectedLanguages: source.DetectedLanguages, Bullets: bulletInputs}).Validate()
		if err != nil {
			return ProfileBackup{}, fmt.Errorf("project %d: %w", index+1, err)
		}
		validated.Position, validated.CreatedAt, validated.UpdatedAt = source.Position, source.CreatedAt, source.UpdatedAt
		for bulletIndex := range validated.Bullets {
			validated.Bullets[bulletIndex].Position = source.Bullets[bulletIndex].Position
			validated.Bullets[bulletIndex].CreatedAt = source.Bullets[bulletIndex].CreatedAt
			validated.Bullets[bulletIndex].UpdatedAt = source.Bullets[bulletIndex].UpdatedAt
		}
		projects = append(projects, validated)
	}
	backup.Projects = projects

	educations := make([]Education, 0, len(backup.Educations))
	for index, source := range backup.Educations {
		if source.Position < 0 {
			return ProfileBackup{}, fmt.Errorf("education %d: position cannot be negative", index+1)
		}
		if err := validateBackupTimestamp(fmt.Sprintf("education %d creation time", index+1), source.CreatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		if err := validateBackupTimestamp(fmt.Sprintf("education %d update time", index+1), source.UpdatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		validated, err := (EducationInput{ID: source.ID, Institution: source.Institution, Degree: source.Degree, FieldOfStudy: source.FieldOfStudy, Location: source.Location, StartDate: source.StartDate, EndDate: source.EndDate, Current: source.Current, Details: source.Details}).Validate()
		if err != nil {
			return ProfileBackup{}, fmt.Errorf("education %d: %w", index+1, err)
		}
		validated.Position, validated.CreatedAt, validated.UpdatedAt = source.Position, source.CreatedAt, source.UpdatedAt
		educations = append(educations, validated)
	}
	backup.Educations = educations

	jobs := make([]Job, 0, len(backup.Jobs))
	for index, source := range backup.Jobs {
		if err := validateBackupTimestamp(fmt.Sprintf("job %d creation time", index+1), source.CreatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		if err := validateBackupTimestamp(fmt.Sprintf("job %d update time", index+1), source.UpdatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		validated, err := (JobInput{ID: source.ID, Company: source.Company, Role: source.Role, Description: source.Description}).Validate()
		if err != nil {
			return ProfileBackup{}, fmt.Errorf("job %d: %w", index+1, err)
		}
		validated.CreatedAt, validated.UpdatedAt = source.CreatedAt, source.UpdatedAt
		jobs = append(jobs, validated)
	}
	backup.Jobs = jobs

	applications := make([]Application, 0, len(backup.Applications))
	for index, source := range backup.Applications {
		if source.Status != "draft" && source.Status != "submitted" && source.Status != "archived" {
			return ProfileBackup{}, fmt.Errorf("application %d status is not valid", index+1)
		}
		if err := validateBackupTimestamp(fmt.Sprintf("application %d creation time", index+1), source.CreatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		if err := validateBackupTimestamp(fmt.Sprintf("application %d update time", index+1), source.UpdatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		selection, err := (CreateResumeVersionInput{JobID: source.JobID, TemplateID: "backup-validation", SelectedFactIDs: source.SelectedFactIDs}).Validate()
		if err != nil {
			return ProfileBackup{}, fmt.Errorf("application %d: %w", index+1, err)
		}
		application := Application{ID: source.ID, JobID: selection.JobID, Status: source.Status, SelectedFactIDs: selection.SelectedFactIDs, Versions: make([]ResumeVersion, 0, len(source.Versions)), CreatedAt: source.CreatedAt, UpdatedAt: source.UpdatedAt}
		versionNumbers := make(map[int]struct{}, len(source.Versions))
		for versionIndex, version := range source.Versions {
			if version.ApplicationID != source.ID {
				return ProfileBackup{}, fmt.Errorf("application %d resume version %d belongs to another application", index+1, versionIndex+1)
			}
			if version.VersionNumber < 1 {
				return ProfileBackup{}, fmt.Errorf("application %d resume version %d number is not valid", index+1, versionIndex+1)
			}
			if _, exists := versionNumbers[version.VersionNumber]; exists {
				return ProfileBackup{}, fmt.Errorf("application %d resume version numbers must be unique", index+1)
			}
			versionNumbers[version.VersionNumber] = struct{}{}
			if err := validateBackupTimestamp(fmt.Sprintf("application %d resume version %d creation time", index+1, versionIndex+1), version.CreatedAt, true); err != nil {
				return ProfileBackup{}, err
			}
			validatedVersion, err := (CreateResumeVersionInput{JobID: source.JobID, TemplateID: version.TemplateID, SelectedFactIDs: version.SelectedFactIDs}).Validate()
			if err != nil {
				return ProfileBackup{}, fmt.Errorf("application %d resume version %d: %w", index+1, versionIndex+1, err)
			}
			if len(version.JobDescriptionSnapshot) < 40 {
				return ProfileBackup{}, fmt.Errorf("application %d resume version %d job snapshot is not valid", index+1, versionIndex+1)
			}
			if len(version.LatexSource) == 0 || len(version.LatexSource) > MaxTemplateSourceBytes {
				return ProfileBackup{}, fmt.Errorf("application %d resume version %d LaTeX source is not valid", index+1, versionIndex+1)
			}
			if version.ContentHash == "" {
				version.ContentHash = ResumeContentHash(version.LatexSource)
			}
			if version.ContentHash != ResumeContentHash(version.LatexSource) {
				return ProfileBackup{}, fmt.Errorf("application %d resume version %d content hash does not match its source", index+1, versionIndex+1)
			}
			if version.CompileDurationMS < 0 {
				return ProfileBackup{}, fmt.Errorf("application %d resume version %d compile duration is not valid", index+1, versionIndex+1)
			}
			if err := validateBackupTimestamp(fmt.Sprintf("application %d resume version %d compile time", index+1, versionIndex+1), version.CompiledAt, false); err != nil {
				return ProfileBackup{}, err
			}
			version.PDFPath = ""
			version.PDFAvailable = false
			version.SelectedFactIDs = validatedVersion.SelectedFactIDs
			application.Versions = append(application.Versions, version)
		}
		applications = append(applications, application)
	}
	backup.Applications = applications

	templates := make([]ResumeTemplate, 0, len(backup.Templates))
	for index, source := range backup.Templates {
		if source.BuiltIn {
			return ProfileBackup{}, fmt.Errorf("template %d cannot be marked as built in", index+1)
		}
		if err := validateBackupTimestamp(fmt.Sprintf("template %d creation time", index+1), source.CreatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		if err := validateBackupTimestamp(fmt.Sprintf("template %d update time", index+1), source.UpdatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		validated, err := (ResumeTemplateInput{ID: source.ID, Name: source.Name, Description: source.Description, Source: source.Source}).Validate()
		if err != nil {
			return ProfileBackup{}, fmt.Errorf("template %d: %w", index+1, err)
		}
		validated.CreatedAt, validated.UpdatedAt = source.CreatedAt, source.UpdatedAt
		templates = append(templates, validated)
	}
	backup.Templates = templates
	backup.SelectedTemplateID = strings.TrimSpace(backup.SelectedTemplateID)
	if len(backup.SelectedTemplateID) > 200 {
		return ProfileBackup{}, fmt.Errorf("selected template ID is not valid")
	}

	runs := make([]AIRun, 0, len(backup.AIRuns))
	for index, source := range backup.AIRuns {
		if strings.TrimSpace(source.Provider) == "" || strings.TrimSpace(source.Model) == "" || strings.TrimSpace(source.PromptVersion) == "" || strings.TrimSpace(source.SchemaVersion) == "" {
			return ProfileBackup{}, fmt.Errorf("AI run %d metadata is incomplete", index+1)
		}
		if err := validateBackupTimestamp(fmt.Sprintf("AI run %d creation time", index+1), source.CreatedAt, true); err != nil {
			return ProfileBackup{}, err
		}
		if err := validateBackupTimestamp(fmt.Sprintf("AI run %d acceptance time", index+1), source.AcceptedAt, false); err != nil {
			return ProfileBackup{}, err
		}
		if source.ResumeVersionID == "" && source.AcceptedAt != "" {
			return ProfileBackup{}, fmt.Errorf("AI run %d has an acceptance time without a resume version", index+1)
		}
		if !source.ValidationPassed && source.ResumeVersionID != "" {
			return ProfileBackup{}, fmt.Errorf("AI run %d was accepted without passing validation", index+1)
		}
		for proposalIndex, proposal := range source.Proposals {
			if strings.TrimSpace(proposal.TargetFactID) == "" || len(proposal.SupportingFactIDs) == 0 || len(strings.TrimSpace(proposal.Text)) < 20 || len(proposal.Text) > 1200 {
				return ProfileBackup{}, fmt.Errorf("AI run %d proposal %d is not valid", index+1, proposalIndex+1)
			}
		}
		source.Provider = strings.TrimSpace(source.Provider)
		source.Model = strings.TrimSpace(source.Model)
		source.PromptVersion = strings.TrimSpace(source.PromptVersion)
		source.SchemaVersion = strings.TrimSpace(source.SchemaVersion)
		runs = append(runs, source)
	}
	backup.AIRuns = runs
	if backup.SchemaVersion < 3 {
		backup.AISettings = DefaultAISettings()
	}
	settings, err := backup.AISettings.Validate()
	if err != nil {
		return ProfileBackup{}, fmt.Errorf("AI settings: %w", err)
	}
	backup.AISettings = settings
	backup.SchemaVersion = ProfileBackupSchemaVersion
	return backup, nil
}

func validateBackupTimestamp(field, value string, required bool) error {
	if value == "" && !required {
		return nil
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return fmt.Errorf("%s is not valid", field)
	}
	return nil
}

func (backup ProfileBackup) Result(path string) BackupResult {
	versionCount := 0
	for _, application := range backup.Applications {
		versionCount += len(application.Versions)
	}
	return BackupResult{Path: path, ExperienceCount: len(backup.Experiences), ProjectCount: len(backup.Projects), EducationCount: len(backup.Educations), JobCount: len(backup.Jobs), ApplicationCount: len(backup.Applications), ResumeVersionCount: versionCount, TemplateCount: len(backup.Templates), AIRunCount: len(backup.AIRuns)}
}
