package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/ai"
	"github.com/jscyril/tailorcv/internal/credentials"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/resume"
)

// CheckOllama verifies the configured endpoint and returns its locally
// available models. It never sends profile or job data.
func (a *App) CheckOllama(endpoint string) (domain.AIProviderStatus, error) {
	if err := a.ready(); err != nil {
		return domain.AIProviderStatus{}, err
	}
	provider, err := a.newAIProvider("ollama", endpoint)
	if err != nil {
		return domain.AIProviderStatus{}, err
	}
	validatedEndpoint, _ := domain.ValidateOllamaEndpoint(endpoint)
	ctx, cancel := context.WithTimeout(a.appContext(), 8*time.Second)
	defer cancel()
	models, err := provider.Models(ctx)
	if err != nil {
		return domain.AIProviderStatus{Provider: provider.Name(), Endpoint: validatedEndpoint, Models: []string{}, Message: err.Error()}, nil
	}
	message := fmt.Sprintf("Ollama is available with %d local model(s).", len(models))
	if len(models) == 0 {
		message = "Ollama is available, but no local models were found."
	}
	return domain.AIProviderStatus{Provider: provider.Name(), Endpoint: validatedEndpoint, Available: true, Models: models, Message: message}, nil
}

// CheckGemini verifies the OS-stored credential and returns models that
// currently support generateContent. It never returns the API key.
func (a *App) CheckGemini() (domain.AIProviderStatus, error) {
	if err := a.ready(); err != nil {
		return domain.AIProviderStatus{}, err
	}
	provider, err := a.newAIProvider("gemini", "")
	if err != nil {
		return domain.AIProviderStatus{Provider: "gemini", Models: []string{}, Message: err.Error()}, nil
	}
	ctx, cancel := context.WithTimeout(a.appContext(), 12*time.Second)
	defer cancel()
	models, err := provider.Models(ctx)
	if err != nil {
		return domain.AIProviderStatus{Provider: "gemini", Models: []string{}, Message: err.Error()}, nil
	}
	message := fmt.Sprintf("Gemini is available with %d generation model(s).", len(models))
	if len(models) == 0 {
		message = "Gemini is available, but no generation models were found."
	}
	return domain.AIProviderStatus{Provider: "gemini", Available: true, Models: models, Message: message}, nil
}

func (a *App) GetAISettings() (domain.AISettings, error) {
	if err := a.ready(); err != nil {
		return domain.AISettings{}, err
	}
	return a.store.GetAISettings(a.appContext())
}

func (a *App) SaveAISettings(input domain.AISettings) (domain.AISettings, error) {
	if err := a.ready(); err != nil {
		return domain.AISettings{}, err
	}
	return a.store.SaveAISettings(a.appContext(), input)
}

func (a *App) GetGeminiCredentialStatus() (domain.CredentialStatus, error) {
	if err := a.ready(); err != nil {
		return domain.CredentialStatus{}, err
	}
	_, err := a.credentialStore().Get(credentials.Service, credentials.GeminiAPIKey)
	if errors.Is(err, credentials.ErrNotFound) {
		return domain.CredentialStatus{Message: "Gemini API key is not configured."}, nil
	}
	if err != nil {
		return domain.CredentialStatus{}, fmt.Errorf("read Gemini credential from OS keyring: %w", err)
	}
	return domain.CredentialStatus{Configured: true, Message: "Gemini API key is stored in the OS keyring."}, nil
}

func (a *App) SaveGeminiAPIKey(apiKey string) (domain.CredentialStatus, error) {
	if err := a.ready(); err != nil {
		return domain.CredentialStatus{}, err
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" || len(apiKey) > 4096 {
		return domain.CredentialStatus{}, fmt.Errorf("enter a valid Gemini API key")
	}
	if err := a.credentialStore().Set(credentials.Service, credentials.GeminiAPIKey, apiKey); err != nil {
		return domain.CredentialStatus{}, fmt.Errorf("save Gemini credential in OS keyring: %w", err)
	}
	return domain.CredentialStatus{Configured: true, Message: "Gemini API key is stored in the OS keyring."}, nil
}

func (a *App) DeleteGeminiAPIKey() (domain.CredentialStatus, error) {
	if err := a.ready(); err != nil {
		return domain.CredentialStatus{}, err
	}
	err := a.credentialStore().Delete(credentials.Service, credentials.GeminiAPIKey)
	if err != nil && !errors.Is(err, credentials.ErrNotFound) {
		return domain.CredentialStatus{}, fmt.Errorf("delete Gemini credential from OS keyring: %w", err)
	}
	return domain.CredentialStatus{Message: "Gemini API key is not configured."}, nil
}

