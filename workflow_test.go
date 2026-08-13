package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/ai"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/resume"
	"github.com/jscyril/tailorcv/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func newWorkflowTestApp(t *testing.T) *App {
	t.Helper()
	store, err := storage.Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return &App{ctx: context.Background(), store: store}
}

func TestNoAIWorkflowAtAppBoundary(t *testing.T) {
	app := newWorkflowTestApp(t)
	profile, err := app.SaveProfile(domain.ProfileInput{
		Name: "Ada Lovelace", Email: "ada@example.com", Headline: "Platform Engineer",
		Skills: []string{"Go", "Kubernetes"}, ContactLinks: []domain.ContactLinkInput{{Label: "Portfolio", URL: "https://example.com/work"}},
	})
	if err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	if profile.Name != "Ada Lovelace" || len(profile.ContactLinks) != 1 {
		t.Fatalf("saved profile = %#v", profile)
	}

	experience, err := app.SaveExperience(domain.ExperienceInput{
		Company: "Example Systems", Title: "Platform Engineer", StartDate: "2024-01", Current: true,
		Bullets: []domain.EvidenceBulletInput{{
			Text:         "Built reliable Go deployment services for production workloads",
			Verification: domain.VerificationVerified, Importance: domain.EvidenceImportanceEssential,
		}},
	})
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	project, err := app.SaveProject(domain.ProjectInput{
		Name: "Release Console", Role: "Maintainer", StartDate: "2025-01", Ongoing: true,
		Skills: []string{"Kubernetes"}, Verification: domain.VerificationVerified, ResumeEligible: true,
		Bullets: []domain.EvidenceBulletInput{{Text: "Reduced release toil with an audited Kubernetes workflow", Verification: domain.VerificationVerified}},
	})
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}
	if _, err := app.SaveCertification(domain.CertificationInput{Name: "Cloud Professional", Issuer: "Example Institute"}); err != nil {
		t.Fatalf("SaveCertification() error = %v", err)
	}
	if _, err := app.SaveAchievement(domain.AchievementInput{Title: "Engineering Award", Description: "Recognized for improving a fictional audited release system."}); err != nil {
		t.Fatalf("SaveAchievement() error = %v", err)
	}

	jobInjection := `\immediate\write18{touch /tmp/tailorcv-job-injection}`
	analysis, err := app.AnalyzeJobDescription(domain.JobAnalysisInput{
		Company: "Fictional Cloud", Role: "Senior Platform Engineer",
		Description: "Build reliable Go services and Kubernetes deployment workflows for secure production systems and improve release operations. " + jobInjection,
	})
	if err != nil {
		t.Fatalf("AnalyzeJobDescription() error = %v", err)
	}
	if analysis.Job.ID == "" || len(analysis.RankedEvidence) < 2 || analysis.RankedEvidence[0].FactID != experience.Bullets[0].ID {
		t.Fatalf("analysis = %#v", analysis)
	}
	if !containsReason(analysis.RankedEvidence[0].Reasons, "Marked essential by you") || !containsReason(analysis.RankedEvidence[0].Reasons, "Role is current") {
		t.Fatalf("ranking reasons = %#v", analysis.RankedEvidence[0].Reasons)
	}

	selected := []string{experience.Bullets[0].ID, project.Bullets[0].ID}
	created, err := app.CreateResumeVersion(domain.CreateResumeVersionInput{JobID: analysis.Job.ID, SelectedFactIDs: selected, TemplateID: resume.DefaultTemplateID})
	if err != nil {
		t.Fatalf("CreateResumeVersion() error = %v", err)
	}
	for _, expected := range []string{"Built reliable Go deployment services", "Reduced release toil", "Cloud Professional", "Engineering Award"} {
		if !strings.Contains(created.Version.LatexSource, expected) {
			t.Fatalf("resume source does not contain %q", expected)
		}
	}
	if created.Version.VersionNumber != 1 || len(created.Version.RankingExplanations) != 2 {
		t.Fatalf("created version = %#v", created.Version)
	}
	if !strings.Contains(created.Version.JobDescriptionSnapshot, jobInjection) || strings.Contains(created.Version.LatexSource, jobInjection) {
		t.Fatalf("untrusted job text crossed into executable resume source")
	}

	editedSource := strings.Replace(created.Version.LatexSource, "Senior Platform Engineer", "Staff Platform Engineer", 1)
	if editedSource == created.Version.LatexSource {
		editedSource += "\n% reviewed edit"
	}
	edited, err := app.SaveResumeVersionEdit(domain.SaveResumeVersionEditInput{ApplicationID: created.Application.ID, BaseVersionID: created.Version.ID, LatexSource: editedSource})
	if err != nil {
		t.Fatalf("SaveResumeVersionEdit() error = %v", err)
	}
	if edited.VersionNumber != 2 || edited.ContentHash == created.Version.ContentHash || edited.JobDescriptionSnapshot != created.Version.JobDescriptionSnapshot {
		t.Fatalf("edited version = %#v", edited)
	}
	applications, err := app.ListApplications()
	if err != nil || len(applications) != 1 || len(applications[0].Versions) != 2 {
		t.Fatalf("ListApplications() = %#v, %v", applications, err)
	}
	for _, version := range applications[0].Versions {
		if version.ID == created.Version.ID && version.LatexSource != created.Version.LatexSource {
			t.Fatal("editing a resume mutated the original immutable version")
		}
	}

	// Compile the edited immutable source, clear process memory to simulate a
	// restart, reopen the persisted artifact, and export it through the native
	// dialog boundary without compiling it again.
	pdf := []byte("%PDF-1.7\nfictional TailorCV workflow fixture\n%%EOF\n")
	app.compiler = recordedCompiler{result: domain.CompileResult{
		Success: true, Engine: "Recorded Tectonic", DurationMS: 12,
		PDFBase64: base64.StdEncoding.EncodeToString(pdf),
	}, pdf: pdf}
	app.artifactDirectory = filepath.Join(t.TempDir(), "artifacts")
	compileResult, err := app.CompileResumeVersion(edited.ID)
	if err != nil || !compileResult.Success {
		t.Fatalf("CompileResumeVersion() = %#v, %v", compileResult, err)
	}
	app.setLastPDF(nil)
	opened, err := app.OpenResumeVersion(edited.ID)
	if err != nil {
		t.Fatalf("OpenResumeVersion() error = %v", err)
	}
	if opened.Version.ID != edited.ID || !opened.Version.PDFAvailable || opened.CompileResult.PDFBase64 == "" {
		t.Fatalf("opened workspace = %#v", opened)
	}
	exportWithoutExtension := filepath.Join(t.TempDir(), "staff-platform-resume")
	exportPath := exportWithoutExtension + ".pdf"
	if err := os.WriteFile(exportPath, []byte("old PDF"), 0o600); err != nil {
		t.Fatalf("seed export destination: %v", err)
	}
	app.saveFileDialog = func(_ context.Context, options runtime.SaveDialogOptions) (string, error) {
		if options.DefaultFilename != "resume.pdf" {
			t.Errorf("export dialog options = %#v", options)
		}
		return exportWithoutExtension, nil
	}
	exported, err := app.ExportCompiledPDF()
	if err != nil {
		t.Fatalf("ExportCompiledPDF() error = %v", err)
	}
	exportedPDF, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("ReadFile(export) error = %v", err)
	}
	if exported.Path != exportPath || string(exportedPDF) != string(pdf) {
		t.Fatalf("exported = %#v, bytes = %q", exported, exportedPDF)
	}
	app.compiler = recordedCompiler{err: fmt.Errorf("Tectonic is unavailable")}
	if _, err := app.CompileResumeVersion(edited.ID); err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("CompileResumeVersion(unavailable compiler) error = %v", err)
	}
	if _, err := app.ExportCompiledPDF(); err == nil || !strings.Contains(err.Error(), "compile the current LaTeX source") {
		t.Fatalf("ExportCompiledPDF(after unavailable compiler) error = %v", err)
	}
	opened, err = app.OpenResumeVersion(edited.ID)
	if err != nil || !opened.Version.PDFAvailable || opened.CompileResult.PDFBase64 == "" {
		t.Fatalf("OpenResumeVersion(after unavailable compiler) = %#v, %v", opened, err)
	}
	artifactPath, err := app.resumeArtifactPath(edited.ID)
	if err != nil {
		t.Fatalf("resumeArtifactPath() error = %v", err)
	}
	if err := os.WriteFile(artifactPath, []byte("not a PDF"), 0o600); err != nil {
		t.Fatalf("corrupt artifact fixture: %v", err)
	}
	app.setLastPDF(pdf)
	if _, err := app.OpenResumeVersion(edited.ID); err == nil || !strings.Contains(err.Error(), "not a valid bounded PDF") {
		t.Fatalf("OpenResumeVersion(corrupt artifact) error = %v", err)
	}
	if _, err := app.ExportCompiledPDF(); err == nil || !strings.Contains(err.Error(), "compile the current LaTeX source") {
		t.Fatalf("ExportCompiledPDF(after corrupt artifact) error = %v", err)
	}
	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("remove artifact fixture: %v", err)
	}
	opened, err = app.OpenResumeVersion(edited.ID)
	if err != nil || opened.Version.PDFAvailable || opened.CompileResult.PDFBase64 != "" {
		t.Fatalf("OpenResumeVersion(missing artifact) = %#v, %v", opened, err)
	}
}

