package storage

import (
	"context"
	"os"
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

func TestUpdateApplicationStatusPreservesResumeVersions(t *testing.T) {
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
	created, err := store.CreateResumeVersion(ctx, job.ID, []string{uuid.NewString()}, "builtin-jake-style", job.Description, "immutable source")
	if err != nil {
		t.Fatalf("CreateResumeVersion() error = %v", err)
	}
	updated, err := store.UpdateApplicationStatus(ctx, domain.UpdateApplicationStatusInput{ApplicationID: created.Application.ID, Status: domain.ApplicationStatusSubmitted})
	if err != nil {
		t.Fatalf("UpdateApplicationStatus() error = %v", err)
	}
	if updated.Status != domain.ApplicationStatusSubmitted || len(updated.Versions) != 1 || updated.Versions[0].LatexSource != "immutable source" || updated.UpdatedAt == created.Application.UpdatedAt {
		t.Fatalf("updated application = %#v", updated)
	}
	if _, err := store.UpdateApplicationStatus(ctx, domain.UpdateApplicationStatusInput{ApplicationID: uuid.NewString(), Status: domain.ApplicationStatusArchived}); err == nil {
		t.Fatal("UpdateApplicationStatus() expected missing application error")
	}
}

func TestEditedResumeVersionsAndCompilationMetadataPreserveSnapshots(t *testing.T) {
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
	factID := uuid.NewString()
	ranking := []domain.EvidenceMatch{{FactID: factID, Score: 88, Reasons: []string{"Matches Go"}}}
	created, err := store.CreateResumeVersion(ctx, job.ID, []string{factID}, "builtin-jake-style", job.Description, "original source", ranking)
	if err != nil {
		t.Fatalf("CreateResumeVersion() error = %v", err)
	}
	edited, err := store.CreateEditedResumeVersion(ctx, domain.SaveResumeVersionEditInput{ApplicationID: created.Application.ID, BaseVersionID: created.Version.ID, LatexSource: "edited source"})
	if err != nil {
		t.Fatalf("CreateEditedResumeVersion() error = %v", err)
	}
	if edited.VersionNumber != 2 || edited.ContentHash != domain.ResumeContentHash("edited source") || len(edited.RankingExplanations) != 1 {
		t.Fatalf("edited version = %#v", edited)
	}
	compile := domain.CompileResult{Success: true, Engine: "Tectonic", DurationMS: 37, Diagnostics: []domain.CompileDiagnostic{{Line: 4, Severity: "warning", Message: "example"}}}
	artifactPath := filepath.Join(t.TempDir(), edited.ID+".pdf")
	if err := os.WriteFile(artifactPath, []byte("pdf"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := store.RecordResumeCompilation(ctx, edited.ID, artifactPath, compile); err != nil {
		t.Fatalf("RecordResumeCompilation() error = %v", err)
	}
	applications, err := store.ListApplications(ctx)
	if err != nil {
		t.Fatalf("ListApplications() error = %v", err)
	}
	if len(applications) != 1 || len(applications[0].Versions) != 2 {
		t.Fatalf("applications = %#v", applications)
	}
	latest, original := applications[0].Versions[0], applications[0].Versions[1]
	if !latest.CompileSuccess || !latest.PDFAvailable || latest.CompileEngine != "Tectonic" || len(latest.CompileDiagnostics) != 1 {
		t.Fatalf("latest compilation = %#v", latest)
	}
	if original.LatexSource != "original source" || original.CompileSuccess || original.PDFAvailable {
		t.Fatalf("original snapshot was changed: %#v", original)
	}
}
