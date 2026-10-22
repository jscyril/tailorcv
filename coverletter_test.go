package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/ai"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/resume"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func TestCoverLetterDraftAIValidationSaveAndExport(t *testing.T) {
	app := newWorkflowTestApp(t)
	if _, err := app.SaveProfile(domain.ProfileInput{Name: "Ada Lovelace", Skills: []string{"Go"}}); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	experience, err := app.SaveExperience(domain.ExperienceInput{
		Company: "Example", Title: "Engineer", StartDate: "2024-01",
		Bullets: []domain.EvidenceBulletInput{{Text: "Reduced deployment time by 40% with an audited Go release pipeline", Verification: domain.VerificationVerified}},
	})
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	analysis, err := app.AnalyzeJobDescription(domain.JobAnalysisInput{Company: "Fictional Cloud", Role: "Platform Engineer", Description: "Build reliable Go deployment systems and improve release operations for production services."})
	if err != nil {
		t.Fatalf("AnalyzeJobDescription() error = %v", err)
	}
	factID := experience.Bullets[0].ID
	created, err := app.CreateResumeVersion(domain.CreateResumeVersionInput{JobID: analysis.Job.ID, SelectedFactIDs: []string{factID}, TemplateID: resume.DefaultTemplateID})
	if err != nil {
		t.Fatalf("CreateResumeVersion() error = %v", err)
	}
	applicationID := created.Application.ID

	// Without AI the body is the selected evidence, word for word.
	draft, err := app.DraftCoverLetter(applicationID)
	if err != nil {
		t.Fatalf("DraftCoverLetter() error = %v", err)
	}
	for _, expected := range []string{"Dear Fictional Cloud hiring team,", "the Platform Engineer role at Fictional Cloud", "- Reduced deployment time by 40% with an audited Go release pipeline", "Sincerely,\nAda Lovelace"} {
		if !strings.Contains(draft, expected) {
			t.Errorf("draft missing %q:\n%s", expected, draft)
		}
	}

	// AI paragraphs pass the same evidence validation as resume proposals.
	paragraph := "In my recent role I reduced deployment time by 40% by building an audited Go release pipeline."
	respond := func(text string) {
		app.aiProviderFactory = func(string, string) (ai.Provider, error) {
			return recordedAIProvider{response: []byte(fmt.Sprintf(`{"schemaVersion":"%s","proposals":[{"targetFactId":%q,"supportingFactIds":[%q],"text":%q}]}`, ai.SchemaVersion, factID, factID, text))}, nil
		}
	}
	input := domain.GenerateCoverLetterInput{ApplicationID: applicationID, Provider: "ollama", Model: "recorded", Endpoint: domain.DefaultOllamaEndpoint}
	respond(paragraph)
	generated, err := app.GenerateCoverLetter(input)
	if err != nil || !strings.Contains(generated, paragraph) || !strings.HasPrefix(generated, "Dear Fictional Cloud hiring team,") {
		t.Fatalf("GenerateCoverLetter() = %q, %v", generated, err)
	}
	respond("In my recent role I reduced deployment time by 90% using Kubernetes and Go.")
	if _, err := app.GenerateCoverLetter(input); err == nil || !strings.Contains(err.Error(), "evidence validation") {
		t.Fatalf("invented metric and technology should fail validation, got %v", err)
	}

	saved, err := app.SaveCoverLetter(domain.SaveCoverLetterInput{ApplicationID: applicationID, Text: generated})
	if err != nil || saved.CoverLetter != generated {
		t.Fatalf("SaveCoverLetter() = %#v, %v", saved, err)
	}

	app.compiler = recordedCompiler{result: domain.CompileResult{Success: true}, pdf: []byte("%PDF-letter")}
	destination := filepath.Join(t.TempDir(), "letter")
	app.saveFileDialog = func(_ context.Context, options runtime.SaveDialogOptions) (string, error) {
		if options.DefaultFilename != "Ada Lovelace - Fictional Cloud - Cover letter.pdf" {
			t.Errorf("default file name = %q", options.DefaultFilename)
		}
		return destination, nil
	}
	exported, err := app.ExportCoverLetterPDF(generated, "Ada Lovelace - Fictional Cloud - Cover letter")
	if err != nil || exported.Path != destination+".pdf" {
		t.Fatalf("ExportCoverLetterPDF() = %#v, %v", exported, err)
	}
	if data, err := os.ReadFile(exported.Path); err != nil || string(data) != "%PDF-letter" {
		t.Fatalf("exported PDF = %q, %v", data, err)
	}
	app.compiler = recordedCompiler{result: domain.CompileResult{Diagnostics: []domain.CompileDiagnostic{{Severity: "warning", Message: "minor"}, {Severity: "error", Message: "Undefined control sequence"}}}}
	if _, err := app.ExportCoverLetterPDF(generated, "x"); err == nil || !strings.Contains(err.Error(), "Undefined control sequence") {
		t.Fatalf("compile failure should name the first error, got %v", err)
	}

	general, err := app.CreateResume(domain.CreateResumeInput{Name: "General", TemplateID: resume.DefaultTemplateID, LatexSource: "source"})
	if err != nil {
		t.Fatalf("CreateResume() error = %v", err)
	}
	if _, err := app.DraftCoverLetter(general.Application.ID); err == nil {
		t.Fatal("a general resume has no job to write a cover letter for")
	}
}
