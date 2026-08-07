package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func OpenDefault() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user configuration directory: %w", err)
	}
	dataDir := filepath.Join(configDir, "tailorcv")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create application data directory: %w", err)
	}
	return Open(filepath.Join(dataDir, "tailorcv.db"))
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.initialize(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initialize(ctx context.Context) error {
	for _, statement := range []string{
		`PRAGMA foreign_keys = ON`,
		`PRAGMA journal_mode = WAL`,
		`PRAGMA busy_timeout = 5000`,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)`,
	} {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
	}
	if err := s.applyMigrations(ctx); err != nil {
		return err
	}
	if err := s.backfillResumeContentHashes(ctx); err != nil {
		return err
	}
	return rebuildEvidenceSearch(ctx, s.db)
}

func (s *Store) backfillResumeContentHashes(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id, latex_source FROM resume_versions WHERE content_hash = ''`)
	if err != nil {
		return fmt.Errorf("find resume versions without content hashes: %w", err)
	}
	type missingHash struct{ id, source string }
	missing := make([]missingHash, 0)
	for rows.Next() {
		var item missingHash
		if err := rows.Scan(&item.id, &item.source); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan resume version without content hash: %w", err)
		}
		missing = append(missing, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("iterate resume versions without content hashes: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close resume content hash rows: %w", err)
	}
	for _, item := range missing {
		if _, err := s.db.ExecContext(ctx, `UPDATE resume_versions SET content_hash = ? WHERE id = ?`, domain.ResumeContentHash(item.source), item.id); err != nil {
			return fmt.Errorf("backfill resume content hash: %w", err)
		}
	}
	return nil
}

func (s *Store) GetProfile(ctx context.Context) (domain.Profile, error) {
	var profile domain.Profile
	err := s.db.QueryRowContext(ctx, `
		SELECT name, headline, email, phone, location, website,
		       github_username, linkedin_url, summary, updated_at
		FROM profiles WHERE id = 1
	`).Scan(
		&profile.Name,
		&profile.Headline,
		&profile.Email,
		&profile.Phone,
		&profile.Location,
		&profile.Website,
		&profile.GitHubUsername,
		&profile.LinkedInURL,
		&profile.Summary,
		&profile.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		profile.Skills = []string{}
		profile.ContactLinks = []domain.ContactLink{}
		return profile, nil
	}
	if err != nil {
		return domain.Profile{}, fmt.Errorf("read profile: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `SELECT name FROM profile_skills ORDER BY position`)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("read profile skills: %w", err)
	}
	profile.Skills = []string{}
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err != nil {
			return domain.Profile{}, fmt.Errorf("scan profile skill: %w", err)
		}
		profile.Skills = append(profile.Skills, skill)
	}
	if err := rows.Err(); err != nil {
		return domain.Profile{}, fmt.Errorf("iterate profile skills: %w", err)
	}
	if err := rows.Close(); err != nil {
		return domain.Profile{}, fmt.Errorf("close profile skills: %w", err)
	}
	linkRows, err := s.db.QueryContext(ctx, `SELECT id, label, url, position, created_at, updated_at FROM contact_links ORDER BY position, created_at`)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("read contact links: %w", err)
	}
	defer linkRows.Close()
	profile.ContactLinks = []domain.ContactLink{}
	for linkRows.Next() {
		var link domain.ContactLink
		if err := linkRows.Scan(&link.ID, &link.Label, &link.URL, &link.Position, &link.CreatedAt, &link.UpdatedAt); err != nil {
			return domain.Profile{}, fmt.Errorf("scan contact link: %w", err)
		}
		profile.ContactLinks = append(profile.ContactLinks, link)
	}
	if err := linkRows.Err(); err != nil {
		return domain.Profile{}, fmt.Errorf("iterate contact links: %w", err)
	}
	return profile, nil
}

func (s *Store) SaveProfile(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("begin profile update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	profile.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO profiles (
			id, name, headline, email, phone, location, website,
			github_username, linkedin_url, summary, updated_at
		) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			headline = excluded.headline,
			email = excluded.email,
			phone = excluded.phone,
			location = excluded.location,
			website = excluded.website,
			github_username = excluded.github_username,
			linkedin_url = excluded.linkedin_url,
			summary = excluded.summary,
			updated_at = excluded.updated_at
	`,
		profile.Name,
		profile.Headline,
		profile.Email,
		profile.Phone,
		profile.Location,
		profile.Website,
		profile.GitHubUsername,
		profile.LinkedInURL,
		profile.Summary,
		profile.UpdatedAt,
	)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("write profile: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM profile_skills`); err != nil {
		return domain.Profile{}, fmt.Errorf("replace profile skills: %w", err)
	}
	for position, skill := range profile.Skills {
		if _, err := tx.ExecContext(ctx, `INSERT INTO profile_skills(position, name) VALUES (?, ?)`, position, skill); err != nil {
			return domain.Profile{}, fmt.Errorf("write profile skill: %w", err)
		}
	}
	createdLinks := make(map[string]string, len(profile.ContactLinks))
	linkRows, err := tx.QueryContext(ctx, `SELECT id, created_at FROM contact_links`)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("read existing contact links: %w", err)
	}
	for linkRows.Next() {
		var id, created string
		if err := linkRows.Scan(&id, &created); err != nil {
			_ = linkRows.Close()
			return domain.Profile{}, fmt.Errorf("scan existing contact link: %w", err)
		}
		createdLinks[id] = created
	}
	if err := linkRows.Err(); err != nil {
		_ = linkRows.Close()
		return domain.Profile{}, fmt.Errorf("iterate existing contact links: %w", err)
	}
	if err := linkRows.Close(); err != nil {
		return domain.Profile{}, fmt.Errorf("close existing contact links: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM contact_links`); err != nil {
		return domain.Profile{}, fmt.Errorf("replace contact links: %w", err)
	}
	for position := range profile.ContactLinks {
		link := &profile.ContactLinks[position]
		if link.ID == "" {
			link.ID = uuid.NewString()
		} else if _, err := uuid.Parse(link.ID); err != nil {
			return domain.Profile{}, fmt.Errorf("contact link %d ID is not valid", position+1)
		}
		link.Position = position
		link.CreatedAt = createdLinks[link.ID]
		if link.CreatedAt == "" {
			link.CreatedAt = profile.UpdatedAt
		}
		link.UpdatedAt = profile.UpdatedAt
		if _, err := tx.ExecContext(ctx, `INSERT INTO contact_links(id,label,url,position,created_at,updated_at) VALUES(?,?,?,?,?,?)`, link.ID, link.Label, link.URL, position, link.CreatedAt, link.UpdatedAt); err != nil {
			return domain.Profile{}, fmt.Errorf("write contact link: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Profile{}, fmt.Errorf("commit profile update: %w", err)
	}
	return profile, nil
}
