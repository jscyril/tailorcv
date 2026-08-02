package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestProjectCRUDPreservesSkillsEvidenceAndReviewState(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	validated, err := (domain.ProjectInput{
		Name:           "Release Console",
		Role:           "Maintainer",
		RepositoryURL:  "https://github.com/example/release-console",
		Provenance:     domain.ProvenanceManual,
		Verification:   domain.VerificationVerified,
		ResumeEligible: true,
		Skills:         []string{"Go", "SQLite"},
		Bullets: []domain.EvidenceBulletInput{
			{Text: "Reduced release time", Verification: domain.VerificationVerified},
			{Text: "Added auditable deployments", Verification: domain.VerificationUnverified},
		},
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	saved, err := store.SaveProject(context.Background(), validated)
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}
	if saved.ID == "" || saved.Bullets[0].ID == "" || saved.CreatedAt == "" {
		t.Fatalf("SaveProject() = %#v", saved)
	}

	firstID := saved.Bullets[0].ID
	saved.Bullets[0], saved.Bullets[1] = saved.Bullets[1], saved.Bullets[0]
	saved.ResumeEligible = false
	saved.Skills = []string{"Go", "React"}
	updated, err := store.SaveProject(context.Background(), saved)
	if err != nil {
		t.Fatalf("SaveProject(update) error = %v", err)
	}
	if updated.Bullets[1].ID != firstID || updated.ResumeEligible || updated.Skills[1] != "React" {
		t.Fatalf("updated project = %#v", updated)
	}

	got, err := store.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(got) != 1 || got[0].Bullets[1].ID != firstID || got[0].Skills[1] != "React" || got[0].ResumeEligible {
		t.Fatalf("ListProjects() = %#v", got)
	}

	if err := store.DeleteProject(context.Background(), saved.ID); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	var skills, bullets int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM project_skills`).Scan(&skills); err != nil {
		t.Fatalf("count project skills after delete: %v", err)
	}
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM project_bullets`).Scan(&bullets); err != nil {
		t.Fatalf("count project bullets after delete: %v", err)
	}
	if skills != 0 || bullets != 0 {
		t.Fatalf("project children after delete: skills=%d bullets=%d", skills, bullets)
	}
}
