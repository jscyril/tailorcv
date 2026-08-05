package storage

import (
	"context"
	"fmt"
)

type migration struct {
	version    int
	statements []string
}

var migrations = []migration{
	{
		version: 1,
		statements: []string{
			`CREATE TABLE profiles (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				name TEXT NOT NULL DEFAULT '',
				headline TEXT NOT NULL DEFAULT '',
				email TEXT NOT NULL DEFAULT '',
				phone TEXT NOT NULL DEFAULT '',
				location TEXT NOT NULL DEFAULT '',
				website TEXT NOT NULL DEFAULT '',
				github_username TEXT NOT NULL DEFAULT '',
				linkedin_url TEXT NOT NULL DEFAULT '',
				summary TEXT NOT NULL DEFAULT '',
				updated_at TEXT NOT NULL
			)`,
			`CREATE TABLE profile_skills (
				position INTEGER PRIMARY KEY,
				name TEXT NOT NULL COLLATE NOCASE UNIQUE
			)`,
		},
	},
	{
		version: 2,
		statements: []string{
			`CREATE TABLE experiences (
				id TEXT PRIMARY KEY,
				company TEXT NOT NULL,
				title TEXT NOT NULL,
				location TEXT NOT NULL DEFAULT '',
				start_date TEXT NOT NULL,
				end_date TEXT NOT NULL DEFAULT '',
				is_current INTEGER NOT NULL DEFAULT 0 CHECK (is_current IN (0, 1)),
				position INTEGER NOT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX experiences_position_idx ON experiences(position, created_at)`,
			`CREATE TABLE experience_bullets (
				id TEXT PRIMARY KEY,
				experience_id TEXT NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
				text TEXT NOT NULL,
				provenance TEXT NOT NULL CHECK (provenance IN ('manual', 'github', 'imported')),
				source_url TEXT NOT NULL DEFAULT '',
				verification_state TEXT NOT NULL CHECK (verification_state IN ('unverified', 'verified')),
				position INTEGER NOT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX experience_bullets_order_idx ON experience_bullets(experience_id, position)`,
		},
	},
	{
		version: 3,
		statements: []string{
			`CREATE TABLE projects (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				role TEXT NOT NULL DEFAULT '',
				description TEXT NOT NULL DEFAULT '',
				url TEXT NOT NULL DEFAULT '',
				repository_url TEXT NOT NULL DEFAULT '',
				start_date TEXT NOT NULL DEFAULT '',
				end_date TEXT NOT NULL DEFAULT '',
				is_ongoing INTEGER NOT NULL DEFAULT 0 CHECK (is_ongoing IN (0, 1)),
				provenance TEXT NOT NULL CHECK (provenance IN ('manual', 'github', 'imported')),
				verification_state TEXT NOT NULL CHECK (verification_state IN ('unverified', 'verified')),
				resume_eligible INTEGER NOT NULL DEFAULT 0 CHECK (resume_eligible IN (0, 1)),
				position INTEGER NOT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX projects_position_idx ON projects(position, created_at)`,
			`CREATE TABLE project_skills (
				project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
				position INTEGER NOT NULL,
				name TEXT NOT NULL COLLATE NOCASE,
				PRIMARY KEY (project_id, position),
				UNIQUE (project_id, name)
			)`,
			`CREATE TABLE project_bullets (
				id TEXT PRIMARY KEY,
				project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
				text TEXT NOT NULL,
				provenance TEXT NOT NULL CHECK (provenance IN ('manual', 'github', 'imported')),
				source_url TEXT NOT NULL DEFAULT '',
				verification_state TEXT NOT NULL CHECK (verification_state IN ('unverified', 'verified')),
				position INTEGER NOT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX project_bullets_order_idx ON project_bullets(project_id, position)`,
		},
	},
	{
		version: 4,
		statements: []string{
			`CREATE TABLE educations (
				id TEXT PRIMARY KEY,
				institution TEXT NOT NULL,
				degree TEXT NOT NULL,
				field_of_study TEXT NOT NULL DEFAULT '',
				location TEXT NOT NULL DEFAULT '',
				start_date TEXT NOT NULL DEFAULT '',
				end_date TEXT NOT NULL DEFAULT '',
				is_current INTEGER NOT NULL DEFAULT 0 CHECK (is_current IN (0, 1)),
				details TEXT NOT NULL DEFAULT '',
				position INTEGER NOT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX educations_position_idx ON educations(position, created_at)`,
		},
	},
	{
		version: 5,
		statements: []string{
			`CREATE TABLE resume_templates (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				description TEXT NOT NULL DEFAULT '',
				source TEXT NOT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX resume_templates_name_idx ON resume_templates(name COLLATE NOCASE)`,
			`CREATE TABLE app_settings (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL
			)`,
		},
	},
	{
		version: 6,
		statements: []string{
			`CREATE TABLE project_detected_languages (
				project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
				position INTEGER NOT NULL,
				name TEXT NOT NULL COLLATE NOCASE,
				code_bytes INTEGER NOT NULL DEFAULT 0 CHECK (code_bytes >= 0),
				PRIMARY KEY (project_id, position),
				UNIQUE (project_id, name)
			)`,
		},
	},
	{
		version: 7,
		statements: []string{
			`CREATE TABLE jobs (
				id TEXT PRIMARY KEY,
				company TEXT NOT NULL DEFAULT '',
				role TEXT NOT NULL DEFAULT '',
				description TEXT NOT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX jobs_updated_at_idx ON jobs(updated_at DESC)`,
		},
	},
	{
		version: 8,
		statements: []string{
			`CREATE TABLE applications (
				id TEXT PRIMARY KEY,
				job_id TEXT NOT NULL UNIQUE REFERENCES jobs(id) ON DELETE CASCADE,
				status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'archived')),
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)`,
			`CREATE TABLE application_selected_facts (
				application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
				position INTEGER NOT NULL,
				fact_id TEXT NOT NULL,
				PRIMARY KEY (application_id, position),
				UNIQUE (application_id, fact_id)
			)`,
			`CREATE TABLE resume_versions (
				id TEXT PRIMARY KEY,
				application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
				version_number INTEGER NOT NULL CHECK (version_number > 0),
				job_description_snapshot TEXT NOT NULL,
				selected_fact_ids_json TEXT NOT NULL,
				latex_source TEXT NOT NULL,
				template_id TEXT NOT NULL,
				created_at TEXT NOT NULL,
				UNIQUE (application_id, version_number)
			)`,
			`CREATE INDEX resume_versions_application_idx ON resume_versions(application_id, version_number DESC)`,
		},
	},
	{
		version: 9,
		statements: []string{
			`CREATE VIRTUAL TABLE career_evidence_fts USING fts5(
				fact_id UNINDEXED,
				source_id UNINDEXED,
				source_type UNINDEXED,
				source_label,
				text,
				skills,
				tokenize = 'porter unicode61'
			)`,
		},
	},
	{
		version: 10,
		statements: []string{
			`ALTER TABLE resume_versions ADD COLUMN ranking_explanations_json TEXT NOT NULL DEFAULT '[]'`,
			`ALTER TABLE resume_versions ADD COLUMN content_hash TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE resume_versions ADD COLUMN compile_success INTEGER NOT NULL DEFAULT 0 CHECK (compile_success IN (0, 1))`,
			`ALTER TABLE resume_versions ADD COLUMN compile_engine TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE resume_versions ADD COLUMN compile_duration_ms INTEGER NOT NULL DEFAULT 0 CHECK (compile_duration_ms >= 0)`,
			`ALTER TABLE resume_versions ADD COLUMN compile_diagnostics_json TEXT NOT NULL DEFAULT '[]'`,
			`ALTER TABLE resume_versions ADD COLUMN compiled_at TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE resume_versions ADD COLUMN pdf_path TEXT NOT NULL DEFAULT ''`,
		},
	},
}

func (s *Store) applyMigrations(ctx context.Context) error {
	for _, item := range migrations {
		var applied int
		err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, item.version).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", item.version, err)
		}
		if applied > 0 {
			continue
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", item.version, err)
		}
		for _, statement := range item.statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("apply migration %d: %w", item.version, err)
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES (?, CURRENT_TIMESTAMP)`, item.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", item.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", item.version, err)
		}
	}
	return nil
}
