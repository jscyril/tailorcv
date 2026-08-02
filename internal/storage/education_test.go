package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestEducationCRUDPreservesOrderAndMetadata(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	first, err := (domain.EducationInput{
		Institution:  "Example Institute",
		Degree:       "Bachelor of Science",
		FieldOfStudy: "Computer Science",
		StartDate:    "2020-08",
		EndDate:      "2024-05",
		Details:      "Graduated with distinction.",
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	savedFirst, err := store.SaveEducation(context.Background(), first)
	if err != nil {
		t.Fatalf("SaveEducation() error = %v", err)
	}
	second, err := (domain.EducationInput{Institution: "Example University", Degree: "Master of Science", Current: true}).Validate()
	if err != nil {
		t.Fatalf("Validate(second) error = %v", err)
	}
	savedSecond, err := store.SaveEducation(context.Background(), second)
	if err != nil {
		t.Fatalf("SaveEducation(second) error = %v", err)
	}

	savedFirst.Degree = "Bachelor of Engineering"
	updated, err := store.SaveEducation(context.Background(), savedFirst)
	if err != nil {
		t.Fatalf("SaveEducation(update) error = %v", err)
	}
	if updated.ID != savedFirst.ID || updated.Position != 0 || updated.CreatedAt != savedFirst.CreatedAt {
		t.Fatalf("updated education = %#v", updated)
	}

	got, err := store.ListEducations(context.Background())
	if err != nil {
		t.Fatalf("ListEducations() error = %v", err)
	}
	if len(got) != 2 || got[0].Degree != "Bachelor of Engineering" || got[1].ID != savedSecond.ID || !got[1].Current {
		t.Fatalf("ListEducations() = %#v", got)
	}

	if err := store.DeleteEducation(context.Background(), savedFirst.ID); err != nil {
		t.Fatalf("DeleteEducation() error = %v", err)
	}
	got, err = store.ListEducations(context.Background())
	if err != nil || len(got) != 1 || got[0].ID != savedSecond.ID {
		t.Fatalf("ListEducations() after delete = %#v, %v", got, err)
	}
}

func TestOpenUpgradesVersionThreeDatabaseForEducation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tailorcv.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatalf("create migration table: %v", err)
	}
	for _, item := range migrations[:3] {
		for _, statement := range item.statements {
			if _, err := db.Exec(statement); err != nil {
				t.Fatalf("apply legacy migration %d: %v", item.version, err)
			}
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?, CURRENT_TIMESTAMP)`, item.version); err != nil {
			t.Fatalf("record legacy migration %d: %v", item.version, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() upgrade error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	education, err := (domain.EducationInput{Institution: "Example Institute", Degree: "Bachelor of Science"}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if _, err := store.SaveEducation(context.Background(), education); err != nil {
		t.Fatalf("SaveEducation() after upgrade error = %v", err)
	}
}
