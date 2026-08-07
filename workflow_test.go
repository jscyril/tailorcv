package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/ai"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/resume"
	"github.com/jscyril/tailorcv/internal/storage"
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

	analysis, err := app.AnalyzeJobDescription(domain.JobAnalysisInput{
		Company: "Fictional Cloud", Role: "Senior Platform Engineer",
		Description: "Build reliable Go services and Kubernetes deployment workflows for secure production systems and improve release operations.",
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
