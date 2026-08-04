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

func (s *Store) ListProjects(ctx context.Context) ([]domain.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, role, description, url, repository_url, start_date, end_date,
		       is_ongoing, provenance, verification_state, resume_eligible,
		       position, created_at, updated_at
		FROM projects
		ORDER BY position, created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	projects := make([]domain.Project, 0)
	for rows.Next() {
		var project domain.Project
		var ongoing, eligible int
		if err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Role,
			&project.Description,
			&project.URL,
			&project.RepositoryURL,
			&project.StartDate,
			&project.EndDate,
			&ongoing,
			&project.Provenance,
			&project.Verification,
			&eligible,
			&project.Position,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan project: %w", err)
		}
		project.Ongoing = ongoing == 1
		project.ResumeEligible = eligible == 1
		project.Skills = []string{}
		project.Bullets = []domain.EvidenceBullet{}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close project rows: %w", err)
	}

	for index := range projects {
		skills, err := s.listProjectSkills(ctx, projects[index].ID)
		if err != nil {
			return nil, err
		}
		projects[index].Skills = skills
		languages, err := s.listProjectDetectedLanguages(ctx, projects[index].ID)
		if err != nil {
			return nil, err
		}
		projects[index].DetectedLanguages = languages
		bullets, err := s.listProjectBullets(ctx, projects[index].ID)
		if err != nil {
			return nil, err
		}
		projects[index].Bullets = bullets
	}
	return projects, nil
}

func (s *Store) listProjectDetectedLanguages(ctx context.Context, projectID string) ([]domain.RepositoryLanguage, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, code_bytes FROM project_detected_languages WHERE project_id = ? ORDER BY position`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project detected languages: %w", err)
	}
	defer rows.Close()
	languages := make([]domain.RepositoryLanguage, 0)
	for rows.Next() {
		var language domain.RepositoryLanguage
		if err := rows.Scan(&language.Name, &language.Bytes); err != nil {
			return nil, fmt.Errorf("scan project detected language: %w", err)
		}
		languages = append(languages, language)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project detected languages: %w", err)
	}
	return languages, nil
}

func (s *Store) listProjectSkills(ctx context.Context, projectID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name FROM project_skills WHERE project_id = ? ORDER BY position`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project skills: %w", err)
	}
	defer rows.Close()

	skills := make([]string, 0)
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err != nil {
			return nil, fmt.Errorf("scan project skill: %w", err)
		}
		skills = append(skills, skill)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project skills: %w", err)
	}
	return skills, nil
}

func (s *Store) listProjectBullets(ctx context.Context, projectID string) ([]domain.EvidenceBullet, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, text, provenance, source_url, verification_state, position, created_at, updated_at
		FROM project_bullets
		WHERE project_id = ?
		ORDER BY position
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project evidence bullets: %w", err)
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
			return nil, fmt.Errorf("scan project evidence bullet: %w", err)
		}
		bullets = append(bullets, bullet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project evidence bullets: %w", err)
	}
	return bullets, nil
}

