package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/resume"
	"github.com/jscyril/tailorcv/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// verifyPackagedWorkflow exercises the compiled application service with a
// disposable database, real filesystem writes, and the packaged Tectonic
// runtime. Dialog selections are scripted because CI has no interactive
// desktop; interactive native-dialog checks remain a separate manual gate.
func verifyPackagedWorkflow() error {
	verificationRoot, err := os.MkdirTemp("", "tailorcv-packaged-verification-*")
	if err != nil {
		return fmt.Errorf("create verification workspace: %w", err)
	}
	defer os.RemoveAll(verificationRoot)

	store, err := storage.Open(filepath.Join(verificationRoot, "tailorcv.db"))
	if err != nil {
		return fmt.Errorf("open verification store: %w", err)
	}
	defer store.Close()

	ctx := context.Background()
	app := NewApp()
	app.ctx = ctx
	app.store = store
	app.artifactDirectory = filepath.Join(verificationRoot, "artifacts")
	app.github = packagedVerificationGitHubClient{}

	profile, err := app.SaveProfile(domain.ProfileInput{
		Name: "Fictional Candidate", Email: "fictional@example.com", Headline: "Platform Engineer",
		GitHubUsername: "fictional-candidate", Skills: []string{"Go", "Kubernetes"},
	})
	if err != nil {
		return fmt.Errorf("save verification profile: %w", err)
	}
	experience, err := app.SaveExperience(domain.ExperienceInput{
		Company: "Example Systems", Title: "Platform Engineer", StartDate: "2024-01", Current: true,
		Bullets: []domain.EvidenceBulletInput{{
			Text:         "Built reliable Go deployment services for fictional production workloads",
			Verification: domain.VerificationVerified, Importance: domain.EvidenceImportanceEssential,
		}},
	})
	if err != nil {
		return fmt.Errorf("save verification experience: %w", err)
	}
	if _, err := app.ImportGitHubProjects(); err != nil {
		return fmt.Errorf("import verification GitHub project: %w", err)
	}
	projects, err := app.ListProjects()
	if err != nil {
		return fmt.Errorf("list imported verification project: %w", err)
	}
	if len(projects) != 1 || projects[0].ResumeEligible {
		return fmt.Errorf("review imported verification project")
	}
	imported := projects[0]
	project, err := app.SaveProject(domain.ProjectInput{
		ID: imported.ID, Name: imported.Name, Role: "Maintainer", Description: imported.Description,
		RepositoryURL: imported.RepositoryURL, RepositoryID: imported.RepositoryID,
		RepositoryReadme: imported.RepositoryReadme, RepositoryVisibility: imported.RepositoryVisibility,
		RepositoryUpdatedAt: imported.RepositoryUpdatedAt, Provenance: domain.ProvenanceGitHub,
		Verification: domain.VerificationVerified, ResumeEligible: true, Skills: []string{"Kubernetes"},
		DetectedLanguages: imported.DetectedLanguages,
		Bullets: []domain.EvidenceBulletInput{{
			Text:         "Reduced fictional release toil with an audited Kubernetes workflow",
			Verification: domain.VerificationVerified,
		}},
	})
	if err != nil {
		return fmt.Errorf("approve verification GitHub project: %w", err)
	}

	analysis, err := app.AnalyzeJobDescription(domain.JobAnalysisInput{
		Company: "Fictional Cloud", Role: "Senior Platform Engineer",
		Description: "Build reliable Go services and Kubernetes deployment workflows for secure production systems.",
	})
	if err != nil {
		return fmt.Errorf("analyze verification job: %w", err)
	}
	created, err := app.CreateResumeVersion(domain.CreateResumeVersionInput{
		JobID:           analysis.Job.ID,
		SelectedFactIDs: []string{experience.Bullets[0].ID, project.Bullets[0].ID},
		TemplateID:      resume.DefaultTemplateID,
	})
	if err != nil {
		return fmt.Errorf("create verification resume: %w", err)
	}
	editedSource := created.Version.LatexSource + "\n% packaged verification edit\n"
	edited, err := app.SaveResumeVersionEdit(domain.SaveResumeVersionEditInput{
		ApplicationID: created.Application.ID, BaseVersionID: created.Version.ID, LatexSource: editedSource,
	})
	if err != nil {
		return fmt.Errorf("save verification resume edit: %w", err)
	}
	if edited.VersionNumber != 2 {
		return fmt.Errorf("save verification resume edit: got version %d", edited.VersionNumber)
	}
	compileResult, err := app.CompileResumeVersion(edited.ID)
	if err != nil {
		return fmt.Errorf("compile verification resume: %w", err)
	}
	if !compileResult.Success {
		return fmt.Errorf("compile verification resume: %s", firstVerificationDiagnostic(compileResult))
	}
	app.setLastPDF(nil)
	opened, err := app.OpenResumeVersion(edited.ID)
	if err != nil {
		return fmt.Errorf("reopen verification resume artifact: %w", err)
	}
	if !opened.Version.PDFAvailable || opened.CompileResult.PDFBase64 == "" {
		return fmt.Errorf("reopen verification resume artifact: compiled PDF was unavailable")
	}

	pdfPath := filepath.Join(verificationRoot, "exported-resume.pdf")
	backupPath := filepath.Join(verificationRoot, "profile-backup.json")
	app.saveFileDialog = func(_ context.Context, options runtime.SaveDialogOptions) (string, error) {
		switch options.DefaultFilename {
		case "resume.pdf":
			return pdfPath, nil
		case "tailorcv-profile-backup.json":
			return backupPath, nil
		default:
			return "", fmt.Errorf("unexpected verification save dialog %q", options.DefaultFilename)
		}
	}
	if exported, err := app.ExportCompiledPDF(); err != nil {
		return fmt.Errorf("export verification PDF: %w", err)
	} else if exported.Path != pdfPath {
		return fmt.Errorf("export verification PDF: unexpected destination")
	}
	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("read verification PDF: %w", err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-") {
		return fmt.Errorf("read verification PDF: invalid header")
	}
	if exported, err := app.ExportProfileBackup(); err != nil {
		return fmt.Errorf("export verification backup: %w", err)
	} else if exported.Path != backupPath {
		return fmt.Errorf("export verification backup: unexpected destination")
	}
	if _, err := app.SaveProfile(domain.ProfileInput{Name: "Temporary Replacement"}); err != nil {
		return fmt.Errorf("mutate verification profile: %w", err)
	}
	app.openFileDialog = func(_ context.Context, options runtime.OpenDialogOptions) (string, error) {
		if len(options.Filters) != 1 || options.Filters[0].Pattern != "*.json" {
			return "", fmt.Errorf("unexpected verification open dialog")
		}
		return backupPath, nil
	}
	if _, err := app.ImportProfileBackup(); err != nil {
		return fmt.Errorf("restore verification backup: %w", err)
	}
	restored, err := app.GetProfile()
	if err != nil {
		return fmt.Errorf("verify restored profile: %w", err)
	}
	if restored.Name != profile.Name || restored.GitHubUsername != profile.GitHubUsername {
		return fmt.Errorf("verify restored profile: restored values did not match")
	}
	return nil
}

func firstVerificationDiagnostic(result domain.CompileResult) string {
	if len(result.Diagnostics) > 0 {
		return result.Diagnostics[0].Message
	}
	if strings.TrimSpace(result.Log) != "" {
		return strings.TrimSpace(result.Log)
	}
	return "Tectonic returned an unsuccessful result"
}

type packagedVerificationGitHubClient struct{}

func (packagedVerificationGitHubClient) ListPublicRepositories(context.Context, string) ([]domain.GitHubRepository, error) {
	return []domain.GitHubRepository{{
		ID: 4242, Name: "release-console", Description: "Fictional release automation",
		HTMLURL: "https://github.com/fictional-candidate/release-console", Language: "Go",
		Languages:         []domain.RepositoryLanguage{{Name: "Go", Bytes: 900}, {Name: "Shell", Bytes: 100}},
		LanguagesComplete: true, Visibility: "public", UpdatedAt: "2026-08-01T12:00:00Z",
		Readme: "# Fictional Release Console", ReadmeComplete: true,
	}}, nil
}
