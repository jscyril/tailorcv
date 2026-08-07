package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

func (s *Store) UpdateApplicationStatus(ctx context.Context, input domain.UpdateApplicationStatusInput) (domain.Application, error) {
	validated, err := input.Validate()
	if err != nil {
		return domain.Application{}, err
	}
	if _, err := uuid.Parse(validated.ApplicationID); err != nil {
		return domain.Application{}, fmt.Errorf("application ID is not valid")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx, `UPDATE applications SET status = ?, updated_at = ? WHERE id = ?`, validated.Status, now, validated.ApplicationID)
	if err != nil {
		return domain.Application{}, fmt.Errorf("update application status: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return domain.Application{}, fmt.Errorf("application was not found")
	}
	applications, err := s.ListApplications(ctx)
	if err != nil {
		return domain.Application{}, err
	}
	for _, application := range applications {
		if application.ID == validated.ApplicationID {
			return application, nil
		}
	}
	return domain.Application{}, fmt.Errorf("application was not found")
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
		       selected_fact_ids_json, latex_source, template_id,
		       ranking_explanations_json, content_hash, compile_success,
		       compile_engine, compile_duration_ms, compile_diagnostics_json,
		       compiled_at, pdf_path, created_at
		FROM resume_versions WHERE application_id = ? ORDER BY version_number DESC
	`, applicationID)
	if err != nil {
		return nil, fmt.Errorf("list resume versions: %w", err)
	}
	defer rows.Close()
	versions := make([]domain.ResumeVersion, 0)
	for rows.Next() {
		var version domain.ResumeVersion
		var selectedJSON, rankingJSON, diagnosticsJSON string
		var compileSuccess int
		if err := rows.Scan(
			&version.ID, &version.ApplicationID, &version.VersionNumber,
			&version.JobDescriptionSnapshot, &selectedJSON, &version.LatexSource,
			&version.TemplateID, &rankingJSON, &version.ContentHash, &compileSuccess,
			&version.CompileEngine, &version.CompileDurationMS, &diagnosticsJSON,
			&version.CompiledAt, &version.PDFPath, &version.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan resume version: %w", err)
		}
		if err := json.Unmarshal([]byte(selectedJSON), &version.SelectedFactIDs); err != nil {
			return nil, fmt.Errorf("decode resume version selected facts: %w", err)
		}
		if err := json.Unmarshal([]byte(rankingJSON), &version.RankingExplanations); err != nil {
			return nil, fmt.Errorf("decode resume version ranking explanations: %w", err)
		}
		if err := json.Unmarshal([]byte(diagnosticsJSON), &version.CompileDiagnostics); err != nil {
			return nil, fmt.Errorf("decode resume version compile diagnostics: %w", err)
		}
		version.CompileSuccess = compileSuccess != 0
		if version.PDFPath != "" {
			_, artifactErr := os.Stat(version.PDFPath)
			version.PDFAvailable = artifactErr == nil
		}
		if version.ContentHash == "" {
			version.ContentHash = domain.ResumeContentHash(version.LatexSource)
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
func (s *Store) CreateResumeVersion(ctx context.Context, jobID string, selectedFactIDs []string, templateID, jobSnapshot, latexSource string, rankingExplanations ...[]domain.EvidenceMatch) (domain.ApplicationResumeResult, error) {
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
	ranking := []domain.EvidenceMatch{}
	if len(rankingExplanations) > 0 {
		ranking = append(ranking, rankingExplanations[0]...)
	}
	rankingJSON, err := json.Marshal(ranking)
	if err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("encode ranking explanations: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("begin resume version: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	application := domain.Application{JobID: jobID, Status: domain.ApplicationStatusDraft, SelectedFactIDs: append([]string(nil), selectedFactIDs...), Versions: []domain.ResumeVersion{}}
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

	version := domain.ResumeVersion{ID: uuid.NewString(), ApplicationID: application.ID, JobDescriptionSnapshot: jobSnapshot, SelectedFactIDs: append([]string(nil), selectedFactIDs...), LatexSource: latexSource, TemplateID: templateID, RankingExplanations: ranking, ContentHash: domain.ResumeContentHash(latexSource), CompileDiagnostics: []domain.CompileDiagnostic{}, CreatedAt: now}
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number), 0) + 1 FROM resume_versions WHERE application_id = ?`, application.ID).Scan(&version.VersionNumber); err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("choose resume version number: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO resume_versions(
			id, application_id, version_number, job_description_snapshot,
			selected_fact_ids_json, latex_source, template_id,
			ranking_explanations_json, content_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, version.ID, version.ApplicationID, version.VersionNumber, version.JobDescriptionSnapshot, string(selectedJSON), version.LatexSource, version.TemplateID, string(rankingJSON), version.ContentHash, version.CreatedAt)
	if err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("save immutable resume version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ApplicationResumeResult{}, fmt.Errorf("commit resume version: %w", err)
	}
	application.Versions = []domain.ResumeVersion{version}
	return domain.ApplicationResumeResult{Application: application, Version: version}, nil
}

