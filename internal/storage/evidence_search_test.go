package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestEvidenceSearchIndexTracksExperienceAndProjectChanges(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()

	experience, _ := (domain.ExperienceInput{Company: "Example", Title: "Engineer", StartDate: "2024-01", Bullets: []domain.EvidenceBulletInput{{Text: "Automated reliable production deployments"}}}).Validate()
	experience, err = store.SaveExperience(ctx, experience)
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	hits, err := store.SearchEvidence(ctx, []string{"deployment"}, 10)
	if err != nil || len(hits) != 1 || hits[0].FactID != experience.Bullets[0].ID {
		t.Fatalf("deployment search = %#v, %v", hits, err)
	}
	experience.Bullets[0].Text = "Improved observability dashboards"
	if _, err := store.SaveExperience(ctx, experience); err != nil {
		t.Fatalf("SaveExperience(update) error = %v", err)
	}
	hits, err = store.SearchEvidence(ctx, []string{"deployment"}, 10)
	if err != nil || len(hits) != 0 {
		t.Fatalf("stale deployment search = %#v, %v", hits, err)
	}

	project, _ := (domain.ProjectInput{Name: "Release Console", Description: "Internal delivery platform", Skills: []string{"Kubernetes"}, ResumeEligible: true}).Validate()
	project, err = store.SaveProject(ctx, project)
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}
	hits, err = store.SearchEvidence(ctx, []string{"Kubernetes"}, 10)
	if err != nil || len(hits) != 1 || hits[0].FactID != project.ID {
		t.Fatalf("project skill search = %#v, %v", hits, err)
	}
	if err := store.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	hits, err = store.SearchEvidence(ctx, []string{"Kubernetes"}, 10)
	if err != nil || len(hits) != 0 {
		t.Fatalf("deleted project search = %#v, %v", hits, err)
	}
}
