package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func TestGeneralResumeRenameDuplicateAndDelete(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()

	general, err := store.CreateResume(ctx, domain.CreateResumeInput{Name: "  Backend   resume ", TemplateID: "builtin-jake-style", LatexSource: "general source"})
	if err != nil {
		t.Fatalf("CreateResume() error = %v", err)
	}
	if general.Application.Name != "Backend resume" || general.Application.JobID != "" || general.Version.VersionNumber != 1 {
		t.Fatalf("general resume = %#v", general)
	}
	if _, err := store.CreateResume(ctx, domain.CreateResumeInput{Name: " ", TemplateID: "builtin-jake-style", LatexSource: "x"}); err == nil {
		t.Fatal("CreateResume() expected a name error")
	}
	edited, err := store.CreateEditedResumeVersion(ctx, domain.SaveResumeVersionEditInput{ApplicationID: general.Application.ID, BaseVersionID: general.Version.ID, LatexSource: "general source v2"})
	if err != nil || edited.VersionNumber != 2 {
		t.Fatalf("CreateEditedResumeVersion() = %#v, %v", edited, err)
	}

	renamed, err := store.RenameApplication(ctx, domain.RenameResumeInput{ApplicationID: general.Application.ID, Name: "Go resume"})
	if err != nil || renamed.Name != "Go resume" || len(renamed.Versions) != 2 {
		t.Fatalf("RenameApplication() = %#v, %v", renamed, err)
	}
	if _, err := store.RenameApplication(ctx, domain.RenameResumeInput{ApplicationID: uuid.NewString(), Name: "Missing"}); err == nil {
		t.Fatal("RenameApplication() expected missing resume error")
	}

	// A job-tailored resume without a name is copied under its job's title.
	job, _ := (domain.JobInput{Role: "Analyst", Description: "Analyze reliable distributed services with Go and PostgreSQL for production workloads."}).Validate()
	job, err = store.SaveJob(ctx, job)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	fact := uuid.NewString()
	tailored, err := store.CreateResumeVersion(ctx, job.ID, []string{fact}, "builtin-jake-style", job.Description, "tailored v1")
	if err != nil {
		t.Fatalf("CreateResumeVersion() error = %v", err)
	}
	if _, err := store.CreateResumeVersion(ctx, job.ID, []string{fact}, "builtin-jake-style", job.Description, "tailored v2"); err != nil {
		t.Fatalf("CreateResumeVersion(v2) error = %v", err)
	}
	if err := store.RecordResumeCompilation(ctx, tailored.Version.ID, "/tmp/not-copied.pdf", domain.CompileResult{Success: true, Engine: "recorded"}); err != nil {
		t.Fatalf("RecordResumeCompilation() error = %v", err)
	}
	copied, err := store.DuplicateApplication(ctx, tailored.Application.ID)
	if err != nil {
		t.Fatalf("DuplicateApplication() error = %v", err)
	}
	if copied.Name != "Copy of Analyst" || copied.JobID != "" || len(copied.Versions) != 1 || copied.Versions[0].LatexSource != "tailored v2" || copied.Versions[0].VersionNumber != 1 || copied.Versions[0].CompileSuccess || copied.Versions[0].JobDescriptionSnapshot != job.Description || copied.SelectedFactIDs[0] != fact {
		t.Fatalf("copy = %#v", copied)
	}

	paths, err := store.DeleteApplication(ctx, tailored.Application.ID)
	if err != nil || len(paths) != 1 || paths[0] != "/tmp/not-copied.pdf" {
		t.Fatalf("DeleteApplication() = %v, %v", paths, err)
	}
	if _, err := store.GetJob(ctx, job.ID); err != nil {
		t.Fatalf("job should survive resume delete: %v", err)
	}
	if _, err := store.GetResumeVersion(ctx, tailored.Version.ID); err == nil {
		t.Fatal("deleted resume versions should be gone")
	}
	if _, err := store.DeleteApplication(ctx, tailored.Application.ID); err == nil {
		t.Fatal("DeleteApplication() expected missing resume error")
	}
	applications, err := store.ListApplications(ctx)
	if err != nil || len(applications) != 2 {
		t.Fatalf("remaining resumes = %#v, %v", applications, err)
	}
	// A job can be tailored again after its resume was deleted.
	if _, err := store.CreateResumeVersion(ctx, job.ID, []string{fact}, "builtin-jake-style", job.Description, "retailored"); err != nil {
		t.Fatalf("CreateResumeVersion(after delete) error = %v", err)
	}
}