func (s *Store) CreateEditedResumeVersion(ctx context.Context, input domain.SaveResumeVersionEditInput) (domain.ResumeVersion, error) {
	validated, err := input.Validate()
	if err != nil {
		return domain.ResumeVersion{}, err
	}
	if _, err := uuid.Parse(validated.ApplicationID); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("application ID is not valid")
	}
	if _, err := uuid.Parse(validated.BaseVersionID); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("base resume version ID is not valid")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("begin edited resume version: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var base domain.ResumeVersion
	var selectedJSON, rankingJSON string
	err = tx.QueryRowContext(ctx, `
		SELECT job_description_snapshot, selected_fact_ids_json, template_id,
		       ranking_explanations_json
		FROM resume_versions WHERE id = ? AND application_id = ?
	`, validated.BaseVersionID, validated.ApplicationID).Scan(
		&base.JobDescriptionSnapshot, &selectedJSON, &base.TemplateID, &rankingJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ResumeVersion{}, fmt.Errorf("base resume version was not found in this application")
	}
	if err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("read base resume version: %w", err)
	}
	if err := json.Unmarshal([]byte(selectedJSON), &base.SelectedFactIDs); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("decode base resume selection: %w", err)
	}
	if err := json.Unmarshal([]byte(rankingJSON), &base.RankingExplanations); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("decode base ranking explanations: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	version := domain.ResumeVersion{
		ID: uuid.NewString(), ApplicationID: validated.ApplicationID,
		JobDescriptionSnapshot: base.JobDescriptionSnapshot,
		SelectedFactIDs:        append([]string(nil), base.SelectedFactIDs...),
		LatexSource:            validated.LatexSource, TemplateID: base.TemplateID,
		RankingExplanations: append([]domain.EvidenceMatch(nil), base.RankingExplanations...),
		ContentHash:         domain.ResumeContentHash(validated.LatexSource),
		CompileDiagnostics:  []domain.CompileDiagnostic{}, CreatedAt: now,
	}
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number), 0) + 1 FROM resume_versions WHERE application_id = ?`, validated.ApplicationID).Scan(&version.VersionNumber); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("choose edited resume version number: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO resume_versions(
			id, application_id, version_number, job_description_snapshot,
			selected_fact_ids_json, latex_source, template_id,
			ranking_explanations_json, content_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, version.ID, version.ApplicationID, version.VersionNumber,
		version.JobDescriptionSnapshot, selectedJSON, version.LatexSource,
		version.TemplateID, rankingJSON, version.ContentHash, version.CreatedAt)
	if err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("save edited resume version: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE applications SET updated_at = ? WHERE id = ?`, now, validated.ApplicationID); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("update application after resume edit: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("commit edited resume version: %w", err)
	}
	return version, nil
}

func (s *Store) GetResumeVersion(ctx context.Context, id string) (domain.ResumeVersion, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("resume version ID is not valid")
	}
	var applicationID string
	if err := s.db.QueryRowContext(ctx, `SELECT application_id FROM resume_versions WHERE id = ?`, id).Scan(&applicationID); errors.Is(err, sql.ErrNoRows) {
		return domain.ResumeVersion{}, fmt.Errorf("resume version was not found")
	} else if err != nil {
		return domain.ResumeVersion{}, fmt.Errorf("find resume version: %w", err)
	}
	versions, err := s.ListResumeVersions(ctx, applicationID)
	if err != nil {
		return domain.ResumeVersion{}, err
	}
	for _, version := range versions {
		if version.ID == id {
			return version, nil
		}
	}
	return domain.ResumeVersion{}, fmt.Errorf("resume version was not found")
}

func (s *Store) RecordResumeCompilation(ctx context.Context, versionID, pdfPath string, result domain.CompileResult) error {
	if _, err := uuid.Parse(versionID); err != nil {
		return fmt.Errorf("resume version ID is not valid")
	}
	diagnosticsJSON, err := json.Marshal(result.Diagnostics)
	if err != nil {
		return fmt.Errorf("encode compile diagnostics: %w", err)
	}
	compiledAt := time.Now().UTC().Format(time.RFC3339Nano)
	if !result.Success {
		pdfPath = ""
	}
	update, err := s.db.ExecContext(ctx, `
		UPDATE resume_versions SET compile_success = ?, compile_engine = ?,
			compile_duration_ms = ?, compile_diagnostics_json = ?, compiled_at = ?, pdf_path = ?
		WHERE id = ?
	`, boolInt(result.Success), result.Engine, result.DurationMS, string(diagnosticsJSON), compiledAt, pdfPath, versionID)
	if err != nil {
		return fmt.Errorf("record resume compilation: %w", err)
	}
	if count, err := update.RowsAffected(); err != nil || count != 1 {
		return fmt.Errorf("resume version was not found")
	}
	return nil
}