// GenerateAITailoring sends only normalized job requirements and explicitly
// selected evidence to the configured provider. Provider and validation failures are persisted
// as auditable runs and returned to the review UI.
func (a *App) GenerateAITailoring(input domain.GenerateAITailoringInput) (domain.AIRun, error) {
	if err := a.ready(); err != nil {
		return domain.AIRun{}, err
	}
	validated, err := input.Validate()
	if err != nil {
		return domain.AIRun{}, err
	}
	request, err := a.aiRequest(validated.JobID, validated.SelectedFactIDs)
	if err != nil {
		return domain.AIRun{}, err
	}
	provider, err := a.newAIProvider(validated.Provider, validated.Endpoint)
	if err != nil {
		return domain.AIRun{}, err
	}
	run := domain.AIRun{
		JobID: validated.JobID, Provider: provider.Name(), Model: validated.Model,
		PromptVersion: ai.PromptVersion, SchemaVersion: ai.SchemaVersion,
		SelectedFactIDs:  append([]string(nil), validated.SelectedFactIDs...),
		ValidationErrors: []string{}, Proposals: []domain.AIProposal{},
	}
	raw, generateErr := a.generateProposals(provider, validated.Model, request)
	if generateErr != nil {
		run.FailureCategory = "provider"
		run.ValidationErrors = []string{"provider request failed: " + generateErr.Error()}
		return a.store.SaveAIRun(a.appContext(), run)
	}
	validation := ai.DecodeAndValidate(request, raw)
	run.ValidationErrors = validation.Errors
	run.ValidationPassed = len(validation.Errors) == 0
	if run.ValidationPassed {
		run.Proposals = ai.NormalizeProposals(validation.Response.Proposals)
	} else {
		run.FailureCategory = "validation"
	}
	return a.store.SaveAIRun(a.appContext(), run)
}

// GenerateProjectReadmeBullets turns a saved GitHub README snapshot into up to
// three editable draft bullets. The returned text is deliberately not stored.
func (a *App) GenerateProjectReadmeBullets(input domain.GenerateProjectReadmeBulletsInput) (domain.ProjectReadmeBulletsResult, error) {
	if err := a.ready(); err != nil {
		return domain.ProjectReadmeBulletsResult{}, err
	}
	validated, err := input.Validate()
	if err != nil {
		return domain.ProjectReadmeBulletsResult{}, err
	}
	projects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return domain.ProjectReadmeBulletsResult{}, err
	}
	var project domain.Project
	for _, candidate := range projects {
		if candidate.ID == validated.ProjectID {
			project = candidate
			break
		}
	}
	if project.ID == "" {
		return domain.ProjectReadmeBulletsResult{}, fmt.Errorf("project not found")
	}
	if strings.TrimSpace(project.RepositoryReadme) == "" {
		return domain.ProjectReadmeBulletsResult{}, fmt.Errorf("this project has no README snapshot; sync GitHub first")
	}
	provider, err := a.newAIProvider(validated.Provider, validated.Endpoint)
	if err != nil {
		return domain.ProjectReadmeBulletsResult{}, err
	}
	request := ai.ProjectReadmeBulletsRequest(project)
	raw, err := a.generateProposals(provider, validated.Model, request)
	if err != nil {
		return domain.ProjectReadmeBulletsResult{}, fmt.Errorf("generate README bullets: %w", err)
	}
	validation := ai.DecodeAndValidate(request, raw)
	if len(validation.Errors) > 0 {
		return domain.ProjectReadmeBulletsResult{}, fmt.Errorf("README bullets failed evidence validation: %s", strings.Join(validation.Errors, "; "))
	}
	proposals := ai.NormalizeProposals(validation.Response.Proposals)
	if len(proposals) > 3 {
		return domain.ProjectReadmeBulletsResult{}, fmt.Errorf("README generation returned more than three bullets")
	}
	result := domain.ProjectReadmeBulletsResult{Bullets: make([]string, 0, len(proposals))}
	for _, proposal := range proposals {
		result.Bullets = append(result.Bullets, proposal.Text)
	}
	return result, nil
}

