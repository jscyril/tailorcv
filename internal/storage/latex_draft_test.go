package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestLatexDraftSaveReadClearAndDamagedValue(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()

	if draft, err := store.LatexDraft(ctx); err != nil || draft.Source != "" {
		t.Fatalf("empty LatexDraft() = %#v, %v", draft, err)
	}
	want := domain.LatexDraft{Source: "\\section{Edited}", VersionID: "v1", SavedAt: "2026-10-21T10:00:00Z"}
	if err := store.SaveLatexDraft(ctx, want); err != nil {
		t.Fatalf("SaveLatexDraft() error = %v", err)
	}
	if got, err := store.LatexDraft(ctx); err != nil || got != want {
		t.Fatalf("LatexDraft() = %#v, %v", got, err)
	}
	if err := store.ClearLatexDraft(ctx); err != nil {
		t.Fatalf("ClearLatexDraft() error = %v", err)
	}
	if got, _ := store.LatexDraft(ctx); got.Source != "" {
		t.Fatalf("draft after clear = %#v", got)
	}

	if _, err := store.db.ExecContext(ctx, `INSERT INTO app_settings(key, value) VALUES (?, '{broken')`, latexDraftSetting); err != nil {
		t.Fatalf("seed damaged draft: %v", err)
	}
	if got, err := store.LatexDraft(ctx); err != nil || got.Source != "" {
		t.Fatalf("damaged LatexDraft() = %#v, %v", got, err)
	}
}
