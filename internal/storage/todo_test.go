package storage

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

func TestTodosLifecycleOrderingAndBackup(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "tailorcv.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()

	resume, err := store.CreateResume(ctx, domain.CreateResumeInput{Name: "Backend", TemplateID: "builtin-jake-style", LatexSource: "source"})
	if err != nil {
		t.Fatalf("CreateResume() error = %v", err)
	}
	save := func(input domain.TodoInput) domain.Todo {
		t.Helper()
		item, err := input.Validate()
		if err != nil {
			t.Fatalf("Validate(%#v) error = %v", input, err)
		}
		saved, err := store.SaveTodo(ctx, item)
		if err != nil {
			t.Fatalf("SaveTodo() error = %v", err)
		}
		return saved
	}
	undated := save(domain.TodoInput{Title: "Update LinkedIn"})
	later := save(domain.TodoInput{Title: "Follow up", DueDate: "2026-11-20", ApplicationID: resume.Application.ID})
	sooner := save(domain.TodoInput{Title: "  Apply   to Acme ", Notes: "Use the backend resume", DueDate: "2026-11-01"})
	if sooner.Title != "Apply to Acme" || later.ApplicationID != resume.Application.ID || sooner.CompletedAt != "" {
		t.Fatalf("saved to-dos = %#v %#v", sooner, later)
	}

	done, err := store.SetTodoDone(ctx, sooner.ID, true)
	if err != nil || done.CompletedAt == "" {
		t.Fatalf("SetTodoDone(true) = %#v, %v", done, err)
	}
	// Editing keeps completion; only SetTodoDone changes it.
	edited := save(domain.TodoInput{ID: sooner.ID, Title: "Applied to Acme", DueDate: "2026-11-01"})
	if edited.CompletedAt != done.CompletedAt || edited.CreatedAt != sooner.CreatedAt {
		t.Fatalf("edited to-do = %#v", edited)
	}

	list, err := store.ListTodos(ctx)
	if err != nil {
		t.Fatalf("ListTodos() error = %v", err)
	}
	if got := []string{list[0].ID, list[1].ID, list[2].ID}; got[0] != later.ID || got[1] != undated.ID || got[2] != sooner.ID {
		t.Fatalf("order = %v, want dated open, undated open, completed", got)
	}

	if _, err := store.SaveTodo(ctx, domain.Todo{Title: "x", ApplicationID: uuid.NewString()}); err == nil {
		t.Fatal("SaveTodo() expected unknown resume error")
	}
	if _, err := store.SetTodoDone(ctx, uuid.NewString(), true); err == nil {
		t.Fatal("SetTodoDone() expected not found error")
	}

	if enabled, err := store.TodosEnabled(ctx); err != nil || !enabled {
		t.Fatalf("TodosEnabled() default = %v, %v", enabled, err)
	}
	if err := store.SetTodosEnabled(ctx, false); err != nil {
		t.Fatalf("SetTodosEnabled() error = %v", err)
	}

	backup, err := store.CreateProfileBackup(ctx)
	if err != nil {
		t.Fatalf("CreateProfileBackup() error = %v", err)
	}
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	decoded, err := domain.DecodeProfileBackup(data)
	if err != nil {
		t.Fatalf("DecodeProfileBackup() error = %v", err)
	}
	if err := store.DeleteTodo(ctx, undated.ID); err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if err := store.SetTodosEnabled(ctx, true); err != nil {
		t.Fatalf("SetTodosEnabled() error = %v", err)
	}
	if err := store.ReplaceProfileFromBackup(ctx, decoded); err != nil {
		t.Fatalf("ReplaceProfileFromBackup() error = %v", err)
	}
	restored, err := store.ListTodos(ctx)
	if err != nil || len(restored) != 3 || restored[0].ApplicationID != resume.Application.ID || restored[2].CompletedAt != done.CompletedAt {
		t.Fatalf("restored to-dos = %#v, %v", restored, err)
	}
	if enabled, _ := store.TodosEnabled(ctx); enabled {
		t.Fatal("restored backup should keep the to-do list turned off")
	}

	// Deleting a resume keeps its to-dos and unlinks them.
	if _, err := store.DeleteApplication(ctx, resume.Application.ID); err != nil {
		t.Fatalf("DeleteApplication() error = %v", err)
	}
	unlinked, err := store.getTodo(ctx, later.ID)
	if err != nil || unlinked.ApplicationID != "" {
		t.Fatalf("to-do after resume delete = %#v, %v", unlinked, err)
	}
}

func TestTodoInputValidation(t *testing.T) {
	for _, input := range []domain.TodoInput{
		{Title: "   "},
		{Title: "ok", DueDate: "2026-02-30"},
		{Title: "ok", DueDate: "next week"},
	} {
		if _, err := input.Validate(); err == nil {
			t.Errorf("Validate(%#v) expected an error", input)
		}
	}
}
