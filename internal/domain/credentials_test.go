package domain

import "testing"

func TestCertificationAndAchievementValidation(t *testing.T) {
	certification, err := (CertificationInput{Name: " Cloud Professional ", Issuer: " Example Institute ", IssueDate: "2025-01", ExpiryDate: "2027-01", CredentialURL: "https://example.com/credential", Verification: VerificationVerified}).Validate()
	if err != nil {
		t.Fatalf("certification Validate() error = %v", err)
	}
	if certification.Name != "Cloud Professional" || certification.Provenance != ProvenanceManual {
		t.Fatalf("certification = %#v", certification)
	}
	achievement, err := (AchievementInput{Title: "Engineering Award", Description: "Recognized for building a reliable fictional release system.", Date: "2025-06", SourceURL: "https://example.com/award"}).Validate()
	if err != nil {
		t.Fatalf("achievement Validate() error = %v", err)
	}
	if achievement.Verification != VerificationUnverified {
		t.Fatalf("achievement = %#v", achievement)
	}
	if _, err := (CertificationInput{Name: "Example", Issuer: "Issuer", IssueDate: "2026-01", ExpiryDate: "2025-01"}).Validate(); err == nil {
		t.Fatal("expected invalid certification dates")
	}
	if _, err := (AchievementInput{Title: "Example"}).Validate(); err == nil {
		t.Fatal("expected missing achievement description")
	}
}

func TestProfileValidatesAdditionalContactLinks(t *testing.T) {
	profile, err := (ProfileInput{ContactLinks: []ContactLinkInput{{Label: "Portfolio", URL: "https://example.com/work"}}}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(profile.ContactLinks) != 1 || profile.ContactLinks[0].Position != 0 {
		t.Fatalf("profile links = %#v", profile.ContactLinks)
	}
	if _, err := (ProfileInput{ContactLinks: []ContactLinkInput{{Label: "Unsafe", URL: "javascript:alert(1)"}}}).Validate(); err == nil {
		t.Fatal("expected unsafe contact URL error")
	}
}
