package main

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/ai"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/resume"
	"github.com/jscyril/tailorcv/internal/storage"
)

func TestValidateReviewedProposalsPreservesProviderCitations(t *testing.T) {
	generated := []domain.AIProposal{{TargetFactID: "fact", SupportingFactIDs: []string{"fact", "support"}, Text: "Original proposal text long enough to review."}}
	if err := validateReviewedProposals(generated, []domain.AIProposal{{TargetFactID: "fact", SupportingFactIDs: []string{"support", "fact"}, Text: "User edited proposal text that remains supported."}}); err != nil {
		t.Fatalf("validateReviewedProposals() error = %v", err)
	}
	if err := validateReviewedProposals(generated, []domain.AIProposal{{TargetFactID: "fact", SupportingFactIDs: []string{"fact"}, Text: "Changed citations are not allowed in review."}}); err == nil {
		t.Fatal("validateReviewedProposals() accepted changed citations")
	}
}

func TestGenerateAITailoringUsesOllamaAndPersistsValidatedRun(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(filepath.Join(t.TempDir(), "generate-ai.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	app := &App{ctx: ctx, store: store}
	profile, _ := (domain.ProfileInput{Name: "Ada Lovelace", Skills: []string{"Go"}}).Validate()
	_, _ = store.SaveProfile(ctx, profile)
	experience, _ := (domain.ExperienceInput{Company: "Example", Title: "Engineer", StartDate: "2024-01", Bullets: []domain.EvidenceBulletInput{{Text: "Reduced deployment time by 40% with an audited Go release pipeline", Verification: domain.VerificationVerified}}}).Validate()
	savedExperience, err := store.SaveExperience(ctx, experience)
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	factID := savedExperience.Bullets[0].ID
	job, _ := (domain.JobInput{Role: "Platform Engineer", Description: "Build reliable Go deployment systems and improve release operations for production services."}).Validate()
	savedJob, err := store.SaveJob(ctx, job)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	app.aiProviderFactory = func(string, string) (ai.Provider, error) {
		proposal := fmt.Sprintf(`{"schemaVersion":"tailorcv.ai.tailoring.v1","proposals":[{"targetFactId":%q,"supportingFactIds":[%q],"text":"Reduced deployment time by 40%% through an audited Go release pipeline."}]}`, factID, factID)
		return recordedAIProvider{response: []byte(proposal), inspect: func(request ai.Request) {
			if len(request.Facts) != 1 || request.Facts[0].ID != factID || request.Job.Role != "Platform Engineer" {
				t.Errorf("AI request = %#v", request)
			}
		}}, nil
	}
	run, err := app.GenerateAITailoring(domain.GenerateAITailoringInput{JobID: savedJob.ID, SelectedFactIDs: []string{factID}, Model: "recorded"})
	if err != nil {
		t.Fatalf("GenerateAITailoring() error = %v", err)
	}
	if !run.ValidationPassed || len(run.Proposals) != 1 || run.Proposals[0].TargetFactID != factID {
		t.Fatalf("AI run = %#v", run)
	}
	loaded, err := store.GetAIRun(ctx, run.ID)
	if err != nil || !loaded.ValidationPassed {
		t.Fatalf("persisted AI run = %#v, %v", loaded, err)
	}
}

type recordedAIProvider struct {
	response    []byte
	generateErr error
	inspect     func(ai.Request)
}

func (recordedAIProvider) Name() string                             { return "ollama" }
func (recordedAIProvider) Models(context.Context) ([]string, error) { return []string{"recorded"}, nil }
func (provider recordedAIProvider) Generate(_ context.Context, _ string, request ai.Request) ([]byte, error) {
	if provider.inspect != nil {
		provider.inspect(request)
	}
	if provider.generateErr != nil {
		return nil, provider.generateErr
	}
	return append([]byte(nil), provider.response...), nil
}

func TestGenerateAITailoringPersistsRecoverableProviderFailure(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(filepath.Join(t.TempDir(), "provider-failure.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	profile, _ := (domain.ProfileInput{Name: "Ada Lovelace", Skills: []string{"Go"}}).Validate()
	_, _ = store.SaveProfile(ctx, profile)
	experience, _ := (domain.ExperienceInput{Company: "Example", Title: "Engineer", StartDate: "2024-01", Bullets: []domain.EvidenceBulletInput{{Text: "Built reliable Go services for production systems", Verification: domain.VerificationVerified}}}).Validate()
	savedExperience, err := store.SaveExperience(ctx, experience)
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	job, _ := (domain.JobInput{Role: "Platform Engineer", Description: "Build reliable Go services for secure production systems and release operations."}).Validate()
	savedJob, err := store.SaveJob(ctx, job)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	app := &App{ctx: ctx, store: store, aiProviderFactory: func(string, string) (ai.Provider, error) {
		return recordedAIProvider{generateErr: fmt.Errorf("recorded provider unavailable")}, nil
	}}
	run, err := app.GenerateAITailoring(domain.GenerateAITailoringInput{JobID: savedJob.ID, SelectedFactIDs: []string{savedExperience.Bullets[0].ID}, Model: "recorded"})
	if err != nil {
		t.Fatalf("GenerateAITailoring() returned transport error instead of audit run: %v", err)
	}
	if run.ValidationPassed || run.FailureCategory != "provider" || len(run.ValidationErrors) != 1 || !strings.Contains(run.ValidationErrors[0], "provider unavailable") {
		t.Fatalf("provider failure run = %#v", run)
	}
	loaded, err := store.GetAIRun(ctx, run.ID)
	if err != nil || loaded.FailureCategory != "provider" {
		t.Fatalf("persisted provider failure = %#v, %v", loaded, err)
	}
}

func TestGenerateProjectReadmeBulletsReturnsOnlyValidatedDrafts(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(filepath.Join(t.TempDir(), "readme-bullets.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	project, _ := (domain.ProjectInput{Name: "TailorCV", RepositoryReadme: "TailorCV is a Go and React desktop app that creates evidence-backed resumes from local career data.", Skills: []string{"Go", "React"}}).Validate()
	savedProject, err := store.SaveProject(ctx, project)
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}
	factID := savedProject.ID + ":readme:1"
	proposal := fmt.Sprintf(`{"schemaVersion":"tailorcv.ai.tailoring.v1","proposals":[{"targetFactId":%q,"supportingFactIds":[%q],"text":"Built TailorCV, a Go and React desktop app for evidence-backed resumes."}]}`, factID, factID)
	app := &App{ctx: ctx, store: store, aiProviderFactory: func(string, string) (ai.Provider, error) {
		return recordedAIProvider{response: []byte(proposal)}, nil
	}}
	result, err := app.GenerateProjectReadmeBullets(domain.GenerateProjectReadmeBulletsInput{ProjectID: savedProject.ID, Model: "recorded"})
	if err != nil {
		t.Fatalf("GenerateProjectReadmeBullets() error = %v", err)
	}
	if got, want := result.Bullets, []string{"Built TailorCV, a Go and React desktop app for evidence-backed resumes."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Bullets = %#v, want %#v", got, want)
	}
}

func TestAcceptAITailoringCreatesImmutableVersionWithoutChangingEvidence(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(filepath.Join(t.TempDir(), "accept-ai.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	app := &App{ctx: ctx, store: store}
	profile, _ := (domain.ProfileInput{Name: "Ada Lovelace", Skills: []string{"Go"}}).Validate()
	if _, err := store.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	experience, _ := (domain.ExperienceInput{Company: "Example", Title: "Engineer", StartDate: "2024-01", Bullets: []domain.EvidenceBulletInput{{Text: "Reduced deployment time by 40% with an audited Go release pipeline", Verification: domain.VerificationVerified}}}).Validate()
	savedExperience, err := store.SaveExperience(ctx, experience)
	if err != nil {
		t.Fatalf("SaveExperience() error = %v", err)
	}
	factID := savedExperience.Bullets[0].ID
	job, _ := (domain.JobInput{Role: "Platform Engineer", Description: "Build reliable Go deployment systems and improve release operations for production services."}).Validate()
	savedJob, err := store.SaveJob(ctx, job)
	if err != nil {
		t.Fatalf("SaveJob() error = %v", err)
	}
	proposal := domain.AIProposal{TargetFactID: factID, SupportingFactIDs: []string{factID}, Text: "Reduced deployment time by 40% through an audited Go release pipeline."}
	run, err := store.SaveAIRun(ctx, domain.AIRun{JobID: savedJob.ID, Provider: "ollama", Model: "recorded", PromptVersion: "tailorcv.ai.prompt.v1", SchemaVersion: "tailorcv.ai.tailoring.v1", SelectedFactIDs: []string{factID}, ValidationPassed: true, Proposals: []domain.AIProposal{proposal}})
	if err != nil {
		t.Fatalf("SaveAIRun() error = %v", err)
	}
	result, err := app.AcceptAITailoring(domain.AcceptAITailoringInput{RunID: run.ID, TemplateID: resume.DefaultTemplateID, Proposals: []domain.AIProposal{proposal}})
	if err != nil {
		t.Fatalf("AcceptAITailoring() error = %v", err)
	}
	if !strings.Contains(result.Version.LatexSource, "Reduced deployment time by 40\\% through") {
		t.Fatalf("accepted source = %q", result.Version.LatexSource)
	}
	loadedExperiences, _ := store.ListExperiences(ctx)
	if loadedExperiences[0].Bullets[0].Text != experience.Bullets[0].Text {
		t.Fatalf("source evidence changed to %q", loadedExperiences[0].Bullets[0].Text)
	}
	acceptedRun, _ := store.GetAIRun(ctx, run.ID)
	if acceptedRun.ResumeVersionID != result.Version.ID || acceptedRun.AcceptedAt == "" {
		t.Fatalf("accepted AI run = %#v", acceptedRun)
	}
	if _, err := app.AcceptAITailoring(domain.AcceptAITailoringInput{RunID: run.ID, TemplateID: resume.DefaultTemplateID, Proposals: []domain.AIProposal{proposal}}); err == nil {
		t.Fatal("AcceptAITailoring() accepted the same run twice")
	}
}

func TestApplyAIProposalsChangesCopiesWithoutVerification(t *testing.T) {
	experiences := []domain.Experience{{Bullets: []domain.EvidenceBullet{{ID: "experience-fact", Text: "Original", Verification: domain.VerificationVerified}}}}
	projects := []domain.Project{{ID: "project", Description: "Original project", Bullets: []domain.EvidenceBullet{{ID: "project-fact", Text: "Original bullet", Verification: domain.VerificationVerified}}}}
	applyAIProposals([]domain.AIProposal{
		{TargetFactID: "experience-fact", Text: "Rewritten experience evidence"},
		{TargetFactID: "project", Text: "Rewritten project evidence"},
		{TargetFactID: "project-fact", Text: "Rewritten project bullet"},
	}, experiences, projects)
	if experiences[0].Bullets[0].Text != "Rewritten experience evidence" || projects[0].Description != "Rewritten project evidence" || projects[0].Bullets[0].Text != "Rewritten project bullet" {
		t.Fatalf("rewritten evidence = %#v, %#v", experiences, projects)
	}
	if experiences[0].Bullets[0].Verification != domain.VerificationVerified {
		t.Fatal("applyAIProposals() changed source verification metadata")
	}
}
