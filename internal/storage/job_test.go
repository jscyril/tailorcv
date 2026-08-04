package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestJobCRUDPreservesIdentityAndOrdersRecentFirst(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()

	first, err := (domain.JobInput{Company: "Example", Role: "Platform Engineer", Description: "Build reliable distributed services using Go, PostgreSQL, and Kubernetes in production."}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	first, err = store.SaveJob(ctx, first)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	second, _ := (domain.JobInput{Company: "Other", Role: "Backend Engineer", Description: "Design backend APIs and maintain observable services for a growing customer platform."}).Validate()
	second, err = store.SaveJob(ctx, second)
	if err != nil {
		t.Fatalf("SaveJob(second) error = %v", err)
	}
	first.Role = "Senior Platform Engineer"
	updated, err := store.SaveJob(ctx, first)
	if err != nil {
		t.Fatalf("SaveJob(update) error = %v", err)
	}
	if updated.ID != first.ID || updated.CreatedAt != first.CreatedAt || updated.Role != "Senior Platform Engineer" {
		t.Fatalf("updated job = %#v", updated)
	}

	jobs, err := store.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs() error = %v", err)
	}
	if len(jobs) != 2 || jobs[0].ID != first.ID || jobs[1].ID != second.ID {
		t.Fatalf("ListJobs() = %#v", jobs)
	}
	if err := store.DeleteJob(ctx, second.ID); err != nil {
		t.Fatalf("DeleteJob() error = %v", err)
	}
}
