package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func TestResumeImportPreviewApplyAndDuplicates(t *testing.T) {
	app := newWorkflowTestApp(t)
	fixture, err := filepath.Abs(filepath.Join("internal", "resumeimport", "testdata", "jake.tex"))
	if err != nil {
		t.Fatal(err)
	}
	app.openFileDialog = func(context.Context, runtime.OpenDialogOptions) (string, error) { return fixture, nil }
	// An existing value wins over the imported one.
	if _, err := app.SaveProfile(domain.ProfileInput{Name: "Casey F.", Skills: []string{"go"}}); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	preview, err := app.PreviewResumeImport()
	if err != nil || preview.FileName != "jake.tex" || len(preview.Experiences) != 2 || len(preview.Projects) != 1 || len(preview.Educations) != 1 {
		t.Fatalf("PreviewResumeImport() = %#v, %v", preview, err)
	}
	if experiences, _ := app.ListExperiences(); len(experiences) != 0 {
		t.Fatal("preview must not save anything")
	}

	result, err := app.ApplyResumeImport(preview)
	if err != nil {
		t.Fatalf("ApplyResumeImport() error = %v", err)
	}
	if result.Experiences != 2 || result.Projects != 1 || result.Educations != 1 || result.Skills != 9 || result.Duplicates != 0 {
		t.Fatalf("result = %#v", result)
	}
	profile, _ := app.GetProfile()
	if profile.Name != "Casey F." || profile.Email != "casey@example.com" || profile.GitHubUsername != "caseyf" || len(profile.Skills) != 10 {
		t.Fatalf("merged profile = %#v", profile)
	}
	experiences, _ := app.ListExperiences()
	if len(experiences) != 2 || experiences[0].Bullets[0].Provenance != domain.ProvenanceImported {
		t.Fatalf("saved experiences = %#v", experiences)
	}

	again, err := app.ApplyResumeImport(preview)
	if err != nil || again.Duplicates != 4 || again.Experiences+again.Projects+again.Educations != 0 || again.Skills != 0 {
		t.Fatalf("second import = %#v, %v", again, err)
	}

	// The review step sends data back from the UI, so it is validated again.
	bad := preview
	bad.Experiences = []domain.ExperienceInput{{Title: "No company", StartDate: "2020-01"}}
	if _, err := app.ApplyResumeImport(bad); err == nil || !strings.Contains(err.Error(), "experience 1") {
		t.Fatalf("invalid entry should be rejected before saving, got %v", err)
	}
}

func TestTemplateImportCommentsOutPDFTeXOnlyLines(t *testing.T) {
	app := newWorkflowTestApp(t)
	fixture, _ := filepath.Abs(filepath.Join("internal", "resumeimport", "testdata", "jake.tex"))
	app.openFileDialog = func(context.Context, runtime.OpenDialogOptions) (string, error) { return fixture, nil }
	template, err := app.ImportResumeTemplate()
	if err != nil {
		t.Fatalf("ImportResumeTemplate() error = %v", err)
	}
	if strings.Contains(template.Source, "\n\\input{glyphtounicode}") || strings.Contains(template.Source, "\\pdfgentounicode") || !strings.Contains(template.Description, "commented out 2 pdfTeX-only") {
		t.Fatalf("imported template = %q / %q", template.Description, template.Source[:400])
	}
}
