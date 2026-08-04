package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func TestCreateResumeVersionAppendsImmutableSnapshots(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	job, _ := (domain.JobInput{Role: "Platform Engineer", Description: "Build reliable distributed services with Go and PostgreSQL for production workloads."}).Validate()
	job, err = store.SaveJob(ctx, job)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	firstFact, secondFact := uuid.NewString(), uuid.NewString()
	first, err := store.CreateResumeVersion(ctx, job.ID, []string{firstFact}, "builtin-jake-style", job.Description, "first latex snapshot")
	if err != nil {
		t.Fatalf("CreateResumeVersion(first) error = %v", err)
	}
	second, err := store.CreateResumeVersion(ctx, job.ID, []string{secondFact}, "builtin-jake-style", job.Description, "second latex snapshot")
	if err != nil {
		t.Fatalf("CreateResumeVersion(second) error = %v", err)
	}
	if first.Application.ID != second.Application.ID || first.Version.VersionNumber != 1 || second.Version.VersionNumber != 2 || first.Version.ID == second.Version.ID {
		t.Fatalf("version results: first=%#v second=%#v", first, second)
	}
	applications, err := store.ListApplications(ctx)
	if err != nil {
		t.Fatalf("ListApplications() error = %v", err)
	}
	if len(applications) != 1 || len(applications[0].Versions) != 2 || applications[0].Versions[1].LatexSource != "first latex snapshot" || applications[0].SelectedFactIDs[0] != secondFact {
		t.Fatalf("applications = %#v", applications)
	}
}
