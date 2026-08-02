package domain

import "testing"

func TestGitHubRepositoryProjectRequiresReview(t *testing.T) {
	project, err := (GitHubRepository{Name: "release-console", Description: "Release automation", HTMLURL: "https://github.com/example/release-console", Homepage: "javascript:alert(1)", Language: "Go", Topics: []string{"sqlite", "Go"}}).Project(nil)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if project.Provenance != ProvenanceGitHub || project.Verification != VerificationUnverified || project.ResumeEligible {
		t.Fatalf("review state = %#v", project)
	}
	if project.URL != "" || len(project.Skills) != 2 {
		t.Fatalf("normalized metadata = %#v", project)
	}
}

func TestGitHubRepositoryProjectRefreshPreservesUserReviewAndEvidence(t *testing.T) {
	existing := Project{ID: "project-id", Name: "Old name", Role: "Creator", URL: "https://example.com/demo", StartDate: "2024-01", Verification: VerificationVerified, ResumeEligible: true, Skills: []string{"SQLite"}, Bullets: []EvidenceBullet{{ID: "bullet-id", Text: "Reduced release time", Provenance: ProvenanceManual, Verification: VerificationVerified}}}
	project, err := (GitHubRepository{Name: "new-name", Description: "Updated description", HTMLURL: "https://github.com/example/new-name", Language: "Go"}).Project(&existing)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if project.Name != "new-name" || project.Role != "Creator" || project.URL != existing.URL || project.Verification != VerificationVerified || !project.ResumeEligible || len(project.Skills) != 2 || len(project.Bullets) != 1 {
		t.Fatalf("refreshed project = %#v", project)
	}
}
