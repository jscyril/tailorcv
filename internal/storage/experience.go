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

func (s *Store) ListExperiences(ctx context.Context) ([]domain.Experience, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company, title, location, start_date, end_date, is_current,
		       position, created_at, updated_at
		FROM experiences
		ORDER BY position, created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("list experiences: %w", err)
	}

	experiences := make([]domain.Experience, 0)
	for rows.Next() {
		var experience domain.Experience
		var current int
		if err := rows.Scan(
			&experience.ID,
			&experience.Company,
			&experience.Title,
			&experience.Location,
			&experience.StartDate,
			&experience.EndDate,
			&current,
			&experience.Position,
			&experience.CreatedAt,
			&experience.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan experience: %w", err)
		}
		experience.Current = current == 1
		experience.Bullets = []domain.EvidenceBullet{}
		experiences = append(experiences, experience)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate experiences: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close experience rows: %w", err)
	}

	for index := range experiences {
		bullets, err := s.listExperienceBullets(ctx, experiences[index].ID)
		if err != nil {
			return nil, err
		}
		experiences[index].Bullets = bullets
	}
	return experiences, nil
}

func (s *Store) listExperienceBullets(ctx context.Context, experienceID string) ([]domain.EvidenceBullet, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, text, provenance, source_url, verification_state, position, created_at, updated_at
		FROM experience_bullets
		WHERE experience_id = ?
		ORDER BY position
	`, experienceID)
	if err != nil {
		return nil, fmt.Errorf("list evidence bullets: %w", err)
	}
	defer rows.Close()

	bullets := make([]domain.EvidenceBullet, 0)
	for rows.Next() {
		var bullet domain.EvidenceBullet
		if err := rows.Scan(
			&bullet.ID,
			&bullet.Text,
			&bullet.Provenance,
			&bullet.SourceURL,
			&bullet.Verification,
			&bullet.Position,
			&bullet.CreatedAt,
			&bullet.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan evidence bullet: %w", err)
		}
		bullets = append(bullets, bullet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evidence bullets: %w", err)
	}
	return bullets, nil
}

func (s *Store) SaveExperience(ctx context.Context, experience domain.Experience) (domain.Experience, error) {
	if experience.ID == "" {
		experience.ID = uuid.NewString()
	} else if _, err := uuid.Parse(experience.ID); err != nil {
		return domain.Experience{}, fmt.Errorf("experience ID is not valid")
	}
	for index := range experience.Bullets {
		if experience.Bullets[index].ID == "" {
			experience.Bullets[index].ID = uuid.NewString()
		} else if _, err := uuid.Parse(experience.Bullets[index].ID); err != nil {
			return domain.Experience{}, fmt.Errorf("evidence bullet %d ID is not valid", index+1)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Experience{}, fmt.Errorf("begin experience update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	var createdAt string
	var position int
	err = tx.QueryRowContext(ctx, `SELECT created_at, position FROM experiences WHERE id = ?`, experience.ID).Scan(&createdAt, &position)
	if errors.Is(err, sql.ErrNoRows) {
		createdAt = now
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM experiences`).Scan(&position); err != nil {
			return domain.Experience{}, fmt.Errorf("choose experience position: %w", err)
		}
	} else if err != nil {
		return domain.Experience{}, fmt.Errorf("read existing experience: %w", err)
	}

	current := 0
	if experience.Current {
		current = 1
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO experiences (
			id, company, title, location, start_date, end_date, is_current,
			position, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			company = excluded.company,
			title = excluded.title,
			location = excluded.location,
			start_date = excluded.start_date,
			end_date = excluded.end_date,
			is_current = excluded.is_current,
			updated_at = excluded.updated_at
	`,
		experience.ID,
		experience.Company,
		experience.Title,
		experience.Location,
		experience.StartDate,
		experience.EndDate,
		current,
		position,
		createdAt,
		now,
	)
	if err != nil {
		return domain.Experience{}, fmt.Errorf("write experience: %w", err)
	}

	bulletCreatedAt := make(map[string]string, len(experience.Bullets))
	rows, err := tx.QueryContext(ctx, `SELECT id, created_at FROM experience_bullets WHERE experience_id = ?`, experience.ID)
	if err != nil {
		return domain.Experience{}, fmt.Errorf("read existing evidence bullets: %w", err)
	}
	for rows.Next() {
		var id, timestamp string
		if err := rows.Scan(&id, &timestamp); err != nil {
			_ = rows.Close()
			return domain.Experience{}, fmt.Errorf("scan existing evidence bullet: %w", err)
		}
		bulletCreatedAt[id] = timestamp
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return domain.Experience{}, fmt.Errorf("iterate existing evidence bullets: %w", err)
	}
	if err := rows.Close(); err != nil {
		return domain.Experience{}, fmt.Errorf("close existing evidence bullets: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM experience_bullets WHERE experience_id = ?`, experience.ID); err != nil {
		return domain.Experience{}, fmt.Errorf("replace evidence bullets: %w", err)
	}

	for index := range experience.Bullets {
		bullet := &experience.Bullets[index]
		bullet.Position = index
		bullet.CreatedAt = bulletCreatedAt[bullet.ID]
		if bullet.CreatedAt == "" {
			bullet.CreatedAt = now
		}
		bullet.UpdatedAt = now
		_, err := tx.ExecContext(ctx, `
			INSERT INTO experience_bullets (
				id, experience_id, text, provenance, source_url,
				verification_state, position, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			bullet.ID,
			experience.ID,
			bullet.Text,
			bullet.Provenance,
			bullet.SourceURL,
			bullet.Verification,
			bullet.Position,
			bullet.CreatedAt,
			bullet.UpdatedAt,
		)
		if err != nil {
			return domain.Experience{}, fmt.Errorf("write evidence bullet: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Experience{}, fmt.Errorf("commit experience update: %w", err)
	}
	experience.Position = position
	experience.CreatedAt = createdAt
	experience.UpdatedAt = now
	return experience, nil
}

func (s *Store) DeleteExperience(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("experience ID is not valid")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM experiences WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete experience: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if deleted == 0 {
		return fmt.Errorf("experience was not found")
	}
	return nil
}
