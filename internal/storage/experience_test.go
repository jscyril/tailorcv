package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestExperienceCRUDPreservesEvidenceIDsAndOrder(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	validated, err := (domain.ExperienceInput{
		Company:   "Example Systems",
		Title:     "Engineer",
		StartDate: "2023-01",
		Current:   true,
		Bullets: []domain.EvidenceBulletInput{
			{Text: "Reduced deployment time", Verification: domain.VerificationVerified, Importance: domain.EvidenceImportanceEssential},
			{Text: "Built the release service", Verification: domain.VerificationUnverified},
		},
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	saved, err := store.SaveExperience(context.Background(), validated)
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	if saved.ID == "" || saved.Bullets[0].ID == "" || saved.CreatedAt == "" {
		t.Fatalf("SaveExperience() = %#v", saved)
	}

	firstID := saved.Bullets[0].ID
	saved.Bullets[0], saved.Bullets[1] = saved.Bullets[1], saved.Bullets[0]
	saved.Title = "Senior Engineer"
	updated, err := store.SaveExperience(context.Background(), saved)
	if err != nil {
		t.Fatalf("SaveExperience(update) error = %v", err)
	}
	if updated.Bullets[1].ID != firstID || updated.Bullets[1].Position != 1 {
		t.Fatalf("updated bullets = %#v", updated.Bullets)
	}

	got, err := store.ListExperiences(context.Background())
	if err != nil {
		t.Fatalf("ListExperiences() error = %v", err)
	}
	if len(got) != 1 || got[0].Title != "Senior Engineer" || got[0].Bullets[1].ID != firstID || got[0].Bullets[1].Importance != domain.EvidenceImportanceEssential {
		t.Fatalf("ListExperiences() = %#v", got)
	}

	if err := store.DeleteExperience(context.Background(), saved.ID); err != nil {
		t.Fatalf("DeleteExperience() error = %v", err)
	}
	var remainingBullets int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM experience_bullets`).Scan(&remainingBullets); err != nil {
		t.Fatalf("count evidence bullets after delete: %v", err)
	}
	if remainingBullets != 0 {
		t.Fatalf("evidence bullets after delete = %d", remainingBullets)
	}
	got, err = store.ListExperiences(context.Background())
	if err != nil || len(got) != 0 {
		t.Fatalf("ListExperiences() after delete = %#v, %v", got, err)
	}
}

func TestOpenAppliesAllMigrations(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != len(migrations) {
		t.Fatalf("migration count = %d, want %d", count, len(migrations))
	}
}

func TestOpenUpgradesVersionOneDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tailorcv.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatalf("create legacy migration table: %v", err)
	}
	for _, statement := range migrations[0].statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create version one schema: %v", err)
		}
	}
	if _, err := db.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (1, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("record legacy migration: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() upgrade error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	validated, err := (domain.ExperienceInput{Company: "Example", Title: "Engineer", StartDate: "2024-01"}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if _, err := store.SaveExperience(context.Background(), validated); err != nil {
		t.Fatalf("SaveExperience() after upgrade error = %v", err)
	}
}