// generateProposals runs the validate-and-revise agent when the provider can
// call tools, and a single structured request otherwise.
func (a *App) generateProposals(provider ai.Provider, model string, request ai.Request) ([]byte, error) {
	if agent, ok := provider.(ai.AgentProvider); ok {
		ctx, done := a.beginAIRequest(4 * time.Minute)
		defer done()
		return ai.RunProposalAgent(ctx, agent, model, request)
	}
	ctx, done := a.beginAIRequest(90 * time.Second)
	defer done()
	return provider.Generate(ctx, model, request)
}

// beginAIRequest cancels any running AI request and returns a context that
// CancelAITailoring can stop. Call done when the request finishes.
func (a *App) beginAIRequest(timeout time.Duration) (context.Context, func()) {
	ctx, cancel := context.WithTimeout(a.appContext(), timeout)
	a.aiMu.Lock()
	if a.aiCancel != nil {
		a.aiCancel()
	}
	a.aiGeneration++
	generation := a.aiGeneration
	a.aiCancel = cancel
	a.aiMu.Unlock()
	return ctx, func() {
		cancel()
		a.aiMu.Lock()
		if a.aiGeneration == generation {
			a.aiCancel = nil
		}
		a.aiMu.Unlock()
	}
}

// GenerateCoverLetter drafts a cover letter whose body paragraphs come from
// the provider and pass the same evidence validation as resume proposals.
// The draft is returned for editing and is not saved.
func (a *App) GenerateCoverLetter(input domain.GenerateCoverLetterInput) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	validated, err := input.Validate()
	if err != nil {
		return "", err
	}
	job, request, err := a.coverLetterRequest(validated.ApplicationID)
	if err != nil {
		return "", err
	}
	provider, err := a.newAIProvider(validated.Provider, validated.Endpoint)
	if err != nil {
		return "", err
	}
	request.Task = ai.CoverLetterTask
	raw, err := a.generateProposals(provider, validated.Model, request)
	if err != nil {
		return "", fmt.Errorf("generate cover letter: %w", err)
	}
	validation := ai.DecodeAndValidate(request, raw)
	if len(validation.Errors) > 0 {
		return "", fmt.Errorf("cover letter failed evidence validation: %s", strings.Join(validation.Errors, "; "))
	}
	proposals := ai.NormalizeProposals(validation.Response.Proposals)
	if len(proposals) > 4 {
		proposals = proposals[:4]
	}
	body := make([]string, 0, len(proposals))
	for _, proposal := range proposals {
		body = append(body, proposal.Text)
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return "", err
	}
	return domain.ComposeCoverLetter(profile, job, body), nil
}

// DraftCoverLetter builds a cover letter without AI: the body lists the
// application's selected evidence word for word.
func (a *App) DraftCoverLetter(applicationID string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	job, request, err := a.coverLetterRequest(applicationID)
	if err != nil {
		return "", err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return "", err
	}
	lines := []string{"Relevant work from my background:"}
	for index, fact := range request.Facts {
		if index == 5 {
			break
		}
		lines = append(lines, "- "+strings.Join(strings.Fields(fact.Text), " "))
	}
	return domain.ComposeCoverLetter(profile, job, []string{strings.Join(lines, "\n")}), nil
}

// coverLetterRequest loads a job-tailored application and its selected
// evidence as an AI request.
func (a *App) coverLetterRequest(applicationID string) (domain.Job, ai.Request, error) {
	application, err := a.store.GetApplication(a.appContext(), strings.TrimSpace(applicationID))
	if err != nil {
		return domain.Job{}, ai.Request{}, err
	}
	if application.JobID == "" {
		return domain.Job{}, ai.Request{}, fmt.Errorf("cover letters need a job; tailor this resume to a job first")
	}
	if len(application.SelectedFactIDs) == 0 {
		return domain.Job{}, ai.Request{}, fmt.Errorf("select evidence for this job before drafting a cover letter")
	}
	job, err := a.store.GetJob(a.appContext(), application.JobID)
	if err != nil {
		return domain.Job{}, ai.Request{}, err
	}
	request, err := a.aiRequest(job.ID, application.SelectedFactIDs)
	return job, request, err
}

func (a *App) newAIProvider(name, endpoint string) (ai.Provider, error) {
	if a.aiProviderFactory != nil {
		return a.aiProviderFactory(name, endpoint)
	}
	switch name {
	case "ollama":
		return ai.NewOllama(endpoint, nil)
	case "gemini":
		apiKey, err := a.credentialStore().Get(credentials.Service, credentials.GeminiAPIKey)
		if errors.Is(err, credentials.ErrNotFound) {
			return nil, fmt.Errorf("Gemini API key is not configured")
		}
		if err != nil {
			return nil, fmt.Errorf("read Gemini credential from OS keyring: %w", err)
		}
		return ai.NewGemini(apiKey, nil)
	default:
		return nil, fmt.Errorf("AI provider %q is not supported", name)
	}
}

