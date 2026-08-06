package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func TestAIRunRoundTripAndAcceptance(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "ai.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	job, err := store.SaveJob(ctx, domain.Job{Description: "A sufficiently detailed job description for testing persisted AI tailoring runs."})
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	factID := uuid.NewString()
	run, err := store.SaveAIRun(ctx, domain.AIRun{
		JobID: job.ID, Provider: "ollama", Model: "qwen3:8b", PromptVersion: "prompt-v1", SchemaVersion: "schema-v1",
		SelectedFactIDs: []string{factID}, ValidationPassed: true,
		Proposals: []domain.AIProposal{{TargetFactID: factID, SupportingFactIDs: []string{factID}, Text: "A supported rewritten evidence bullet for testing."}},
	})
	if err != nil {
		t.Fatalf("SaveAIRun() error = %v", err)
	}
	loaded, err := store.GetAIRun(ctx, run.ID)
	if err != nil || len(loaded.Proposals) != 1 || !loaded.ValidationPassed {
		t.Fatalf("GetAIRun() = %#v, %v", loaded, err)
	}
	if err := store.MarkAIRunAccepted(ctx, run.ID, uuid.NewString()); err == nil {
		t.Fatal("MarkAIRunAccepted() accepted an unknown resume version")
	}
}
