package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestProfileBackupRoundTripReplacesAllCurrentData(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()

	profile, err := (domain.ProfileInput{Name: "Ada Lovelace", Email: "ada@example.com", Skills: []string{"Go", "SQLite"}}).Validate()
	if err != nil {
		t.Fatalf("profile Validate() error = %v", err)
	}
	if _, err := store.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	experience, err := (domain.ExperienceInput{Company: "Example Systems", Title: "Engineer", StartDate: "2023-01", Bullets: []domain.EvidenceBulletInput{{Text: "Built an audited release pipeline", Verification: domain.VerificationVerified}}}).Validate()
	if err != nil {
		t.Fatalf("experience Validate() error = %v", err)
	}
	if _, err := store.SaveExperience(ctx, experience); err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	project, err := (domain.ProjectInput{Name: "Release Console", Skills: []string{"Go"}, DetectedLanguages: []domain.RepositoryLanguage{{Name: "Go", Bytes: 900}, {Name: "Shell", Bytes: 100}}, ResumeEligible: true}).Validate()
	if err != nil {
		t.Fatalf("project Validate() error = %v", err)
	}
	if _, err := store.SaveProject(ctx, project); err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}
	education, err := (domain.EducationInput{Institution: "Example Institute", Degree: "Bachelor of Science"}).Validate()
	if err != nil {
		t.Fatalf("education Validate() error = %v", err)
	}
	if _, err := store.SaveEducation(ctx, education); err != nil {
		t.Fatalf("SaveEducation() error = %v", err)
	}
	job, err := (domain.JobInput{Company: "Example Systems", Role: "Platform Engineer", Description: "Build reliable Go services and audited deployment pipelines for a growing platform."}).Validate()
	if err != nil {
		t.Fatalf("job Validate() error = %v", err)
	}
	if _, err := store.SaveJob(ctx, job); err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}

	backup, err := store.CreateProfileBackup(ctx)
	if err != nil {
		t.Fatalf("CreateProfileBackup() error = %v", err)
	}
	replacement, err := (domain.ProfileInput{Name: "Temporary User"}).Validate()
	if err != nil {
		t.Fatalf("replacement Validate() error = %v", err)
	}
	if _, err := store.SaveProfile(ctx, replacement); err != nil {
		t.Fatalf("SaveProfile(replacement) error = %v", err)
	}
	if err := store.ReplaceProfileFromBackup(ctx, backup); err != nil {
		t.Fatalf("ReplaceProfileFromBackup() error = %v", err)
	}

	got, err := store.CreateProfileBackup(ctx)
	if err != nil {
		t.Fatalf("CreateProfileBackup(after import) error = %v", err)
	}
	if got.Profile.Name != "Ada Lovelace" || len(got.Experiences) != 1 || len(got.Projects) != 1 || len(got.Educations) != 1 || len(got.Jobs) != 1 {
		t.Fatalf("restored backup = %#v", got)
	}
	if got.Experiences[0].Bullets[0].ID != backup.Experiences[0].Bullets[0].ID || got.Projects[0].Skills[0] != "Go" || len(got.Projects[0].DetectedLanguages) != 2 || got.Projects[0].DetectedLanguages[1].Name != "Shell" {
		t.Fatalf("restored child data = %#v", got)
	}
}

func TestProfileBackupImportIsAtomicOnInvalidID(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	profile, _ := (domain.ProfileInput{Name: "Keep Me"}).Validate()
	if _, err := store.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	backup, err := store.CreateProfileBackup(ctx)
	if err != nil {
		t.Fatalf("CreateProfileBackup() error = %v", err)
	}
	backup.Experiences = []domain.Experience{{ID: "not-a-uuid", Company: "Example", Title: "Engineer", StartDate: "2024-01", CreatedAt: backup.ExportedAt, UpdatedAt: backup.ExportedAt, Bullets: []domain.EvidenceBullet{}}}
	if err := store.ReplaceProfileFromBackup(ctx, backup); err == nil {
		t.Fatal("ReplaceProfileFromBackup() expected invalid ID error")
	}
	got, err := store.GetProfile(ctx)
	if err != nil || got.Name != "Keep Me" {
		t.Fatalf("profile after rejected import = %#v, %v", got, err)
	}
}