func (a *App) credentialStore() credentials.Store {
	if a.credentials == nil {
		return credentials.OSStore{}
	}
	return a.credentials
}

func (a *App) CancelAITailoring() {
	a.aiMu.Lock()
	defer a.aiMu.Unlock()
	if a.aiCancel != nil {
		a.aiCancel()
	}
}

func (a *App) ListAIRuns() ([]domain.AIRun, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.ListAIRuns(a.appContext())
}

// AcceptAITailoring revalidates user-edited proposals against current selected
// evidence, renders them through a trusted template, and creates a new
// immutable resume version. It never writes generated wording back to profile
// evidence or marks it verified.
func (a *App) AcceptAITailoring(input domain.AcceptAITailoringInput) (domain.ApplicationResumeResult, error) {
	if err := a.ready(); err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	validated, err := input.Validate()
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	run, err := a.store.GetAIRun(a.appContext(), validated.RunID)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	if !run.ValidationPassed {
		return domain.ApplicationResumeResult{}, fmt.Errorf("AI proposals did not pass evidence validation")
	}
	if run.ResumeVersionID != "" {
		return domain.ApplicationResumeResult{}, fmt.Errorf("AI proposals were already accepted into a resume version")
	}
	if err := validateReviewedProposals(run.Proposals, validated.Proposals); err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	request, err := a.aiRequest(run.JobID, run.SelectedFactIDs)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	proposals := ai.NormalizeProposals(validated.Proposals)
	validationErrors := ai.ValidateProposals(request, ai.Response{SchemaVersion: run.SchemaVersion, Proposals: proposals})
	if len(validationErrors) > 0 {
		return domain.ApplicationResumeResult{}, fmt.Errorf("edited AI proposals failed validation: %s", strings.Join(validationErrors, "; "))
	}
	job, err := a.store.GetJob(a.appContext(), run.JobID)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	template, err := a.resumeTemplate(validated.TemplateID)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	experiences, err := a.store.ListExperiences(a.appContext())
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	projects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	educations, err := a.store.ListEducations(a.appContext())
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	certifications, err := a.store.ListCertifications(a.appContext())
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	achievements, err := a.store.ListAchievements(a.appContext())
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	selectedExperiences, selectedProjects, err := selectResumeEvidence(run.SelectedFactIDs, experiences, projects)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	applyAIProposals(proposals, selectedExperiences, selectedProjects)
	source := resume.Render(template.Source, resume.Data{Profile: profile, Experiences: selectedExperiences, Projects: selectedProjects, Educations: educations, Certifications: certifications, Achievements: achievements})
	ranking, err := a.rankSelectedEvidence(job, profile, experiences, projects, run.SelectedFactIDs)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	result, err := a.store.CreateResumeVersion(a.appContext(), job.ID, run.SelectedFactIDs, template.ID, job.Description, source, ranking)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	if err := a.store.MarkAIRunAccepted(a.appContext(), run.ID, result.Version.ID); err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	return result, nil
}

func (a *App) aiRequest(jobID string, selectedFactIDs []string) (ai.Request, error) {
	job, err := a.store.GetJob(a.appContext(), jobID)
	if err != nil {
		return ai.Request{}, err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return ai.Request{}, err
	}
	experiences, err := a.store.ListExperiences(a.appContext())
	if err != nil {
		return ai.Request{}, err
	}
	projects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return ai.Request{}, err
	}
	selectedExperiences, selectedProjects, err := selectResumeEvidence(selectedFactIDs, experiences, projects)
	if err != nil {
		return ai.Request{}, err
	}
	analysis, err := domain.AnalyzeCareerEvidence(domain.JobAnalysisInput{ID: job.ID, Company: job.Company, Role: job.Role, Description: job.Description}, profile.Skills, experiences, projects)
	if err != nil {
		return ai.Request{}, err
	}
	facts := make([]ai.Fact, 0, len(selectedFactIDs))
	for _, experience := range selectedExperiences {
		label := strings.TrimSpace(experience.Title + " · " + experience.Company)
		for _, bullet := range experience.Bullets {
			facts = append(facts, ai.Fact{ID: bullet.ID, SourceType: "experience", SourceLabel: label, Text: bullet.Text, Technologies: domain.DetectKnownSkills(bullet.Text)})
		}
	}
	for _, project := range selectedProjects {
		technologies := append([]string(nil), project.Skills...)
		for _, language := range project.DetectedLanguages {
			technologies = appendUniqueStrings(technologies, language.Name)
		}
		if containsString(selectedFactIDs, project.ID) {
			text := project.Description
			if strings.TrimSpace(text) == "" {
				text = project.Name
			}
			facts = append(facts, ai.Fact{ID: project.ID, SourceType: "project", SourceLabel: project.Name, Text: text, Technologies: appendUniqueStrings(technologies, domain.DetectKnownSkills(text)...)})
		}
		for _, bullet := range project.Bullets {
			facts = append(facts, ai.Fact{ID: bullet.ID, SourceType: "project", SourceLabel: project.Name, Text: bullet.Text, Technologies: appendUniqueStrings(technologies, domain.DetectKnownSkills(bullet.Text)...)})
		}
	}
	sort.SliceStable(facts, func(left, right int) bool { return facts[left].ID < facts[right].ID })
	return ai.NewRequest(analysis, facts), nil
}

