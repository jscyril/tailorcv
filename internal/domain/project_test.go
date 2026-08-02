package domain

import "testing"

func TestProjectInputValidateNormalizesMetadata(t *testing.T) {
	project, err := (ProjectInput{
		Name:           "  Release   Console ",
		Role:           " Maintainer ",
		RepositoryURL:  "https://github.com/example/release-console",
		Ongoing:        true,
		EndDate:        "2025-01",
		ResumeEligible: true,
		Skills:         []string{" Go ", "go", "SQLite"},
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if project.Name != "Release Console" || project.EndDate != "" || len(project.Skills) != 2 {
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
}