func (s *Store) SaveProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	if project.ID == "" {
		project.ID = uuid.NewString()
	} else if _, err := uuid.Parse(project.ID); err != nil {
		return domain.Project{}, fmt.Errorf("project ID is not valid")
	}
	for index := range project.Bullets {
		if project.Bullets[index].ID == "" {
			project.Bullets[index].ID = uuid.NewString()
		} else if _, err := uuid.Parse(project.Bullets[index].ID); err != nil {
			return domain.Project{}, fmt.Errorf("evidence bullet %d ID is not valid", index+1)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Project{}, fmt.Errorf("begin project update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	var createdAt string
	var position int
	err = tx.QueryRowContext(ctx, `SELECT created_at, position FROM projects WHERE id = ?`, project.ID).Scan(&createdAt, &position)
	if errors.Is(err, sql.ErrNoRows) {
		createdAt = now
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM projects`).Scan(&position); err != nil {
			return domain.Project{}, fmt.Errorf("choose project position: %w", err)
		}
	} else if err != nil {
		return domain.Project{}, fmt.Errorf("read existing project: %w", err)
	}

	ongoing := boolInt(project.Ongoing)
	eligible := boolInt(project.ResumeEligible)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO projects (
			id, name, role, description, url, repository_url, start_date, end_date,
			is_ongoing, provenance, verification_state, resume_eligible,
			position, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			role = excluded.role,
			description = excluded.description,
			url = excluded.url,
			repository_url = excluded.repository_url,
			start_date = excluded.start_date,
			end_date = excluded.end_date,
			is_ongoing = excluded.is_ongoing,
			provenance = excluded.provenance,
			verification_state = excluded.verification_state,
			resume_eligible = excluded.resume_eligible,
			updated_at = excluded.updated_at
	`,
		project.ID,
		project.Name,
		project.Role,
		project.Description,
		project.URL,
		project.RepositoryURL,
		project.StartDate,
		project.EndDate,
		ongoing,
		project.Provenance,
		project.Verification,
		eligible,
		position,
		createdAt,
		now,
	)
	if err != nil {
		return domain.Project{}, fmt.Errorf("write project: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM project_skills WHERE project_id = ?`, project.ID); err != nil {
		return domain.Project{}, fmt.Errorf("replace project skills: %w", err)
	}
	for index, skill := range project.Skills {
		if _, err := tx.ExecContext(ctx, `INSERT INTO project_skills(project_id, position, name) VALUES (?, ?, ?)`, project.ID, index, skill); err != nil {
			return domain.Project{}, fmt.Errorf("write project skill: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_detected_languages WHERE project_id = ?`, project.ID); err != nil {
		return domain.Project{}, fmt.Errorf("replace project detected languages: %w", err)
	}
	for index, language := range project.DetectedLanguages {
		if _, err := tx.ExecContext(ctx, `INSERT INTO project_detected_languages(project_id, position, name, code_bytes) VALUES (?, ?, ?, ?)`, project.ID, index, language.Name, language.Bytes); err != nil {
			return domain.Project{}, fmt.Errorf("write project detected language: %w", err)
		}
	}

	bulletCreatedAt := make(map[string]string, len(project.Bullets))
	rows, err := tx.QueryContext(ctx, `SELECT id, created_at FROM project_bullets WHERE project_id = ?`, project.ID)
	if err != nil {
		return domain.Project{}, fmt.Errorf("read existing project evidence bullets: %w", err)
	}
	for rows.Next() {
		var id, timestamp string
		if err := rows.Scan(&id, &timestamp); err != nil {
			_ = rows.Close()
			return domain.Project{}, fmt.Errorf("scan existing project evidence bullet: %w", err)
		}
		bulletCreatedAt[id] = timestamp
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return domain.Project{}, fmt.Errorf("iterate existing project evidence bullets: %w", err)
	}
	if err := rows.Close(); err != nil {
		return domain.Project{}, fmt.Errorf("close existing project evidence bullets: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_bullets WHERE project_id = ?`, project.ID); err != nil {
		return domain.Project{}, fmt.Errorf("replace project evidence bullets: %w", err)
	}

	for index := range project.Bullets {
		bullet := &project.Bullets[index]
		bullet.Position = index
		bullet.CreatedAt = bulletCreatedAt[bullet.ID]
		if bullet.CreatedAt == "" {
			bullet.CreatedAt = now
		}
		bullet.UpdatedAt = now
		_, err := tx.ExecContext(ctx, `
			INSERT INTO project_bullets (
				id, project_id, text, provenance, source_url,
				verification_state, position, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			bullet.ID,
			project.ID,
			bullet.Text,
			bullet.Provenance,
			bullet.SourceURL,
			bullet.Verification,
			bullet.Position,
			bullet.CreatedAt,
			bullet.UpdatedAt,
		)
		if err != nil {
			return domain.Project{}, fmt.Errorf("write project evidence bullet: %w", err)
		}
	}

	if err := rebuildEvidenceSearch(ctx, tx); err != nil {
		return domain.Project{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Project{}, fmt.Errorf("commit project update: %w", err)
	}
	project.Position = position
	project.CreatedAt = createdAt
	project.UpdatedAt = now
	return project, nil
}

func (s *Store) DeleteProject(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("project ID is not valid")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin project delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if deleted == 0 {
		return fmt.Errorf("project was not found")
	}
	if err := rebuildEvidenceSearch(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit project delete: %w", err)
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
