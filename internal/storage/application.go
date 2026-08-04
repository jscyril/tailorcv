package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func (s *Store) ListApplications(ctx context.Context) ([]domain.Application, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, job_id, status, created_at, updated_at FROM applications ORDER BY updated_at DESC, created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	applications := make([]domain.Application, 0)
	for rows.Next() {
		var application domain.Application
		if err := rows.Scan(&application.ID, &application.JobID, &application.Status, &application.CreatedAt, &application.UpdatedAt); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan application: %w", err)
		}
		applications = append(applications, application)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate applications: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close application rows: %w", err)
	}
	for index := range applications {
		selected, err := s.listApplicationSelectedFacts(ctx, applications[index].ID)
		if err != nil {
			return nil, err
		}
		applications[index].SelectedFactIDs = selected
		versions, err := s.ListResumeVersions(ctx, applications[index].ID)
		if err != nil {
			return nil, err
		}
		applications[index].Versions = versions
	}
	return applications, nil
}

func (s *Store) listApplicationSelectedFacts(ctx context.Context, applicationID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT fact_id FROM application_selected_facts WHERE application_id = ? ORDER BY position`, applicationID)
	if err != nil {
		return nil, fmt.Errorf("list selected facts: %w", err)
	}
	defer rows.Close()
	selected := make([]string, 0)
	for rows.Next() {
		var factID string
		if err := rows.Scan(&factID); err != nil {
			return nil, fmt.Errorf("scan selected fact: %w", err)
		}
		selected = append(selected, factID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate selected facts: %w", err)
	}
	return selected, nil
}

func (s *Store) ListResumeVersions(ctx context.Context, applicationID string) ([]domain.ResumeVersion, error) {
	if _, err := uuid.Parse(applicationID); err != nil {
		return nil, fmt.Errorf("application ID is not valid")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, application_id, version_number, job_description_snapshot,
		       selected_fact_ids_json, latex_source, template_id, created_at
		FROM resume_versions WHERE application_id = ? ORDER BY version_number DESC
	`, applicationID)
	if err != nil {
		return nil, fmt.Errorf("list resume versions: %w", err)
	}
	defer rows.Close()
	versions := make([]domain.ResumeVersion, 0)
	for rows.Next() {
		var version domain.ResumeVersion
		var selectedJSON string
		if err := rows.Scan(&version.ID, &version.ApplicationID, &version.VersionNumber, &version.JobDescriptionSnapshot, &selectedJSON, &version.LatexSource, &version.TemplateID, &version.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan resume version: %w", err)
		}
		if err := json.Unmarshal([]byte(selectedJSON), &version.SelectedFactIDs); err != nil {
			return nil, fmt.Errorf("decode resume version selected facts: %w", err)
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resume versions: %w", err)
	}
	return versions, nil
}

// CreateResumeVersion updates the application's current selection and appends
// a new immutable snapshot. Existing resume-version rows are never updated.
func (s *Store) CreateResumeVersion(ctx context.Context, jobID string, selectedFactIDs []string, templateID, jobSnapshot, latexSource string) (domain.ApplicationResumeResult, error) {
	if _, err := uuid.Parse(jobID); err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("job ID is not valid")
	}
	for _, factID := range selectedFactIDs {
		if _, err := uuid.Parse(factID); err != nil {
			return domain.ApplicationResumeResult{}, fmt.Errorf("selected evidence ID is not valid")
		}
	}
	selectedJSON, err := json.Marshal(selectedFactIDs)
	if err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("encode selected evidence: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("begin resume version: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	application := domain.Application{JobID: jobID, Status: "draft", SelectedFactIDs: append([]string(nil), selectedFactIDs...), Versions: []domain.ResumeVersion{}}
	err = tx.QueryRowContext(ctx, `SELECT id, status, created_at FROM applications WHERE job_id = ?`, jobID).Scan(&application.ID, &application.Status, &application.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		application.ID = uuid.NewString()
		application.CreatedAt = now
		_, err = tx.ExecContext(ctx, `INSERT INTO applications(id, job_id, status, created_at, updated_at) VALUES (?, ?, 'draft', ?, ?)`, application.ID, jobID, now, now)
	} else if err == nil {
		_, err = tx.ExecContext(ctx, `UPDATE applications SET updated_at = ? WHERE id = ?`, now, application.ID)
	}
	if err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("save application: %w", err)
	}
	application.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `DELETE FROM application_selected_facts WHERE application_id = ?`, application.ID); err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("replace selected facts: %w", err)
	}
	for position, factID := range selectedFactIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO application_selected_facts(application_id, position, fact_id) VALUES (?, ?, ?)`, application.ID, position, factID); err != nil {
			return domain.ApplicationResumeResult{}, fmt.Errorf("save selected fact: %w", err)
		}
	}

	version := domain.ResumeVersion{ID: uuid.NewString(), ApplicationID: application.ID, JobDescriptionSnapshot: jobSnapshot, SelectedFactIDs: append([]string(nil), selectedFactIDs...), LatexSource: latexSource, TemplateID: templateID, CreatedAt: now}
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number), 0) + 1 FROM resume_versions WHERE application_id = ?`, application.ID).Scan(&version.VersionNumber); err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("choose resume version number: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO resume_versions(id, application_id, version_number, job_description_snapshot, selected_fact_ids_json, latex_source, template_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, version.ID, version.ApplicationID, version.VersionNumber, version.JobDescriptionSnapshot, string(selectedJSON), version.LatexSource, version.TemplateID, version.CreatedAt)
	if err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("save immutable resume version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("commit resume version: %w", err)
	}
	application.Versions = []domain.ResumeVersion{version}
	return domain.ApplicationResumeResult{Application: application, Version: version}, nil
}
