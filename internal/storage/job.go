package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func (s *Store) ListJobs(ctx context.Context) ([]domain.Job, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company, role, description, created_at, updated_at
		FROM jobs
		ORDER BY updated_at DESC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()
	jobs := make([]domain.Job, 0)
	for rows.Next() {
		var job domain.Job
		if err := rows.Scan(&job.ID, &job.Company, &job.Role, &job.Description, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs: %w", err)
	}
	return jobs, nil
}

func (s *Store) SaveJob(ctx context.Context, job domain.Job) (domain.Job, error) {
	if job.ID == "" {
		job.ID = uuid.NewString()
	} else if _, err := uuid.Parse(job.ID); err != nil {
		return domain.Job{}, fmt.Errorf("job ID is not valid")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	var createdAt string
	err := s.db.QueryRowContext(ctx, `SELECT created_at FROM jobs WHERE id = ?`, job.ID).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		createdAt = now
	} else if err != nil {
		return domain.Job{}, fmt.Errorf("read existing job: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO jobs (id, company, role, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			company = excluded.company,
			role = excluded.role,
			description = excluded.description,
			updated_at = excluded.updated_at
	`, job.ID, job.Company, job.Role, job.Description, createdAt, now)
	if err != nil {
		return domain.Job{}, fmt.Errorf("write job: %w", err)
	}
	job.CreatedAt = createdAt
	job.UpdatedAt = now
	return job, nil
}

func (s *Store) DeleteJob(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("job ID is not valid")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if deleted == 0 {
		return fmt.Errorf("job was not found")
	}
	return nil
}
