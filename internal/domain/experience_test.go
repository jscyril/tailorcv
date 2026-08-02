package domain

import "testing"

func TestExperienceInputValidateNormalizesEvidence(t *testing.T) {
	experience, err := (ExperienceInput{
		Company:   "  Analytical   Engines Ltd ",
		Title:     " Programmer ",
		StartDate: "1842-01",
		Current:   true,
		EndDate:   "1843-01",
		Bullets: []EvidenceBulletInput{{
			Text:         "  Published   the first algorithm. ",
			Provenance:   ProvenanceManual,
			Verification: VerificationVerified,
		}},
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if experience.Company != "Analytical Engines Ltd" || experience.EndDate != "" {
		t.Fatalf("Experience = %#v", experience)
	}
	if got := experience.Bullets[0].Text; got != "Published the first algorithm." {
		t.Fatalf("Bullet text = %q", got)
	}
}

func TestExperienceInputValidateRejectsInvalidDatesAndURLs(t *testing.T) {
	_, err := (ExperienceInput{
		Company:   "Example",
		Title:     "Engineer",
		StartDate: "2025-02",
		EndDate:   "2024-12",
	}).Validate()
	if err == nil {
		t.Fatal("Validate() expected date error")
	}

	_, err = (ExperienceInput{
		Company:   "Example",
		Title:     "Engineer",
		StartDate: "2024-01",
		Bullets: []EvidenceBulletInput{{
			Text:      "Built a service",
			SourceURL: "javascript:alert(1)",
		}},
	}).Validate()
	if err == nil {
		t.Fatal("Validate() expected source URL error")
	}
}