func TestResumeLibraryMigrationKeepsExistingResumes(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "upgrade.db"))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	for _, statement := range []string{`PRAGMA foreign_keys = ON`, `CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("prepare database: %v", err)
		}
	}
	store := &Store{db: db}
	if err := store.applyMigrationSet(ctx, migrations[:len(migrations)-1]); err != nil {
		t.Fatalf("apply previous migrations: %v", err)
	}
	jobID, applicationID, versionID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, statement := range []string{
		`INSERT INTO jobs(id, company, role, description, created_at, updated_at) VALUES ('` + jobID + `', 'Acme', 'Engineer', 'desc', 't', 't')`,
		`INSERT INTO applications(id, job_id, status, created_at, updated_at) VALUES ('` + applicationID + `', '` + jobID + `', 'submitted', 't', 't')`,
		`INSERT INTO resume_versions(id, application_id, version_number, job_description_snapshot, selected_fact_ids_json, latex_source, template_id, created_at) VALUES ('` + versionID + `', '` + applicationID + `', 1, 'desc', '[]', 'kept source', 'builtin-jake-style', 't')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("seed old schema: %v", err)
		}
	}
	if err := store.applyMigrationSet(ctx, migrations); err != nil {
		t.Fatalf("apply resume library migration: %v", err)
	}
	applications, err := store.ListApplications(ctx)
	if err != nil || len(applications) != 1 || applications[0].JobID != jobID || applications[0].Status != domain.ApplicationStatusSubmitted || len(applications[0].Versions) != 1 || applications[0].Versions[0].LatexSource != "kept source" {
		t.Fatalf("migrated applications = %#v, %v", applications, err)
	}
	if _, err := db.Exec(`INSERT INTO resume_versions(id, application_id, version_number, job_description_snapshot, selected_fact_ids_json, latex_source, template_id, created_at) VALUES (?, ?, 1, '', '[]', 'x', 'x', 't')`, uuid.NewString(), uuid.NewString()); err == nil || !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Fatalf("foreign keys should be enforced again after migration, got %v", err)
	}
	if _, err := db.Exec(`INSERT INTO applications(id, job_id, status, created_at, updated_at) VALUES (?, ?, 'draft', 't', 't')`, uuid.NewString(), jobID); err == nil {
		t.Fatal("a job should still have at most one tailored resume")
	}
	if _, err := db.Exec(`DELETE FROM jobs WHERE id = ?`, jobID); err != nil {
		t.Fatalf("delete job: %v", err)
	}
	if applications, _ := store.ListApplications(ctx); len(applications) != 0 {
		t.Fatalf("deleting a job should still remove its tailored resume: %#v", applications)
	}
}

func TestBackupRoundTripKeepsGeneralResumes(t *testing.T) {
	ctx := context.Background()
	source, err := Open(filepath.Join(t.TempDir(), "source.db"))
	if err != nil {
		t.Fatalf("Open(source) error = %v", err)
	}
	t.Cleanup(func() { _ = source.Close() })
	created, err := source.CreateResume(ctx, domain.CreateResumeInput{Name: "General", TemplateID: "builtin-jake-style", LatexSource: "general source"})
	if err != nil {
		t.Fatalf("CreateResume() error = %v", err)
	}
	backup, err := source.CreateProfileBackup(ctx)
	if err != nil {
		t.Fatalf("CreateProfileBackup() error = %v", err)
	}
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("encode backup: %v", err)
	}
	decoded, err := domain.DecodeProfileBackup(data)
	if err != nil {
		t.Fatalf("DecodeProfileBackup() error = %v", err)
	}
	target, err := Open(filepath.Join(t.TempDir(), "target.db"))
	if err != nil {
		t.Fatalf("Open(target) error = %v", err)
	}
	t.Cleanup(func() { _ = target.Close() })
	if err := target.ReplaceProfileFromBackup(ctx, decoded); err != nil {
		t.Fatalf("ReplaceProfileFromBackup() error = %v", err)
	}
	applications, err := target.ListApplications(ctx)
	if err != nil || len(applications) != 1 || applications[0].ID != created.Application.ID || applications[0].Name != "General" || applications[0].JobID != "" || applications[0].Versions[0].LatexSource != "general source" {
		t.Fatalf("restored resumes = %#v, %v", applications, err)
	}
}
