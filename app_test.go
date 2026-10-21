package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/storage"
)

func TestUpdateApplicationStatusAtAppBoundary(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	job, _ := (domain.JobInput{Role: "Platform Engineer", Description: "Build reliable distributed systems and audited release pipelines for production workloads."}).Validate()
	job, err = store.SaveJob(ctx, job)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	created, err := store.CreateResumeVersion(ctx, job.ID, []string{uuid.NewString()}, "builtin-jake-style", job.Description, "immutable source")
	if err != nil {
		t.Fatalf("CreateResumeVersion() error = %v", err)
	}
	app := &App{ctx: ctx, store: store}
	updated, err := app.UpdateApplicationStatus(domain.UpdateApplicationStatusInput{ApplicationID: created.Application.ID, Status: domain.ApplicationStatusArchived})
	if err != nil {
		t.Fatalf("UpdateApplicationStatus() error = %v", err)
	}
	if updated.Status != domain.ApplicationStatusArchived || len(updated.Versions) != 1 {
		t.Fatalf("updated application = %#v", updated)
	}
}

func TestSelectResumeEvidenceFiltersFactsAndRejectsLockedProjects(t *testing.T) {
	experiences := []domain.Experience{{ID: "experience", Bullets: []domain.EvidenceBullet{{ID: "experience-fact-1", Text: "Selected"}, {ID: "experience-fact-2", Text: "Not selected"}}}}
	projects := []domain.Project{
		{ID: "eligible", Name: "Eligible", ResumeEligible: true, Bullets: []domain.EvidenceBullet{{ID: "project-fact", Text: "Selected project fact"}}},
		{ID: "locked", Name: "Locked", ResumeEligible: false, Bullets: []domain.EvidenceBullet{{ID: "locked-fact", Text: "Locked project fact"}}},
	}
	selectedExperiences, selectedProjects, err := selectResumeEvidence([]string{"experience-fact-1", "project-fact"}, experiences, projects)
	if err != nil {
		t.Fatalf("selectResumeEvidence() error = %v", err)
	}
	if len(selectedExperiences) != 1 || len(selectedExperiences[0].Bullets) != 1 || selectedExperiences[0].Bullets[0].ID != "experience-fact-1" {
		t.Fatalf("selected experiences = %#v", selectedExperiences)
	}
	if len(selectedProjects) != 1 || len(selectedProjects[0].Bullets) != 1 || selectedProjects[0].Bullets[0].ID != "project-fact" {
		t.Fatalf("selected projects = %#v", selectedProjects)
	}
	if _, _, err := selectResumeEvidence([]string{"locked-fact"}, experiences, projects); err == nil {
		t.Fatal("selectResumeEvidence() expected locked project error")
	}
	if _, _, err := selectResumeEvidence([]string{"missing"}, experiences, projects); err == nil {
		t.Fatal("selectResumeEvidence() expected unknown evidence error")
	}
}

func TestResumeLibraryAtAppBoundaryRemovesArtifacts(t *testing.T) {
	app := newWorkflowTestApp(t)
	app.artifactDirectory = t.TempDir()
	app.compiler = recordedCompiler{result: domain.CompileResult{Success: true, Engine: "recorded"}, pdf: []byte("%PDF-1.7\n%%EOF\n")}

	if _, err := app.CreateResume(domain.CreateResumeInput{Name: "General", TemplateID: "missing-template", LatexSource: "x"}); err == nil {
		t.Fatal("CreateResume() expected unknown template error")
	}
	created, err := app.CreateResume(domain.CreateResumeInput{Name: "General", TemplateID: "builtin-jake-style", LatexSource: "\\documentclass{article}"})
	if err != nil {
		t.Fatalf("CreateResume() error = %v", err)
	}
	if _, err := app.CompileResumeVersion(created.Version.ID); err != nil {
		t.Fatalf("CompileResumeVersion() error = %v", err)
	}
	artifact := filepath.Join(app.artifactDirectory, created.Version.ID+".pdf")
	if _, err := os.Stat(artifact); err != nil {
		t.Fatalf("artifact missing before delete: %v", err)
	}
	copied, err := app.DuplicateResume(created.Application.ID)
	if err != nil || copied.Name != "Copy of General" {
		t.Fatalf("DuplicateResume() = %#v, %v", copied, err)
	}
	if renamed, err := app.RenameResume(domain.RenameResumeInput{ApplicationID: copied.ID, Name: "Renamed"}); err != nil || renamed.Name != "Renamed" {
		t.Fatalf("RenameResume() = %#v, %v", renamed, err)
	}
	if err := app.DeleteResume(created.Application.ID); err != nil {
		t.Fatalf("DeleteResume() error = %v", err)
	}
	if _, err := os.Stat(artifact); !os.IsNotExist(err) {
		t.Fatalf("artifact should be removed, stat err = %v", err)
	}
	if applications, err := app.ListApplications(); err != nil || len(applications) != 1 || applications[0].ID != copied.ID {
		t.Fatalf("remaining resumes = %#v, %v", applications, err)
	}
}

func TestExportFilenameIsSafeForSaveDialogs(t *testing.T) {
	cases := map[string]string{
		"Jacob Cyril - Acme - Staff Engineer": "Jacob Cyril - Acme - Staff Engineer.pdf",
		"  ../a/b:c*?\x00 ":                   "a-b-c.pdf",
		"":                                    "resume.pdf",
		"///":                                 "resume.pdf",
		strings.Repeat("é", 150):              strings.Repeat("é", 100) + ".pdf",
	}
	for input, want := range cases {
		if got := exportFilename(input, ".pdf"); got != want {
			t.Errorf("exportFilename(%q) = %q, want %q", input, got, want)
		}
	}
}
