package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func (s *Store) CreateProfileBackup(ctx context.Context) (domain.ProfileBackup, error) {
	profile, err := s.GetProfile(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	experiences, err := s.ListExperiences(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	projects, err := s.ListProjects(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	educations, err := s.ListEducations(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	jobs, err := s.ListJobs(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	applications, err := s.ListApplications(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	templates, err := s.ListTemplates(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	selectedTemplateID, err := s.SelectedTemplateID(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	aiRuns, err := s.ListAIRuns(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	aiSettings, err := s.GetAISettings(ctx)
	if err != nil {
		return domain.ProfileBackup{}, err
	}
	return domain.ProfileBackup{
		SchemaVersion:      domain.ProfileBackupSchemaVersion,
		ExportedAt:         time.Now().UTC().Format(time.RFC3339),
		Profile:            profile,
		Experiences:        experiences,
		Projects:           projects,
		Educations:         educations,
		Jobs:               jobs,
		Applications:       applications,
		Templates:          templates,
		SelectedTemplateID: selectedTemplateID,
		AIRuns:             aiRuns,
		AISettings:         aiSettings,
	}, nil
}

func (s *Store) ReplaceProfileFromBackup(ctx context.Context, source domain.ProfileBackup) error {
	backup, err := source.Validate()
	if err != nil {
		return err
	}
	if err := validateBackupIDs(backup); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin backup import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, statement := range []string{
		`DELETE FROM ai_runs`,
		`DELETE FROM applications`,
		`DELETE FROM jobs`,
		`DELETE FROM projects`,
		`DELETE FROM experiences`,
		`DELETE FROM educations`,
		`DELETE FROM profile_skills`,
		`DELETE FROM profiles`,
		`DELETE FROM resume_templates`,
		`DELETE FROM app_settings`,
	} {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("clear profile data for import: %w", err)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	profileUpdatedAt := timestampOr(backup.Profile.UpdatedAt, now)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO profiles (
			id, name, headline, email, phone, location, website,
			github_username, linkedin_url, summary, updated_at
		) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, backup.Profile.Name, backup.Profile.Headline, backup.Profile.Email, backup.Profile.Phone,
		backup.Profile.Location, backup.Profile.Website, backup.Profile.GitHubUsername,
		backup.Profile.LinkedInURL, backup.Profile.Summary, profileUpdatedAt)
	if err != nil {
		return fmt.Errorf("import profile: %w", err)
	}
	for position, skill := range backup.Profile.Skills {
		if _, err := tx.ExecContext(ctx, `INSERT INTO profile_skills(position, name) VALUES (?, ?)`, position, skill); err != nil {
			return fmt.Errorf("import profile skill: %w", err)
		}
	}

	for _, experience := range backup.Experiences {
		createdAt := timestampOr(experience.CreatedAt, now)
		updatedAt := timestampOr(experience.UpdatedAt, createdAt)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO experiences (id, company, title, location, start_date, end_date, is_current, position, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, experience.ID, experience.Company, experience.Title, experience.Location, experience.StartDate,
			experience.EndDate, boolInt(experience.Current), experience.Position, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("import experience: %w", err)
		}
		for _, bullet := range experience.Bullets {
			bulletCreatedAt := timestampOr(bullet.CreatedAt, createdAt)
			bulletUpdatedAt := timestampOr(bullet.UpdatedAt, bulletCreatedAt)
			_, err := tx.ExecContext(ctx, `
				INSERT INTO experience_bullets (id, experience_id, text, provenance, source_url, verification_state, position, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, bullet.ID, experience.ID, bullet.Text, bullet.Provenance, bullet.SourceURL, bullet.Verification,
				bullet.Position, bulletCreatedAt, bulletUpdatedAt)
			if err != nil {
				return fmt.Errorf("import experience evidence: %w", err)
			}
		}
	}

	for _, project := range backup.Projects {
		createdAt := timestampOr(project.CreatedAt, now)
		updatedAt := timestampOr(project.UpdatedAt, createdAt)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO projects (
				id, name, role, description, url, repository_url, repository_id,
				repository_readme, repository_visibility, repository_updated_at, start_date, end_date,
				is_ongoing, provenance, verification_state, resume_eligible, position, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, project.ID, project.Name, project.Role, project.Description, project.URL, project.RepositoryURL,
			project.RepositoryID, project.RepositoryReadme, project.RepositoryVisibility, project.RepositoryUpdatedAt,
			project.StartDate, project.EndDate, boolInt(project.Ongoing), project.Provenance, project.Verification,
			boolInt(project.ResumeEligible), project.Position, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("import project: %w", err)
		}
		for position, skill := range project.Skills {
			if _, err := tx.ExecContext(ctx, `INSERT INTO project_skills(project_id, position, name) VALUES (?, ?, ?)`, project.ID, position, skill); err != nil {
				return fmt.Errorf("import project skill: %w", err)
			}
		}
		for position, language := range project.DetectedLanguages {
			if _, err := tx.ExecContext(ctx, `INSERT INTO project_detected_languages(project_id, position, name, code_bytes) VALUES (?, ?, ?, ?)`, project.ID, position, language.Name, language.Bytes); err != nil {
				return fmt.Errorf("import project detected language: %w", err)
			}
		}
		for _, bullet := range project.Bullets {
			bulletCreatedAt := timestampOr(bullet.CreatedAt, createdAt)
			bulletUpdatedAt := timestampOr(bullet.UpdatedAt, bulletCreatedAt)
			_, err := tx.ExecContext(ctx, `
				INSERT INTO project_bullets (id, project_id, text, provenance, source_url, verification_state, position, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, bullet.ID, project.ID, bullet.Text, bullet.Provenance, bullet.SourceURL, bullet.Verification,
				bullet.Position, bulletCreatedAt, bulletUpdatedAt)
			if err != nil {
				return fmt.Errorf("import project evidence: %w", err)
			}
		}
	}

	for _, education := range backup.Educations {
		createdAt := timestampOr(education.CreatedAt, now)
		updatedAt := timestampOr(education.UpdatedAt, createdAt)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO educations (
				id, institution, degree, field_of_study, location, start_date,
				end_date, is_current, details, position, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, education.ID, education.Institution, education.Degree, education.FieldOfStudy, education.Location,
			education.StartDate, education.EndDate, boolInt(education.Current), education.Details,
			education.Position, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("import education: %w", err)
		}
	}

	for _, job := range backup.Jobs {
		createdAt := timestampOr(job.CreatedAt, now)
		updatedAt := timestampOr(job.UpdatedAt, createdAt)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO jobs (id, company, role, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, job.ID, job.Company, job.Role, job.Description, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("import job: %w", err)
		}
	}

	for _, application := range backup.Applications {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO applications(id, job_id, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`, application.ID, application.JobID, application.Status, application.CreatedAt, application.UpdatedAt)
		if err != nil {
			return fmt.Errorf("import application: %w", err)
		}
		for position, factID := range application.SelectedFactIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO application_selected_facts(application_id, position, fact_id) VALUES (?, ?, ?)`, application.ID, position, factID); err != nil {
				return fmt.Errorf("import application selected fact: %w", err)
			}
		}
		for _, version := range application.Versions {
			selectedJSON, err := json.Marshal(version.SelectedFactIDs)
			if err != nil {
				return fmt.Errorf("encode imported resume version selection: %w", err)
			}
			rankingJSON, err := json.Marshal(version.RankingExplanations)
			if err != nil {
				return fmt.Errorf("encode imported resume version ranking explanations: %w", err)
			}
			diagnosticsJSON, err := json.Marshal(version.CompileDiagnostics)
			if err != nil {
				return fmt.Errorf("encode imported resume version diagnostics: %w", err)
			}
			_, err = tx.ExecContext(ctx, `
				INSERT INTO resume_versions(
					id, application_id, version_number, job_description_snapshot,
					selected_fact_ids_json, latex_source, template_id,
					ranking_explanations_json, content_hash, compile_success,
					compile_engine, compile_duration_ms, compile_diagnostics_json,
					compiled_at, pdf_path, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?)
			`, version.ID, application.ID, version.VersionNumber,
				version.JobDescriptionSnapshot, string(selectedJSON), version.LatexSource,
				version.TemplateID, string(rankingJSON), version.ContentHash,
				boolInt(version.CompileSuccess), version.CompileEngine,
				version.CompileDurationMS, string(diagnosticsJSON), version.CompiledAt,
				version.CreatedAt)
			if err != nil {
				return fmt.Errorf("import resume version: %w", err)
			}
		}
	}

	for _, template := range backup.Templates {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO resume_templates(id, name, description, source, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, template.ID, template.Name, template.Description, template.Source, template.CreatedAt, template.UpdatedAt)
		if err != nil {
			return fmt.Errorf("import resume template: %w", err)
		}
	}
	if backup.SelectedTemplateID != "" {
		if _, err := tx.ExecContext(ctx, `INSERT INTO app_settings(key, value) VALUES ('selected_template_id', ?)`, backup.SelectedTemplateID); err != nil {
			return fmt.Errorf("import selected resume template: %w", err)
		}
	}
	settingsJSON, err := json.Marshal(backup.AISettings)
	if err != nil {
		return fmt.Errorf("encode imported AI settings: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO app_settings(key, value) VALUES (?, ?)`, aiSettingsKey, string(settingsJSON)); err != nil {
		return fmt.Errorf("import AI settings: %w", err)
	}

	for _, run := range backup.AIRuns {
		selectedJSON, err := json.Marshal(run.SelectedFactIDs)
		if err != nil {
			return fmt.Errorf("encode imported AI run selection: %w", err)
		}
		errorsJSON, err := json.Marshal(run.ValidationErrors)
		if err != nil {
			return fmt.Errorf("encode imported AI run validation errors: %w", err)
		}
		proposalsJSON, err := json.Marshal(run.Proposals)
		if err != nil {
			return fmt.Errorf("encode imported AI run proposals: %w", err)
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO ai_runs(
				id, job_id, provider, model, prompt_version, schema_version,
				selected_fact_ids_json, validation_passed, failure_category, validation_errors_json,
				proposals_json, resume_version_id, created_at, accepted_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?)
		`, run.ID, run.JobID, run.Provider, run.Model, run.PromptVersion,
			run.SchemaVersion, string(selectedJSON), boolInt(run.ValidationPassed),
			run.FailureCategory, string(errorsJSON), string(proposalsJSON), run.ResumeVersionID,
			run.CreatedAt, run.AcceptedAt)
		if err != nil {
			return fmt.Errorf("import AI run: %w", err)
		}
	}

	if err := rebuildEvidenceSearch(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit backup import: %w", err)
	}
	return nil
}

func validateBackupIDs(backup domain.ProfileBackup) error {
	experienceIDs := make(map[string]struct{}, len(backup.Experiences))
	experienceBulletIDs := make(map[string]struct{})
	for index, experience := range backup.Experiences {
		if err := validateUniqueUUID(experience.ID, experienceIDs); err != nil {
			return fmt.Errorf("experience %d ID: %w", index+1, err)
		}
		for bulletIndex, bullet := range experience.Bullets {
			if err := validateUniqueUUID(bullet.ID, experienceBulletIDs); err != nil {
				return fmt.Errorf("experience %d evidence %d ID: %w", index+1, bulletIndex+1, err)
			}
		}
	}
	projectIDs := make(map[string]struct{}, len(backup.Projects))
	projectBulletIDs := make(map[string]struct{})
	for index, project := range backup.Projects {
		if err := validateUniqueUUID(project.ID, projectIDs); err != nil {
			return fmt.Errorf("project %d ID: %w", index+1, err)
		}
		for bulletIndex, bullet := range project.Bullets {
			if err := validateUniqueUUID(bullet.ID, projectBulletIDs); err != nil {
				return fmt.Errorf("project %d evidence %d ID: %w", index+1, bulletIndex+1, err)
			}
		}
	}
	educationIDs := make(map[string]struct{}, len(backup.Educations))
	for index, education := range backup.Educations {
		if err := validateUniqueUUID(education.ID, educationIDs); err != nil {
			return fmt.Errorf("education %d ID: %w", index+1, err)
		}
	}
	jobIDs := make(map[string]struct{}, len(backup.Jobs))
	for index, job := range backup.Jobs {
		if err := validateUniqueUUID(job.ID, jobIDs); err != nil {
			return fmt.Errorf("job %d ID: %w", index+1, err)
		}
	}
	applicationIDs := make(map[string]struct{}, len(backup.Applications))
	versionIDs := make(map[string]struct{})
	for index, application := range backup.Applications {
		if err := validateUniqueUUID(application.ID, applicationIDs); err != nil {
			return fmt.Errorf("application %d ID: %w", index+1, err)
		}
		if _, exists := jobIDs[application.JobID]; !exists {
			return fmt.Errorf("application %d references an unknown job", index+1)
		}
		selectedIDs := make(map[string]struct{}, len(application.SelectedFactIDs))
		for factIndex, factID := range application.SelectedFactIDs {
			if err := validateUniqueUUID(factID, selectedIDs); err != nil {
				return fmt.Errorf("application %d selected fact %d: %w", index+1, factIndex+1, err)
			}
		}
		for versionIndex, version := range application.Versions {
			if err := validateUniqueUUID(version.ID, versionIDs); err != nil {
				return fmt.Errorf("application %d resume version %d ID: %w", index+1, versionIndex+1, err)
			}
			versionSelectedIDs := make(map[string]struct{}, len(version.SelectedFactIDs))
			for factIndex, factID := range version.SelectedFactIDs {
				if err := validateUniqueUUID(factID, versionSelectedIDs); err != nil {
					return fmt.Errorf("application %d resume version %d selected fact %d: %w", index+1, versionIndex+1, factIndex+1, err)
				}
			}
		}
	}
	templateIDs := make(map[string]struct{}, len(backup.Templates))
	for index, template := range backup.Templates {
		if err := validateUniqueUUID(template.ID, templateIDs); err != nil {
			return fmt.Errorf("template %d ID: %w", index+1, err)
		}
	}
	if backup.SelectedTemplateID != "" {
		if selectedID, err := uuid.Parse(backup.SelectedTemplateID); err == nil {
			if _, exists := templateIDs[selectedID.String()]; !exists {
				return fmt.Errorf("selected template references an unknown custom template")
			}
		}
	}
	runIDs := make(map[string]struct{}, len(backup.AIRuns))
	for index, run := range backup.AIRuns {
		if err := validateUniqueUUID(run.ID, runIDs); err != nil {
			return fmt.Errorf("AI run %d ID: %w", index+1, err)
		}
		if _, exists := jobIDs[run.JobID]; !exists {
			return fmt.Errorf("AI run %d references an unknown job", index+1)
		}
		if run.ResumeVersionID != "" {
			if _, exists := versionIDs[run.ResumeVersionID]; !exists {
				return fmt.Errorf("AI run %d references an unknown resume version", index+1)
			}
		}
		selectedIDs := make(map[string]struct{}, len(run.SelectedFactIDs))
		for factIndex, factID := range run.SelectedFactIDs {
			if err := validateUniqueUUID(factID, selectedIDs); err != nil {
				return fmt.Errorf("AI run %d selected fact %d: %w", index+1, factIndex+1, err)
			}
		}
		for proposalIndex, proposal := range run.Proposals {
			if _, exists := selectedIDs[proposal.TargetFactID]; !exists {
				return fmt.Errorf("AI run %d proposal %d targets an unselected fact", index+1, proposalIndex+1)
			}
			for citationIndex, factID := range proposal.SupportingFactIDs {
				if _, exists := selectedIDs[factID]; !exists {
					return fmt.Errorf("AI run %d proposal %d citation %d references an unselected fact", index+1, proposalIndex+1, citationIndex+1)
				}
			}
		}
	}
	return nil
}

func validateUniqueUUID(id string, seen map[string]struct{}) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("is not a valid UUID")
	}
	if _, exists := seen[id]; exists {
		return fmt.Errorf("must be unique")
	}
	seen[id] = struct{}{}
	return nil
}

func timestampOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
