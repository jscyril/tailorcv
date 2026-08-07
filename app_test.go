package main

import (
	"context"
	"path/filepath"
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
