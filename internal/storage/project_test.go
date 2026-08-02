package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestDetectedLanguageMigrationUpgradesExistingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	if _, err := database.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatalf("create migration table: %v", err)
	}
	for _, item := range migrations[:5] {
		for _, statement := range item.statements {
			if _, err := database.Exec(statement); err != nil {
				t.Fatalf("apply legacy migration %d: %v", item.version, err)
			}
		}
		if _, err := database.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?, CURRENT_TIMESTAMP)`, item.version); err != nil {
			t.Fatalf("record legacy migration %d: %v", item.version, err)
		}
	}
	if err := database.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open(upgrade) error = %v", err)
	}
	defer store.Close()
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 6`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("language migration count = %d, %v", count, err)
	}
	if _, err := store.db.Exec(`SELECT name, code_bytes FROM project_detected_languages LIMIT 1`); err != nil {
		t.Fatalf("query detected languages after upgrade: %v", err)
	}
}

func TestProjectCRUDPreservesSkillsEvidenceAndReviewState(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	validated, err := (domain.ProjectInput{
		Name:           "Release Console",
		Role:           "Maintainer",
		RepositoryURL:  "https://github.com/example/release-console",
		Provenance:     domain.ProvenanceManual,
		Verification:   domain.VerificationVerified,
		ResumeEligible: true,
		Skills:         []string{"Go", "SQLite"},
		DetectedLanguages: []domain.RepositoryLanguage{
			{Name: "Go", Bytes: 800},
			{Name: "Shell", Bytes: 200},
		},
		Bullets: []domain.EvidenceBulletInput{
			{Text: "Reduced release time", Verification: domain.VerificationVerified},
			{Text: "Added auditable deployments", Verification: domain.VerificationUnverified},
		},
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	saved, err := store.SaveProject(context.Background(), validated)
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}
	if saved.ID == "" || saved.Bullets[0].ID == "" || len(saved.DetectedLanguages) != 2 || saved.CreatedAt == "" {
		t.Fatalf("SaveProject() = %#v", saved)
	}

	firstID := saved.Bullets[0].ID
	saved.Bullets[0], saved.Bullets[1] = saved.Bullets[1], saved.Bullets[0]
	saved.ResumeEligible = false
	saved.Skills = []string{"Go", "React"}
	updated, err := store.SaveProject(context.Background(), saved)
	if err != nil {
		t.Fatalf("SaveProject(update) error = %v", err)
	}
	if updated.Bullets[1].ID != firstID || updated.ResumeEligible || updated.Skills[1] != "React" {
		t.Fatalf("updated project = %#v", updated)
	}

	got, err := store.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(got) != 1 || got[0].Bullets[1].ID != firstID || got[0].Skills[1] != "React" || len(got[0].DetectedLanguages) != 2 || got[0].DetectedLanguages[1].Name != "Shell" || got[0].ResumeEligible {
		t.Fatalf("ListProjects() = %#v", got)
	}

	if err := store.DeleteProject(context.Background(), saved.ID); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	var skills, languages, bullets int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM project_skills`).Scan(&skills); err != nil {
		t.Fatalf("count project skills after delete: %v", err)
	}
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM project_bullets`).Scan(&bullets); err != nil {
		t.Fatalf("count project bullets after delete: %v", err)
	}
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM project_detected_languages`).Scan(&languages); err != nil {
		t.Fatalf("count project detected languages after delete: %v", err)
	}
	if skills != 0 || languages != 0 || bullets != 0 {
		t.Fatalf("project children after delete: skills=%d languages=%d bullets=%d", skills, languages, bullets)
	}
}
