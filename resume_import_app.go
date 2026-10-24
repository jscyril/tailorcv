package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/ai"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/resumeimport"
)

// PreviewResumeImport asks for an existing .tex resume and reads it without
// saving anything. An empty FileName means the user cancelled.
func (a *App) PreviewResumeImport() (domain.ResumeImport, error) {
	if err := a.ready(); err != nil {
		return domain.ResumeImport{}, err
	}
	path, data, err := a.chooseTexFile("Import your existing resume (.tex)")
	if err != nil || path == "" {
		return domain.ResumeImport{}, err
	}
	a.importMu.Lock()
	a.lastImportSource = string(data)
	a.importMu.Unlock()
	result := resumeimport.Parse(string(data))
	result.FileName = filepath.Base(path)
	return result, nil
}

// PreviewResumeImportWithAI re-reads the file from the last import preview
// with an agent, for resumes the rule-based parser could not fully read.
// The whole file is sent to the provider; nothing is saved.
func (a *App) PreviewResumeImportWithAI(input domain.ResumeImportAIInput) (domain.ResumeImport, error) {
	if err := a.ready(); err != nil {
		return domain.ResumeImport{}, err
	}
	validated, err := input.Validate()
	if err != nil {
		return domain.ResumeImport{}, err
	}
	a.importMu.Lock()
	source := a.lastImportSource
	a.importMu.Unlock()
	if strings.TrimSpace(source) == "" {
		return domain.ResumeImport{}, fmt.Errorf("choose a resume file to import first")
	}
	if len(source) > 256<<10 {
		return domain.ResumeImport{}, fmt.Errorf("resume file is too large for AI import")
	}
	provider, err := a.newAIProvider(validated.Provider, validated.Endpoint)
	if err != nil {
		return domain.ResumeImport{}, err
	}
	agent, ok := provider.(ai.AgentProvider)
	if !ok {
		return domain.ResumeImport{}, fmt.Errorf("AI import needs a provider that can call tools; switch to Gemini")
	}
	ctx, done := a.beginAIRequest(4 * time.Minute)
	defer done()
	result, err := ai.RunResumeImportAgent(ctx, agent, validated.Model, source, resumeimport.PlainText(source))
	if err != nil {
		return domain.ResumeImport{}, fmt.Errorf("read resume with AI: %w", err)
	}
	result.FileName = validated.FileName
	return result, nil
}

// ApplyResumeImport saves the reviewed import. Every entry is validated
// before anything is written. Profile fields that already have a value are
// kept, new skills are added, and entries already in the profile are skipped.
func (a *App) ApplyResumeImport(data domain.ResumeImport) (domain.ResumeImportResult, error) {
	if err := a.ready(); err != nil {
		return domain.ResumeImportResult{}, err
	}
	for index, item := range data.Experiences {
		if _, err := item.Validate(); err != nil {
			return domain.ResumeImportResult{}, fmt.Errorf("experience %d: %w", index+1, err)
		}
	}
	for index, item := range data.Projects {
		if _, err := item.Validate(); err != nil {
			return domain.ResumeImportResult{}, fmt.Errorf("project %d: %w", index+1, err)
		}
	}
	for index, item := range data.Educations {
		if _, err := item.Validate(); err != nil {
			return domain.ResumeImportResult{}, fmt.Errorf("education %d: %w", index+1, err)
		}
	}
	current, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return domain.ResumeImportResult{}, err
	}
	profile, addedSkills := mergeImportedProfile(current.Input(), data.Profile)
	if _, err := profile.Validate(); err != nil {
		return domain.ResumeImportResult{}, fmt.Errorf("profile: %w", err)
	}
	experiences, err := a.store.ListExperiences(a.appContext())
	if err != nil {
		return domain.ResumeImportResult{}, err
	}
	projects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return domain.ResumeImportResult{}, err
	}
	educations, err := a.store.ListEducations(a.appContext())
	if err != nil {
		return domain.ResumeImportResult{}, err
	}

	result := domain.ResumeImportResult{Skills: addedSkills}
	if _, err := a.SaveProfile(profile); err != nil {
		return result, err
	}
	// shortcut: entries are saved one by one, not in one transaction; validation above makes a partial import unlikely.
	for _, item := range data.Experiences {
		if containsKey(experiences, func(existing domain.Experience) bool {
			return sameText(existing.Company, item.Company) && sameText(existing.Title, item.Title) && existing.StartDate == item.StartDate
		}) {
			result.Duplicates++
			continue
		}
		if _, err := a.SaveExperience(item); err != nil {
			return result, fmt.Errorf("save %s at %s: %w", item.Title, item.Company, err)
		}
		result.Experiences++
	}
	for _, item := range data.Projects {
		if containsKey(projects, func(existing domain.Project) bool { return sameText(existing.Name, item.Name) }) {
			result.Duplicates++
			continue
		}
		if _, err := a.SaveProject(item); err != nil {
			return result, fmt.Errorf("save project %s: %w", item.Name, err)
		}
		result.Projects++
	}
	for _, item := range data.Educations {
		if containsKey(educations, func(existing domain.Education) bool {
			return sameText(existing.Institution, item.Institution) && sameText(existing.Degree, item.Degree)
		}) {
			result.Duplicates++
			continue
		}
		if _, err := a.SaveEducation(item); err != nil {
			return result, fmt.Errorf("save education at %s: %w", item.Institution, err)
		}
		result.Educations++
	}
	return result, nil
}

// mergeImportedProfile fills empty profile fields from the import and adds
// skills the profile does not have yet.
func mergeImportedProfile(current, imported domain.ProfileInput) (domain.ProfileInput, int) {
	fill := func(field *string, value string) {
		if strings.TrimSpace(*field) == "" {
			*field = strings.TrimSpace(value)
		}
	}
	fill(&current.Name, imported.Name)
	fill(&current.Headline, imported.Headline)
	fill(&current.Email, imported.Email)
	fill(&current.Phone, imported.Phone)
	fill(&current.Location, imported.Location)
	fill(&current.Website, imported.Website)
	fill(&current.GitHubUsername, imported.GitHubUsername)
	fill(&current.LinkedInURL, imported.LinkedInURL)
	added := 0
	for _, skill := range imported.Skills {
		if !containsKey(current.Skills, func(existing string) bool { return sameText(existing, skill) }) {
			current.Skills = append(current.Skills, skill)
			added++
		}
	}
	return current, added
}

func containsKey[T any](items []T, match func(T) bool) bool {
	for _, item := range items {
		if match(item) {
			return true
		}
	}
	return false
}

func sameText(left, right string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(left), " "), strings.Join(strings.Fields(right), " "))
}
