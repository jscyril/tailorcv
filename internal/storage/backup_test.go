package storage

import (
	"context"
	"path/filepath"
	"strings"
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

	profile, err := (domain.ProfileInput{Name: "Ada Lovelace", Email: "ada@example.com", Skills: []string{"Go", "SQLite"}, ContactLinks: []domain.ContactLinkInput{{Label: "Portfolio", URL: "https://example.com/work"}}}).Validate()
	if err != nil {
		t.Fatalf("profile Validate() error = %v", err)
	}
	if _, err := store.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	experience, err := (domain.ExperienceInput{Company: "Example Systems", Title: "Engineer", StartDate: "2023-01", Bullets: []domain.EvidenceBulletInput{{Text: "Built an audited release pipeline", Verification: domain.VerificationVerified, Importance: domain.EvidenceImportanceEssential}}}).Validate()
	if err != nil {
		t.Fatalf("experience Validate() error = %v", err)
	}
	savedExperience, err := store.SaveExperience(ctx, experience)
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	project, err := (domain.ProjectInput{Name: "Release Console", RepositoryURL: "https://github.com/example/release-console", RepositoryID: 42, RepositoryReadme: "# Release Console", RepositoryVisibility: "public", RepositoryUpdatedAt: "2026-08-01T12:00:00Z", Skills: []string{"Go"}, DetectedLanguages: []domain.RepositoryLanguage{{Name: "Go", Bytes: 900}, {Name: "Shell", Bytes: 100}}, ResumeEligible: true}).Validate()
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
	certification, _ := (domain.CertificationInput{Name: "Cloud Professional", Issuer: "Example Institute"}).Validate()
	if _, err := store.SaveCertification(ctx, certification); err != nil {
		t.Fatalf("SaveCertification() error = %v", err)
	}
	achievement, _ := (domain.AchievementInput{Title: "Engineering Award", Description: "Recognized for a fictional audited release system."}).Validate()
	if _, err := store.SaveAchievement(ctx, achievement); err != nil {
		t.Fatalf("SaveAchievement() error = %v", err)
	}
	job, err := (domain.JobInput{Company: "Example Systems", Role: "Platform Engineer", Description: "Build reliable Go services and audited deployment pipelines for a growing platform."}).Validate()
	if err != nil {
		t.Fatalf("job Validate() error = %v", err)
	}
	savedJob, err := store.SaveJob(ctx, job)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	if _, err := store.CreateResumeVersion(ctx, savedJob.ID, []string{savedExperience.Bullets[0].ID}, "builtin-jake-style", savedJob.Description, `\documentclass{article}\begin{document}Snapshot\end{document}`); err != nil {
		t.Fatalf("CreateResumeVersion() error = %v", err)
	}
	template, err := (domain.ResumeTemplateInput{Name: "Local ATS", Source: `\documentclass{article}\begin{document}{{TAILORCV_NAME}}\end{document}`}).Validate()
	if err != nil {
		t.Fatalf("template Validate() error = %v", err)
	}
	savedTemplate, err := store.SaveTemplate(ctx, template)
	if err != nil {
		t.Fatalf("SaveTemplate() error = %v", err)
	}
	if err := store.SetSelectedTemplateID(ctx, savedTemplate.ID); err != nil {
		t.Fatalf("SetSelectedTemplateID() error = %v", err)
	}
	if _, err := store.SaveAIRun(ctx, domain.AIRun{JobID: savedJob.ID, Provider: "ollama", Model: "recorded", PromptVersion: "prompt-v1", SchemaVersion: "schema-v1", SelectedFactIDs: []string{savedExperience.Bullets[0].ID}, ValidationPassed: false, FailureCategory: "provider", ValidationErrors: []string{"provider unavailable"}, Proposals: []domain.AIProposal{}}); err != nil {
		t.Fatalf("SaveAIRun() error = %v", err)
	}
	if _, err := store.SaveAISettings(ctx, domain.AISettings{Provider: "gemini", OllamaEndpoint: domain.DefaultOllamaEndpoint, GeminiModel: "gemini-test"}); err != nil {
		t.Fatalf("SaveAISettings() error = %v", err)
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
	if got.Profile.Name != "Ada Lovelace" || len(got.Profile.ContactLinks) != 1 || len(got.Experiences) != 1 || len(got.Projects) != 1 || len(got.Educations) != 1 || len(got.Certifications) != 1 || len(got.Achievements) != 1 || len(got.Jobs) != 1 || len(got.Applications) != 1 || len(got.Applications[0].Versions) != 1 || len(got.Templates) != 1 || got.SelectedTemplateID != savedTemplate.ID || len(got.AIRuns) != 1 {
		t.Fatalf("restored backup = %#v", got)
	}
	if got.Experiences[0].Bullets[0].ID != backup.Experiences[0].Bullets[0].ID || got.Experiences[0].Bullets[0].Importance != domain.EvidenceImportanceEssential || got.Projects[0].RepositoryID != 42 || got.Projects[0].RepositoryReadme != "# Release Console" || got.Projects[0].Skills[0] != "Go" || len(got.Projects[0].DetectedLanguages) != 2 || got.Projects[0].DetectedLanguages[1].Name != "Shell" {
		t.Fatalf("restored child data = %#v", got)
	}
	if got.AISettings.Provider != "gemini" || got.AISettings.GeminiModel != "gemini-test" {
		t.Fatalf("restored AI settings = %#v", got.AISettings)
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

func TestProfileBackupImportRollsBackAfterWriteInterruption(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	profile, _ := (domain.ProfileInput{Name: "Keep Me", Skills: []string{"Go"}}).Validate()
	if _, err := store.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	backup, err := store.CreateProfileBackup(ctx)
	if err != nil {
		t.Fatalf("CreateProfileBackup() error = %v", err)
	}
	backup.Profile.Name = "Interrupted Replacement"
	if _, err := store.db.Exec(`CREATE TRIGGER interrupt_backup_import BEFORE INSERT ON profiles BEGIN SELECT RAISE(ABORT, 'simulated interruption'); END`); err != nil {
		t.Fatalf("create interruption trigger: %v", err)
	}
	if err := store.ReplaceProfileFromBackup(ctx, backup); err == nil || !strings.Contains(err.Error(), "simulated interruption") {
		t.Fatalf("ReplaceProfileFromBackup() interruption error = %v", err)
	}
	got, err := store.GetProfile(ctx)
	if err != nil || got.Name != "Keep Me" || len(got.Skills) != 1 || got.Skills[0] != "Go" {
		t.Fatalf("profile after interrupted import = %#v, %v", got, err)
	}
}
