package domain

import "testing"

func TestProjectInputValidateNormalizesMetadata(t *testing.T) {
	project, err := (ProjectInput{
		Name:                 "  Release   Console ",
		Role:                 " Maintainer ",
		RepositoryURL:        "https://github.com/example/release-console",
		RepositoryID:         42,
		RepositoryReadme:     "# Release Console",
		RepositoryVisibility: "PUBLIC",
		RepositoryUpdatedAt:  "2026-08-01T12:00:00Z",
		Ongoing:              true,
		EndDate:              "2025-01",
		ResumeEligible:       true,
		Skills:               []string{" Go ", "go", "SQLite"},
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if project.Name != "Release Console" || project.EndDate != "" || project.RepositoryID != 42 || project.RepositoryVisibility != "public" || project.RepositoryReadme != "# Release Console" || len(project.Skills) != 2 {
		t.Fatalf("Project = %#v", project)
	}
	if project.Provenance != ProvenanceManual || project.Verification != VerificationUnverified {
		t.Fatalf("project review metadata = %q, %q", project.Provenance, project.Verification)
	}
}

func TestProjectInputValidateRejectsInvalidMetadata(t *testing.T) {
	_, err := (ProjectInput{Name: "Example", URL: "file:///etc/passwd"}).Validate()
	if err == nil {
		t.Fatal("Validate() expected URL error")
	}

	_, err = (ProjectInput{Name: "Example", StartDate: "2025-01", EndDate: "2024-01"}).Validate()
	if err == nil {
		t.Fatal("Validate() expected date error")
	}
	if _, err = (ProjectInput{Name: "Example", RepositoryVisibility: "secret"}).Validate(); err == nil {
		t.Fatal("Validate() expected repository visibility error")
	}
}

func TestGitHubProjectMustBeReviewedBeforeResumeEligibility(t *testing.T) {
	if _, err := (ProjectInput{Name: "Imported", Provenance: ProvenanceGitHub, Verification: VerificationUnverified, ResumeEligible: true}).Validate(); err == nil {
		t.Fatal("Validate() accepted an unreviewed resume-eligible GitHub project")
	}
	if _, err := (ProjectInput{Name: "Imported", Provenance: ProvenanceGitHub, Verification: VerificationVerified, ResumeEligible: true}).Validate(); err != nil {
		t.Fatalf("Validate() rejected a reviewed GitHub project: %v", err)
	}
}