func TestBackupRestoreWorkflowAtAppBoundary(t *testing.T) {
	app := newWorkflowTestApp(t)
	if _, err := app.SaveProfile(domain.ProfileInput{Name: "Ada Lovelace", Email: "ada@example.com", Skills: []string{"Go"}}); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	if _, err := app.SaveExperience(domain.ExperienceInput{
		Company: "Example Systems", Title: "Engineer", StartDate: "2024-01",
		Bullets: []domain.EvidenceBulletInput{{Text: "Built a fictional audited release service", Verification: domain.VerificationVerified}},
	}); err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	backupWithoutExtension := filepath.Join(t.TempDir(), "profile-backup")
	backupPath := backupWithoutExtension + ".json"
	app.saveFileDialog = func(_ context.Context, options runtime.SaveDialogOptions) (string, error) {
		if options.DefaultFilename != "tailorcv-profile-backup.json" {
			t.Errorf("backup dialog options = %#v", options)
		}
		return backupWithoutExtension, nil
	}
	exported, err := app.ExportProfileBackup()
	if err != nil {
		t.Fatalf("ExportProfileBackup() error = %v", err)
	}
	if exported.Path != backupPath || exported.ExperienceCount != 1 {
		t.Fatalf("exported backup = %#v", exported)
	}
	if _, err := app.SaveProfile(domain.ProfileInput{Name: "Temporary Replacement", Skills: []string{"Rust"}}); err != nil {
		t.Fatalf("replace profile before restore: %v", err)
	}
	app.openFileDialog = func(_ context.Context, options runtime.OpenDialogOptions) (string, error) {
		if len(options.Filters) != 1 || options.Filters[0].Pattern != "*.json" {
			t.Errorf("restore dialog options = %#v", options)
		}
		return backupPath, nil
	}
	restored, err := app.ImportProfileBackup()
	if err != nil {
		t.Fatalf("ImportProfileBackup() error = %v", err)
	}
	profile, err := app.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if restored.Path != backupPath || profile.Name != "Ada Lovelace" || len(profile.Skills) != 1 || profile.Skills[0] != "Go" {
		t.Fatalf("restored = %#v, profile = %#v", restored, profile)
	}
}

