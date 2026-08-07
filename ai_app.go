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
	ctx, cancel := context.WithTimeout(a.appContext(), 90*time.Second)
	a.aiMu.Lock()
	if a.aiCancel != nil {
		a.aiCancel()
	}
	a.aiGeneration++
	generation := a.aiGeneration
	a.aiCancel = cancel
	a.aiMu.Unlock()
	defer func() {
		cancel()
		a.aiMu.Lock()
		if a.aiGeneration == generation {
			a.aiCancel = nil
		}
		a.aiMu.Unlock()
	}()
	raw, generateErr := provider.Generate(ctx, validated.Model, request)
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
