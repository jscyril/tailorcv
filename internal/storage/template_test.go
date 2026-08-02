package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestTemplateRoundTripAndSelection(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	template, err := (domain.ResumeTemplateInput{Name: "Custom", Description: "Test", Source: `\documentclass{article}\begin{document}Hi\end{document}`}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	saved, err := store.SaveTemplate(ctx, template)
	if err != nil {
		t.Fatalf("SaveTemplate() error = %v", err)
	}
	if saved.ID == "" || saved.CreatedAt == "" || saved.UpdatedAt == "" {
		t.Fatalf("SaveTemplate() = %#v", saved)
	}
	if err := store.SetSelectedTemplateID(ctx, saved.ID); err != nil {
		t.Fatalf("SetSelectedTemplateID() error = %v", err)
	}
	selected, err := store.SelectedTemplateID(ctx)
	if err != nil || selected != saved.ID {
		t.Fatalf("SelectedTemplateID() = %q, %v", selected, err)
	}
	templates, err := store.ListTemplates(ctx)
	if err != nil || len(templates) != 1 || templates[0].Name != "Custom" {
		t.Fatalf("ListTemplates() = %#v, %v", templates, err)
	}
	if err := store.DeleteTemplate(ctx, saved.ID); err != nil {
		t.Fatalf("DeleteTemplate() error = %v", err)
	}
}