func TestGitHubImportReviewWorkflowAtAppBoundary(t *testing.T) {
	app := newWorkflowTestApp(t)
	if _, err := app.SaveProfile(domain.ProfileInput{Name: "Ada Lovelace", GitHubUsername: "ada-example"}); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	repositories := &recordedGitHubClient{repositories: []domain.GitHubRepository{{
		ID: 42, Name: "release-console", Description: "Fictional release automation",
		HTMLURL: "https://github.com/ada-example/release-console", Language: "Go",
		Languages: []domain.RepositoryLanguage{{Name: "Go", Bytes: 900}, {Name: "Shell", Bytes: 100}}, LanguagesComplete: true,
		Visibility: "public", UpdatedAt: "2026-08-01T12:00:00Z", Readme: "# Release Console", ReadmeComplete: true,
	}}}
	app.github = repositories
	result, err := app.ImportGitHubProjects()
	if err != nil {
		t.Fatalf("ImportGitHubProjects() error = %v", err)
	}
	projects, err := app.ListProjects()
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects() = %#v, %v", projects, err)
	}
	imported := projects[0]
	if result.Imported != 1 || imported.RepositoryID != 42 || imported.RepositoryReadme != "# Release Console" || imported.Verification != domain.VerificationUnverified || imported.ResumeEligible {
		t.Fatalf("import result = %#v, project = %#v", result, imported)
	}
	approved, err := app.SaveProject(domain.ProjectInput{
		ID: imported.ID, Name: imported.Name, Role: "Creator", Description: imported.Description,
		RepositoryURL: imported.RepositoryURL, RepositoryID: imported.RepositoryID, RepositoryReadme: imported.RepositoryReadme,
		RepositoryVisibility: imported.RepositoryVisibility, RepositoryUpdatedAt: imported.RepositoryUpdatedAt,
		Provenance: domain.ProvenanceGitHub, Verification: domain.VerificationVerified, ResumeEligible: true,
		Skills: []string{"Go"}, DetectedLanguages: imported.DetectedLanguages,
		Bullets: []domain.EvidenceBulletInput{{Text: "Built an audited fictional release workflow", Verification: domain.VerificationVerified}},
	})
	if err != nil {
		t.Fatalf("SaveProject(approve) error = %v", err)
	}
	if !approved.ResumeEligible || approved.Verification != domain.VerificationVerified {
		t.Fatalf("approved project = %#v", approved)
	}
	repositories.repositories[0].Name = "release-console-next"
	repositories.repositories[0].Description = "Updated upstream description"
	repositories.repositories[0].Readme = ""
	repositories.repositories[0].ReadmeComplete = false
	result, err = app.ImportGitHubProjects()
	if err != nil {
		t.Fatalf("ImportGitHubProjects(refresh) error = %v", err)
	}
	projects, err = app.ListProjects()
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects(after refresh) = %#v, %v", projects, err)
	}
	refreshed := projects[0]
	if result.Updated != 1 || result.ReadmeFallbacks != 1 || refreshed.Name != "release-console-next" || refreshed.RepositoryReadme != "# Release Console" || !refreshed.ResumeEligible || refreshed.Verification != domain.VerificationVerified || len(refreshed.Bullets) != 1 || refreshed.Bullets[0].Text != approved.Bullets[0].Text {
		t.Fatalf("refresh result = %#v, project = %#v", result, refreshed)
	}
}