func validateReviewedProposals(generated, reviewed []domain.AIProposal) error {
	allowed := make(map[string]domain.AIProposal, len(generated))
	for _, proposal := range generated {
		allowed[proposal.TargetFactID] = proposal
	}
	seen := make(map[string]struct{}, len(reviewed))
	for _, proposal := range reviewed {
		original, exists := allowed[strings.TrimSpace(proposal.TargetFactID)]
		if !exists {
			return fmt.Errorf("reviewed proposal targets evidence that was not proposed by the provider")
		}
		if _, duplicate := seen[original.TargetFactID]; duplicate {
			return fmt.Errorf("reviewed proposals contain duplicate target evidence")
		}
		seen[original.TargetFactID] = struct{}{}
		if !sameStringSet(original.SupportingFactIDs, proposal.SupportingFactIDs) {
			return fmt.Errorf("reviewed proposal citations cannot be changed")
		}
	}
	return nil
}

func applyAIProposals(proposals []domain.AIProposal, experiences []domain.Experience, projects []domain.Project) {
	textByFact := make(map[string]string, len(proposals))
	for _, proposal := range proposals {
		textByFact[proposal.TargetFactID] = proposal.Text
	}
	for experienceIndex := range experiences {
		for bulletIndex := range experiences[experienceIndex].Bullets {
			if text, exists := textByFact[experiences[experienceIndex].Bullets[bulletIndex].ID]; exists {
				experiences[experienceIndex].Bullets[bulletIndex].Text = text
			}
		}
	}
	for projectIndex := range projects {
		if text, exists := textByFact[projects[projectIndex].ID]; exists {
			projects[projectIndex].Description = text
		}
		for bulletIndex := range projects[projectIndex].Bullets {
			if text, exists := textByFact[projects[projectIndex].Bullets[bulletIndex].ID]; exists {
				projects[projectIndex].Bullets[bulletIndex].Text = text
			}
		}
	}
}

func appendUniqueStrings(values []string, additions ...string) []string {
	for _, addition := range additions {
		if strings.TrimSpace(addition) == "" || containsStringFold(values, addition) {
			continue
		}
		values = append(values, addition)
	}
	return values
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsStringFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	values := make(map[string]struct{}, len(left))
	for _, value := range left {
		values[strings.TrimSpace(value)] = struct{}{}
	}
	for _, value := range right {
		if _, exists := values[strings.TrimSpace(value)]; !exists {
			return false
		}
	}
	return true
}

