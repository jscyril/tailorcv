package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestProfileRoundTrip(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	input := domain.ProfileInput{
		Name:           "Ada Lovelace",
		Headline:       "Computing pioneer",
		Email:          "ada@example.com",
		GitHubUsername: "ada-lovelace",
		Skills:         []string{"Mathematics", "Go"},
	}
	profile, err := input.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if _, err := store.SaveProfile(context.Background(), profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	got, err := store.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if got.Name != input.Name || len(got.Skills) != 2 || got.UpdatedAt == "" {
		t.Fatalf("GetProfile() = %#v", got)
	}
}