func TestRecordedOllamaWorkflowAtAppBoundary(t *testing.T) {
	app := newWorkflowTestApp(t)
	if _, err := app.SaveProfile(domain.ProfileInput{Name: "Ada Lovelace", Skills: []string{"Go"}}); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	experience, err := app.SaveExperience(domain.ExperienceInput{
		Company: "Example", Title: "Engineer", StartDate: "2024-01",
		Bullets: []domain.EvidenceBulletInput{{Text: "Reduced deployment time by 40% with an audited Go release pipeline", Verification: domain.VerificationVerified}},
	})
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	analysis, err := app.AnalyzeJobDescription(domain.JobAnalysisInput{Role: "Platform Engineer", Description: "Build reliable Go deployment systems and improve release operations for production services."})
	if err != nil {
		t.Fatalf("AnalyzeJobDescription() error = %v", err)
	}
	factID := experience.Bullets[0].ID
	app.aiProviderFactory = func(name, endpoint string) (ai.Provider, error) {
		if name != "ollama" || endpoint != domain.DefaultOllamaEndpoint {
			t.Errorf("provider request = %q, %q", name, endpoint)
		}
		response := fmt.Sprintf(`{"schemaVersion":"%s","proposals":[{"targetFactId":%q,"supportingFactIds":[%q],"text":"Reduced deployment time by 40%% through an audited Go release pipeline."}]}`, ai.SchemaVersion, factID, factID)
		return recordedAIProvider{response: []byte(response)}, nil
	}
	run, err := app.GenerateAITailoring(domain.GenerateAITailoringInput{
		JobID: analysis.Job.ID, SelectedFactIDs: []string{factID}, Provider: "ollama", Model: "recorded", Endpoint: domain.DefaultOllamaEndpoint,
	})
	if err != nil {
		t.Fatalf("GenerateAITailoring() error = %v", err)
	}
	if !run.ValidationPassed || len(run.Proposals) != 1 {
		t.Fatalf("AI run = %#v", run)
	}
	accepted, err := app.AcceptAITailoring(domain.AcceptAITailoringInput{RunID: run.ID, TemplateID: resume.DefaultTemplateID, Proposals: run.Proposals})
	if err != nil {
		t.Fatalf("AcceptAITailoring() error = %v", err)
	}
	if accepted.Version.VersionNumber != 1 || !strings.Contains(accepted.Version.LatexSource, "Reduced deployment time by 40\\% through") {
		t.Fatalf("accepted version = %#v", accepted.Version)
	}
	storedRuns, err := app.ListAIRuns()
	if err != nil || len(storedRuns) != 1 || storedRuns[0].ResumeVersionID != accepted.Version.ID || storedRuns[0].AcceptedAt == "" {
		t.Fatalf("ListAIRuns() = %#v, %v", storedRuns, err)
	}
	storedEvidence, err := app.ListExperiences()
	if err != nil || storedEvidence[0].Bullets[0].Text != experience.Bullets[0].Text {
		t.Fatalf("source evidence after AI acceptance = %#v, %v", storedEvidence, err)
	}
}

func containsReason(reasons []string, expected string) bool {
	for _, reason := range reasons {
		if reason == expected {
			return true
		}
	}
	return false
}

type recordedCompiler struct {
	result domain.CompileResult
	pdf    []byte
	err    error
}

func (compiler recordedCompiler) Compile(_ context.Context, _ string) (domain.CompileResult, []byte, error) {
	return compiler.result, append([]byte(nil), compiler.pdf...), compiler.err
}

type recordedGitHubClient struct {
	repositories []domain.GitHubRepository
	err          error
}

func (client *recordedGitHubClient) ListPublicRepositories(_ context.Context, _ string) ([]domain.GitHubRepository, error) {
	return append([]domain.GitHubRepository(nil), client.repositories...), client.err
}