// FixLatexWithAI lets an agent edit failing LaTeX until the local compiler
// accepts it. The fix is returned as an unsaved draft for the editor.
func (a *App) FixLatexWithAI(input domain.FixLatexInput) (domain.LatexFixResult, error) {
	if err := a.ready(); err != nil {
		return domain.LatexFixResult{}, err
	}
	validated, err := input.Validate()
	if err != nil {
		return domain.LatexFixResult{}, err
	}
	provider, err := a.newAIProvider(validated.Provider, validated.Endpoint)
	if err != nil {
		return domain.LatexFixResult{}, err
	}
	agent, ok := provider.(ai.AgentProvider)
	if !ok {
		return domain.LatexFixResult{}, fmt.Errorf("AI fixes need a provider that can call tools; switch to Gemini")
	}
	ctx, done := a.beginAIRequest(4 * time.Minute)
	defer done()
	initial, _, err := a.compile(ctx, validated.Source)
	if err != nil {
		return domain.LatexFixResult{}, err
	}
	if initial.Success {
		return domain.LatexFixResult{}, fmt.Errorf("this source already compiles")
	}
	var last domain.CompileResult
	var lastPDF []byte
	compile := func(ctx context.Context, source string) (bool, string, error) {
		result, pdf, err := a.compile(ctx, source)
		if err != nil {
			return false, "", err
		}
		last, lastPDF = result, pdf
		return result.Success, compileReport(result), nil
	}
	fixed, err := ai.RunLatexFixAgent(ctx, agent, validated.Model, validated.Source, compileReport(initial), compile)
	if err != nil {
		return domain.LatexFixResult{}, fmt.Errorf("fix LaTeX: %w", err)
	}
	a.setLastPDF(lastPDF)
	return domain.LatexFixResult{Source: fixed, Result: last}, nil
}

// compileReport gives the model diagnostics plus the end of the log, where
// TeX reports the error that stopped compilation.
func compileReport(result domain.CompileResult) string {
	lines := make([]string, 0, len(result.Diagnostics)+1)
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("line %d %s: %s", diagnostic.Line, diagnostic.Severity, diagnostic.Message))
	}
	log := result.Log
	if len(log) > 4000 {
		log = log[len(log)-4000:]
	}
	return strings.Join(append(lines, log), "\n")
}

// SuggestEvidenceWithAI lets an agent choose evidence for a job. It sees only
// the selectable facts in the job's ranked list, which the job screen already shows.
func (a *App) SuggestEvidenceWithAI(input domain.SuggestEvidenceInput) ([]string, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	validated, err := input.Validate()
	if err != nil {
		return nil, err
	}
	job, err := a.store.GetJob(a.appContext(), validated.JobID)
	if err != nil {
		return nil, err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return nil, err
	}
	experiences, err := a.store.ListExperiences(a.appContext())
	if err != nil {
		return nil, err
	}
	projects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return nil, err
	}
	analysis, err := a.analyzeJob(job, profile, experiences, projects)
	if err != nil {
		return nil, err
	}
	candidates := make([]ai.Fact, 0, len(analysis.RankedEvidence))
	for _, evidence := range analysis.RankedEvidence {
		if evidence.Selectable {
			candidates = append(candidates, ai.Fact{ID: evidence.FactID, SourceType: evidence.SourceType, SourceLabel: evidence.SourceLabel, Text: evidence.Text, Technologies: evidence.MatchedSkills})
		}
	}
	provider, err := a.newAIProvider(validated.Provider, validated.Endpoint)
	if err != nil {
		return nil, err
	}
	agent, ok := provider.(ai.AgentProvider)
	if !ok {
		return nil, fmt.Errorf("AI evidence selection needs a provider that can call tools; switch to Gemini")
	}
	ctx, done := a.beginAIRequest(4 * time.Minute)
	defer done()
	selected, err := ai.RunEvidenceSelectionAgent(ctx, agent, validated.Model, ai.NewRequest(analysis, nil).Job, candidates, 8)
	if err != nil {
		return nil, fmt.Errorf("choose evidence: %w", err)
	}
	return selected, nil
}

// DraftFollowUpEmail drafts an unsaved follow-up email from the application's
// job and selected evidence. Application notes are never sent.
func (a *App) DraftFollowUpEmail(input domain.DraftFollowUpInput) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	validated, err := input.Validate()
	if err != nil {
		return "", err
	}
	application, err := a.store.GetApplication(a.appContext(), validated.ApplicationID)
	if err != nil {
		return "", err
	}
	job, request, err := a.coverLetterRequest(application.ID)
	if err != nil {
		return "", err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return "", err
	}
	provider, err := a.newAIProvider(validated.Provider, validated.Endpoint)
	if err != nil {
		return "", err
	}
	agent, ok := provider.(ai.AgentProvider)
	if !ok {
		return "", fmt.Errorf("AI follow-ups need a provider that can call tools; switch to Gemini")
	}
	ctx, done := a.beginAIRequest(4 * time.Minute)
	defer done()
	followUp := ai.FollowUp{Company: job.Company, Role: job.Role, AppliedOn: application.AppliedOn, Sender: profile.Name}
	email, err := ai.RunFollowUpAgent(ctx, agent, validated.Model, followUp, request.Facts)
	if err != nil {
		return "", fmt.Errorf("draft follow-up: %w", err)
	}
	return email, nil
}
