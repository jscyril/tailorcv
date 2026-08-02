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

const selectedTemplateSetting = "selected_template_id"

func (s *Store) ListTemplates(ctx context.Context) ([]domain.ResumeTemplate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, source, created_at, updated_at FROM resume_templates ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("list resume templates: %w", err)
	}
	defer rows.Close()
	templates := []domain.ResumeTemplate{}
	for rows.Next() {
		var template domain.ResumeTemplate
		if err := rows.Scan(&template.ID, &template.Name, &template.Description, &template.Source, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan resume template: %w", err)
		}
		templates = append(templates, template)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resume templates: %w", err)
	}
	return templates, nil
}

func (s *Store) SaveTemplate(ctx context.Context, template domain.ResumeTemplate) (domain.ResumeTemplate, error) {
	if template.ID == "" {
		template.ID = uuid.NewString()
	} else if _, err := uuid.Parse(template.ID); err != nil {
		return domain.ResumeTemplate{}, fmt.Errorf("template ID is not valid")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var createdAt string
	err := s.db.QueryRowContext(ctx, `SELECT created_at FROM resume_templates WHERE id = ?`, template.ID).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		createdAt = now
	} else if err != nil {
		return domain.ResumeTemplate{}, fmt.Errorf("read existing resume template: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO resume_templates (id, name, description, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, description = excluded.description,
			source = excluded.source, updated_at = excluded.updated_at
	`, template.ID, template.Name, template.Description, template.Source, createdAt, now)
	if err != nil {
		return domain.ResumeTemplate{}, fmt.Errorf("save resume template: %w", err)
	}
	template.CreatedAt, template.UpdatedAt = createdAt, now
	return template, nil
}

func (s *Store) DeleteTemplate(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("template ID is not valid")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM resume_templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete resume template: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("template was not found")
	}
	return nil
}

func (s *Store) SelectedTemplateID(ctx context.Context) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM app_settings WHERE key = ?`, selectedTemplateSetting).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read selected template: %w", err)
	}
	return id, nil
}

func (s *Store) SetSelectedTemplateID(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO app_settings(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, selectedTemplateSetting, id)
	if err != nil {
		return fmt.Errorf("save selected template: %w", err)
	}
	return nil
}
