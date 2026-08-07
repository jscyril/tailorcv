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

func (s *Store) ListCertifications(ctx context.Context) ([]domain.Certification, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, issuer, issue_date, expiry_date, credential_id, credential_url, description, provenance, verification_state, position, created_at, updated_at FROM certifications ORDER BY position, created_at`)
	if err != nil {
		return nil, fmt.Errorf("list certifications: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Certification, 0)
	for rows.Next() {
		var item domain.Certification
		if err := rows.Scan(&item.ID, &item.Name, &item.Issuer, &item.IssueDate, &item.ExpiryDate, &item.CredentialID, &item.CredentialURL, &item.Description, &item.Provenance, &item.Verification, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan certification: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate certifications: %w", err)
	}
	return items, nil
}

func (s *Store) SaveCertification(ctx context.Context, item domain.Certification) (domain.Certification, error) {
	if item.ID == "" {
		item.ID = uuid.NewString()
	} else if _, err := uuid.Parse(item.ID); err != nil {
		return domain.Certification{}, fmt.Errorf("certification ID is not valid")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var created string
	var position int
	err := s.db.QueryRowContext(ctx, `SELECT created_at, position FROM certifications WHERE id = ?`, item.ID).Scan(&created, &position)
	if errors.Is(err, sql.ErrNoRows) {
		created = now
		err = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM certifications`).Scan(&position)
	}
	if err != nil {
		return domain.Certification{}, fmt.Errorf("read certification: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO certifications(id,name,issuer,issue_date,expiry_date,credential_id,credential_url,description,provenance,verification_state,position,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,issuer=excluded.issuer,issue_date=excluded.issue_date,expiry_date=excluded.expiry_date,credential_id=excluded.credential_id,credential_url=excluded.credential_url,description=excluded.description,provenance=excluded.provenance,verification_state=excluded.verification_state,updated_at=excluded.updated_at`, item.ID, item.Name, item.Issuer, item.IssueDate, item.ExpiryDate, item.CredentialID, item.CredentialURL, item.Description, item.Provenance, item.Verification, position, created, now)
	if err != nil {
		return domain.Certification{}, fmt.Errorf("write certification: %w", err)
	}
	item.Position, item.CreatedAt, item.UpdatedAt = position, created, now
	return item, nil
}

func (s *Store) DeleteCertification(ctx context.Context, id string) error {
	return deleteCredentialRow(ctx, s.db, "certifications", "certification", id)
}

func (s *Store) ListAchievements(ctx context.Context) ([]domain.Achievement, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,title,description,achievement_date,source_url,provenance,verification_state,position,created_at,updated_at FROM achievements ORDER BY position,created_at`)
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Achievement, 0)
	for rows.Next() {
		var item domain.Achievement
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Date, &item.SourceURL, &item.Provenance, &item.Verification, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievements: %w", err)
	}
	return items, nil
}

func (s *Store) SaveAchievement(ctx context.Context, item domain.Achievement) (domain.Achievement, error) {
	if item.ID == "" {
		item.ID = uuid.NewString()
	} else if _, err := uuid.Parse(item.ID); err != nil {
		return domain.Achievement{}, fmt.Errorf("achievement ID is not valid")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var created string
	var position int
	err := s.db.QueryRowContext(ctx, `SELECT created_at,position FROM achievements WHERE id=?`, item.ID).Scan(&created, &position)
	if errors.Is(err, sql.ErrNoRows) {
		created = now
		err = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(position),-1)+1 FROM achievements`).Scan(&position)
	}
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("read achievement: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO achievements(id,title,description,achievement_date,source_url,provenance,verification_state,position,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,description=excluded.description,achievement_date=excluded.achievement_date,source_url=excluded.source_url,provenance=excluded.provenance,verification_state=excluded.verification_state,updated_at=excluded.updated_at`, item.ID, item.Title, item.Description, item.Date, item.SourceURL, item.Provenance, item.Verification, position, created, now)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("write achievement: %w", err)
	}
	item.Position, item.CreatedAt, item.UpdatedAt = position, created, now
	return item, nil
}

func (s *Store) DeleteAchievement(ctx context.Context, id string) error {
	return deleteCredentialRow(ctx, s.db, "achievements", "achievement", id)
}

func deleteCredentialRow(ctx context.Context, db *sql.DB, table, label, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%s ID is not valid", label)
	}
	var statement string
	switch table {
	case "certifications":
		statement = `DELETE FROM certifications WHERE id = ?`
	case "achievements":
		statement = `DELETE FROM achievements WHERE id = ?`
	default:
		return fmt.Errorf("credential type is not valid")
	}
	result, err := db.ExecContext(ctx, statement, id)
	if err != nil {
		return fmt.Errorf("delete %s: %w", label, err)
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return fmt.Errorf("%s was not found", label)
	}
	return nil
}
