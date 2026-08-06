package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func (s *Store) SaveAIRun(ctx context.Context, run domain.AIRun) (domain.AIRun, error) {
	if _, err := uuid.Parse(run.JobID); err != nil {
		return domain.AIRun{}, fmt.Errorf("AI run job ID is not valid")
	}
	run.Provider = strings.TrimSpace(run.Provider)
	run.Model = strings.TrimSpace(run.Model)
	run.PromptVersion = strings.TrimSpace(run.PromptVersion)
	run.SchemaVersion = strings.TrimSpace(run.SchemaVersion)
	run.FailureCategory = strings.TrimSpace(run.FailureCategory)
	if run.Provider == "" || run.Model == "" || run.PromptVersion == "" || run.SchemaVersion == "" {
		return domain.AIRun{}, fmt.Errorf("AI run metadata is incomplete")
	}
	if len(run.ValidationErrors) > 50 {
		return domain.AIRun{}, fmt.Errorf("AI run contains too many validation errors")
	}
	for _, message := range run.ValidationErrors {
		if len(message) > 2000 {
			return domain.AIRun{}, fmt.Errorf("AI run validation error exceeds the size limit")
		}
	}
	if run.ValidationPassed {
		run.FailureCategory = ""
		if len(run.ValidationErrors) != 0 || len(run.Proposals) == 0 {
			return domain.AIRun{}, fmt.Errorf("a validated AI run must contain proposals and no validation errors")
		}
	} else {
		run.Proposals = []domain.AIProposal{}
	}
	for _, factID := range run.SelectedFactIDs {
		if _, err := uuid.Parse(factID); err != nil {
			return domain.AIRun{}, fmt.Errorf("AI run selected fact ID is not valid")
		}
	}
	selectedJSON, err := json.Marshal(run.SelectedFactIDs)
	if err != nil {
		return domain.AIRun{}, fmt.Errorf("encode AI run selected facts: %w", err)
	}
	errorsJSON, err := json.Marshal(run.ValidationErrors)
	if err != nil {
		return domain.AIRun{}, fmt.Errorf("encode AI run validation errors: %w", err)
	}
	proposalsJSON, err := json.Marshal(run.Proposals)
	if err != nil {
		return domain.AIRun{}, fmt.Errorf("encode AI run proposals: %w", err)
	}
	if run.ID == "" {
		run.ID = uuid.NewString()
	} else if _, err := uuid.Parse(run.ID); err != nil {
		return domain.AIRun{}, fmt.Errorf("AI run ID is not valid")
	}
	if run.CreatedAt == "" {
		run.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO ai_runs(
			id, job_id, provider, model, prompt_version, schema_version,
			selected_fact_ids_json, validation_passed, failure_category, validation_errors_json,
			proposals_json, resume_version_id, created_at, accepted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?)
	`, run.ID, run.JobID, run.Provider, run.Model, run.PromptVersion, run.SchemaVersion,
		string(selectedJSON), boolInt(run.ValidationPassed), run.FailureCategory, string(errorsJSON),
		string(proposalsJSON), run.ResumeVersionID, run.CreatedAt, run.AcceptedAt)
	if err != nil {
		return domain.AIRun{}, fmt.Errorf("save AI run: %w", err)
	}
	return run, nil
}

func (s *Store) GetAIRun(ctx context.Context, id string) (domain.AIRun, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.AIRun{}, fmt.Errorf("AI run ID is not valid")
	}
	var run domain.AIRun
	var selectedJSON, errorsJSON, proposalsJSON string
	var passed int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, job_id, provider, model, prompt_version, schema_version,
		       selected_fact_ids_json, validation_passed, failure_category, validation_errors_json,
		       proposals_json, COALESCE(resume_version_id, ''), created_at, accepted_at
		FROM ai_runs WHERE id = ?
	`, id).Scan(&run.ID, &run.JobID, &run.Provider, &run.Model, &run.PromptVersion,
		&run.SchemaVersion, &selectedJSON, &passed, &run.FailureCategory, &errorsJSON, &proposalsJSON,
		&run.ResumeVersionID, &run.CreatedAt, &run.AcceptedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AIRun{}, fmt.Errorf("AI run was not found")
	}
	if err != nil {
		return domain.AIRun{}, fmt.Errorf("read AI run: %w", err)
	}
	if err := decodeAIRunJSON(&run, selectedJSON, errorsJSON, proposalsJSON); err != nil {
		return domain.AIRun{}, err
	}
	run.ValidationPassed = passed != 0
	return run, nil
}

func (s *Store) ListAIRuns(ctx context.Context) ([]domain.AIRun, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, job_id, provider, model, prompt_version, schema_version,
		       selected_fact_ids_json, validation_passed, failure_category, validation_errors_json,
		       proposals_json, COALESCE(resume_version_id, ''), created_at, accepted_at
		FROM ai_runs ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list AI runs: %w", err)
	}
	defer rows.Close()
	runs := make([]domain.AIRun, 0)
	for rows.Next() {
		var run domain.AIRun
		var selectedJSON, errorsJSON, proposalsJSON string
		var passed int
		if err := rows.Scan(&run.ID, &run.JobID, &run.Provider, &run.Model,
			&run.PromptVersion, &run.SchemaVersion, &selectedJSON, &passed, &run.FailureCategory,
			&errorsJSON, &proposalsJSON, &run.ResumeVersionID, &run.CreatedAt,
			&run.AcceptedAt); err != nil {
			return nil, fmt.Errorf("scan AI run: %w", err)
		}
		if err := decodeAIRunJSON(&run, selectedJSON, errorsJSON, proposalsJSON); err != nil {
			return nil, err
		}
		run.ValidationPassed = passed != 0
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI runs: %w", err)
	}
	return runs, nil
}

func decodeAIRunJSON(run *domain.AIRun, selectedJSON, errorsJSON, proposalsJSON string) error {
	if err := json.Unmarshal([]byte(selectedJSON), &run.SelectedFactIDs); err != nil {
		return fmt.Errorf("decode AI run selected facts: %w", err)
	}
	if err := json.Unmarshal([]byte(errorsJSON), &run.ValidationErrors); err != nil {
		return fmt.Errorf("decode AI run validation errors: %w", err)
	}
	if err := json.Unmarshal([]byte(proposalsJSON), &run.Proposals); err != nil {
		return fmt.Errorf("decode AI run proposals: %w", err)
	}
	return nil
}

func (s *Store) MarkAIRunAccepted(ctx context.Context, runID, versionID string) error {
	if _, err := uuid.Parse(runID); err != nil {
		return fmt.Errorf("AI run ID is not valid")
	}
	if _, err := uuid.Parse(versionID); err != nil {
		return fmt.Errorf("resume version ID is not valid")
	}
	acceptedAt := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx, `
		UPDATE ai_runs SET resume_version_id = ?, accepted_at = ?
		WHERE id = ? AND resume_version_id IS NULL AND validation_passed = 1
	`, versionID, acceptedAt, runID)
	if err != nil {
		return fmt.Errorf("mark AI run accepted: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return fmt.Errorf("AI run was already accepted or did not pass validation")
	}
	return nil
}
