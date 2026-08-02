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

func (s *Store) ListEducations(ctx context.Context) ([]domain.Education, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, institution, degree, field_of_study, location, start_date,
		       end_date, is_current, details, position, created_at, updated_at
		FROM educations
		ORDER BY position, created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("list educations: %w", err)
	}
	defer rows.Close()

	educations := make([]domain.Education, 0)
	for rows.Next() {
		var education domain.Education
		var current int
		if err := rows.Scan(
			&education.ID,
			&education.Institution,
			&education.Degree,
			&education.FieldOfStudy,
			&education.Location,
			&education.StartDate,
			&education.EndDate,
			&current,
			&education.Details,
			&education.Position,
			&education.CreatedAt,
			&education.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan education: %w", err)
		}
		education.Current = current == 1
		educations = append(educations, education)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate educations: %w", err)
	}
	return educations, nil
}

func (s *Store) SaveEducation(ctx context.Context, education domain.Education) (domain.Education, error) {
	if education.ID == "" {
		education.ID = uuid.NewString()
	} else if _, err := uuid.Parse(education.ID); err != nil {
		return domain.Education{}, fmt.Errorf("education ID is not valid")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var createdAt string
	var position int
	err := s.db.QueryRowContext(ctx, `SELECT created_at, position FROM educations WHERE id = ?`, education.ID).Scan(&createdAt, &position)
	if errors.Is(err, sql.ErrNoRows) {
		createdAt = now
		if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM educations`).Scan(&position); err != nil {
			return domain.Education{}, fmt.Errorf("choose education position: %w", err)
		}
	} else if err != nil {
		return domain.Education{}, fmt.Errorf("read existing education: %w", err)
	}

	current := 0
	if education.Current {
		current = 1
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO educations (
			id, institution, degree, field_of_study, location, start_date,
			end_date, is_current, details, position, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			institution = excluded.institution,
			degree = excluded.degree,
			field_of_study = excluded.field_of_study,
			location = excluded.location,
			start_date = excluded.start_date,
			end_date = excluded.end_date,
			is_current = excluded.is_current,
			details = excluded.details,
			updated_at = excluded.updated_at
	`,
		education.ID,
		education.Institution,
		education.Degree,
		education.FieldOfStudy,
		education.Location,
		education.StartDate,
		education.EndDate,
		current,
		education.Details,
		position,
		createdAt,
		now,
	)
	if err != nil {
		return domain.Education{}, fmt.Errorf("write education: %w", err)
	}

	education.Position = position
	education.CreatedAt = createdAt
	education.UpdatedAt = now
	return education, nil
}

func (s *Store) DeleteEducation(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("education ID is not valid")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM educations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete education: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if deleted == 0 {
		return fmt.Errorf("education was not found")
	}
	return nil
}
