package storage

import (
	"context"
	"github.com/jscyril/tailorcv/internal/domain"
	"path/filepath"
	"testing"
)

func TestCredentialsAndContactLinksRoundTrip(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	profile, _ := (domain.ProfileInput{Name: "Ada Example", ContactLinks: []domain.ContactLinkInput{{Label: "Portfolio", URL: "https://example.com/work"}}}).Validate()
	savedProfile, err := store.SaveProfile(ctx, profile)
	if err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	if savedProfile.ContactLinks[0].ID == "" {
		t.Fatal("contact link did not receive an ID")
	}
	cert, _ := (domain.CertificationInput{Name: "Cloud Professional", Issuer: "Example Institute", Verification: domain.VerificationVerified}).Validate()
	savedCert, err := store.SaveCertification(ctx, cert)
	if err != nil {
		t.Fatalf("SaveCertification() error = %v", err)
	}
	achievement, _ := (domain.AchievementInput{Title: "Engineering Award", Description: "Recognized for a fictional audited release system."}).Validate()
	savedAchievement, err := store.SaveAchievement(ctx, achievement)
	if err != nil {
		t.Fatalf("SaveAchievement() error = %v", err)
	}
	profileAgain, err := store.GetProfile(ctx)
	if err != nil || profileAgain.ContactLinks[0].ID != savedProfile.ContactLinks[0].ID {
		t.Fatalf("GetProfile() = %#v, %v", profileAgain, err)
	}
	certifications, _ := store.ListCertifications(ctx)
	achievements, _ := store.ListAchievements(ctx)
	if len(certifications) != 1 || certifications[0].ID != savedCert.ID || len(achievements) != 1 || achievements[0].ID != savedAchievement.ID {
		t.Fatalf("credentials = %#v %#v", certifications, achievements)
	}
}
