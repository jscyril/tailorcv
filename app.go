package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jscyril/tailorcv/internal/ai"
	backupfile "github.com/jscyril/tailorcv/internal/backup"
	"github.com/jscyril/tailorcv/internal/credentials"
	"github.com/jscyril/tailorcv/internal/domain"
	githubclient "github.com/jscyril/tailorcv/internal/github"
	"github.com/jscyril/tailorcv/internal/resume"
	"github.com/jscyril/tailorcv/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-facing application service. Domain and persistence details
// stay behind this deliberately small API.
type App struct {
	ctx               context.Context
	store             *storage.Store
	github            *githubclient.Client
	compiler          *resume.Compiler
	initErr           error
	compileMu         sync.RWMutex
	lastPDF           []byte
	aiMu              sync.Mutex
	aiCancel          context.CancelFunc
	aiGeneration      uint64
	aiProviderFactory func(string, string) (ai.Provider, error)
	credentials       credentials.Store
}

func NewApp() *App {
	return &App{github: githubclient.NewClient(nil), compiler: resume.NewCompiler(""), credentials: credentials.OSStore{}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.store, a.initErr = storage.OpenDefault()
	if a.initErr != nil {
		a.initErr = fmt.Errorf("open local profile store: %w", a.initErr)
	}
}

func (a *App) shutdown(_ context.Context) {
	if a.store != nil {
		if err := a.store.Close(); err != nil {
			fmt.Printf("close local profile store: %v\n", err)
		}
	}
}

// GetProfile returns the single local career profile. A new installation
// receives an empty profile rather than a not-found error.
func (a *App) GetProfile() (domain.Profile, error) {
	if err := a.ready(); err != nil {
		return domain.Profile{}, err
	}
	return a.store.GetProfile(a.appContext())
}

// SaveProfile validates, normalizes, and atomically persists the career
// profile and its skills.
func (a *App) SaveProfile(input domain.ProfileInput) (domain.Profile, error) {
	if err := a.ready(); err != nil {
		return domain.Profile{}, err
	}
	profile, err := input.Validate()
	if err != nil {
		return domain.Profile{}, err
	}
	return a.store.SaveProfile(a.appContext(), profile)
}

// ListExperiences returns manually entered and imported career evidence in its
// user-defined order.
func (a *App) ListExperiences() ([]domain.Experience, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.ListExperiences(a.appContext())
}

// SaveExperience validates an experience and atomically replaces its ordered
// evidence bullets while preserving their stable identifiers.
func (a *App) SaveExperience(input domain.ExperienceInput) (domain.Experience, error) {
	if err := a.ready(); err != nil {
		return domain.Experience{}, err
	}
	experience, err := input.Validate()
	if err != nil {
		return domain.Experience{}, err
	}
	return a.store.SaveExperience(a.appContext(), experience)
}

// DeleteExperience removes an experience and its evidence bullets.
func (a *App) DeleteExperience(id string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteExperience(a.appContext(), id)
}

// ListEducations returns education records in their resume order.
func (a *App) ListEducations() ([]domain.Education, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.ListEducations(a.appContext())
}

// SaveEducation validates and persists an education record while preserving
// its stable identifier and position.
func (a *App) SaveEducation(input domain.EducationInput) (domain.Education, error) {
	if err := a.ready(); err != nil {
		return domain.Education{}, err
	}
	education, err := input.Validate()
	if err != nil {
		return domain.Education{}, err
	}
	return a.store.SaveEducation(a.appContext(), education)
}

// DeleteEducation removes an education record.
func (a *App) DeleteEducation(id string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteEducation(a.appContext(), id)
}

// ListProjects returns manual and imported projects with their review state,
// skills, and ordered evidence.
func (a *App) ListProjects() ([]domain.Project, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.ListProjects(a.appContext())
}

// SaveProject validates and atomically persists a project and its evidence.
func (a *App) SaveProject(input domain.ProjectInput) (domain.Project, error) {
	if err := a.ready(); err != nil {
		return domain.Project{}, err
	}
	project, err := input.Validate()
	if err != nil {
		return domain.Project{}, err
	}
	return a.store.SaveProject(a.appContext(), project)
}

// DeleteProject removes a project and its associated skills and evidence.
func (a *App) DeleteProject(id string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteProject(a.appContext(), id)
}

// ExportProfileBackup writes a versioned snapshot of all currently supported
// career-profile data to a user-selected JSON file.
func (a *App) ExportProfileBackup() (domain.BackupResult, error) {
	if err := a.ready(); err != nil {
		return domain.BackupResult{}, err
	}
	backup, err := a.store.CreateProfileBackup(a.appContext())
	if err != nil {
		return domain.BackupResult{}, err
	}
	path, err := runtime.SaveFileDialog(a.appContext(), runtime.SaveDialogOptions{
		Title:           "Export TailorCV profile backup",
		DefaultFilename: "tailorcv-profile-backup.json",
		Filters:         []runtime.FileFilter{{DisplayName: "JSON backup (*.json)", Pattern: "*.json"}},
	})
	if err != nil {
		return domain.BackupResult{}, fmt.Errorf("choose backup destination: %w", err)
	}
	if path == "" {
		return domain.BackupResult{Cancelled: true}, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		path += ".json"
	}
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return domain.BackupResult{}, fmt.Errorf("encode backup: %w", err)
	}
	data = append(data, '\n')
	if err := backupfile.Write(path, data); err != nil {
		return domain.BackupResult{}, err
	}
	return backup.Result(path), nil
}

// ImportProfileBackup validates a user-selected backup completely before
// replacing profile data in one database transaction.
func (a *App) ImportProfileBackup() (domain.BackupResult, error) {
	if err := a.ready(); err != nil {
		return domain.BackupResult{}, err
	}
	path, err := runtime.OpenFileDialog(a.appContext(), runtime.OpenDialogOptions{
		Title:   "Restore TailorCV profile backup",
		Filters: []runtime.FileFilter{{DisplayName: "JSON backup (*.json)", Pattern: "*.json"}},
	})
	if err != nil {
		return domain.BackupResult{}, fmt.Errorf("choose backup file: %w", err)
	}
	if path == "" {
		return domain.BackupResult{Cancelled: true}, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		return domain.BackupResult{}, fmt.Errorf("backup file must use the .json extension")
	}
	data, err := backupfile.Read(path)
	if err != nil {
		return domain.BackupResult{}, err
	}
	backup, err := domain.DecodeProfileBackup(data)
	if err != nil {
		return domain.BackupResult{}, err
	}
	if err := a.store.ReplaceProfileFromBackup(a.appContext(), backup); err != nil {
		return domain.BackupResult{}, err
	}
	return backup.Result(path), nil
}

// ImportGitHubProjects fetches public repositories owned by the configured
// GitHub user. New repositories require review before becoming resume eligible;
// refreshes preserve user-entered evidence and review decisions.
func (a *App) ImportGitHubProjects() (domain.GitHubImportResult, error) {
	if err := a.ready(); err != nil {
		return domain.GitHubImportResult{}, err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return domain.GitHubImportResult{}, err
	}
	if profile.GitHubUsername == "" {
		return domain.GitHubImportResult{}, fmt.Errorf("add a GitHub username to your profile before syncing repositories")
	}
	repositories, err := a.github.ListPublicRepositories(a.appContext(), profile.GitHubUsername)
	if err != nil {
		return domain.GitHubImportResult{}, err
	}
	existingProjects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return domain.GitHubImportResult{}, err
	}
	existingByRepository := make(map[string]domain.Project, len(existingProjects))
	existingByRepositoryID := make(map[int64]domain.Project, len(existingProjects))
	for _, project := range existingProjects {
		if project.RepositoryID > 0 {
			existingByRepositoryID[project.RepositoryID] = project
		}
		if project.RepositoryURL != "" {
			existingByRepository[strings.ToLower(project.RepositoryURL)] = project
		}
	}

	result := domain.GitHubImportResult{Fetched: len(repositories)}
	for _, repository := range repositories {
		if repository.Fork || repository.Archived {
			result.Skipped++
			continue
		}
		if !repository.LanguagesComplete {
			result.LanguageFallbacks++
		}
		var existing *domain.Project
		project, found := existingByRepositoryID[repository.ID]
		if !found {
			project, found = existingByRepository[strings.ToLower(repository.HTMLURL)]
		}
		if found {
			if project.Provenance != domain.ProvenanceGitHub {
				result.Skipped++
				continue
			}
			existing = &project
		}
		project, err := repository.Project(existing)
		if err != nil {
			return domain.GitHubImportResult{}, fmt.Errorf("prepare GitHub project %q: %w", repository.Name, err)
		}
		if _, err := a.store.SaveProject(a.appContext(), project); err != nil {
			return domain.GitHubImportResult{}, fmt.Errorf("save GitHub project %q: %w", repository.Name, err)
		}
		if existing == nil {
			result.Imported++
		} else {
			result.Updated++
		}
		if !repository.ReadmeComplete {
			result.ReadmeFallbacks++
		}
	}
	return result, nil
}

// AnalyzeJobDescription performs the deterministic first-stage comparison.
// Later embedding and LLM providers will enrich this result, not replace it.
func (a *App) AnalyzeJobDescription(input domain.JobAnalysisInput) (domain.JobAnalysis, error) {
	if err := a.ready(); err != nil {
		return domain.JobAnalysis{}, err
	}
	job, err := input.JobInput().Validate()
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	experiences, err := a.store.ListExperiences(a.appContext())
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	projects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	analysis, err := domain.AnalyzeCareerEvidence(input, profile.Skills, experiences, projects)
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	searchHits, err := a.store.SearchEvidence(a.appContext(), analysis.SearchTerms, 50)
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	analysis, err = domain.AnalyzeCareerEvidenceWithSearch(input, profile.Skills, experiences, projects, searchHits)
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	saved, err := a.store.SaveJob(a.appContext(), job)
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	analysis.Job = saved
	return analysis, nil
}

// ListJobs returns saved opportunities with the most recently analyzed first.
func (a *App) ListJobs() ([]domain.Job, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.ListJobs(a.appContext())
}

func (a *App) DeleteJob(id string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteJob(a.appContext(), id)
}

// ListApplications returns local application records with immutable resume
// versions ordered newest first.
func (a *App) ListApplications() ([]domain.Application, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.store.ListApplications(a.appContext())
}

// UpdateApplicationStatus moves a saved application through its local
// lifecycle without changing any immutable resume version.
func (a *App) UpdateApplicationStatus(input domain.UpdateApplicationStatusInput) (domain.Application, error) {
	if err := a.ready(); err != nil {
		return domain.Application{}, err
	}
	return a.store.UpdateApplicationStatus(a.appContext(), input)
}

// CreateResumeVersion renders only explicitly selected evidence and appends an
// immutable source snapshot to the application associated with the saved job.
func (a *App) CreateResumeVersion(input domain.CreateResumeVersionInput) (domain.ApplicationResumeResult, error) {
	if err := a.ready(); err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	validated, err := input.Validate()
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	job, err := a.store.GetJob(a.appContext(), validated.JobID)
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
	selectedExperiences, selectedProjects, err := selectResumeEvidence(validated.SelectedFactIDs, experiences, projects)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	source := resume.Render(template.Source, resume.Data{Profile: profile, Experiences: selectedExperiences, Projects: selectedProjects, Educations: educations})
	ranking, err := a.rankSelectedEvidence(job, profile, experiences, projects, validated.SelectedFactIDs)
	if err != nil {
		return domain.ApplicationResumeResult{}, err
	}
	return a.store.CreateResumeVersion(a.appContext(), job.ID, validated.SelectedFactIDs, template.ID, job.Description, source, ranking)
}

func (a *App) rankSelectedEvidence(job domain.Job, profile domain.Profile, experiences []domain.Experience, projects []domain.Project, selectedFactIDs []string) ([]domain.EvidenceMatch, error) {
	input := domain.JobAnalysisInput{ID: job.ID, Company: job.Company, Role: job.Role, Description: job.Description}
	analysis, err := domain.AnalyzeCareerEvidence(input, profile.Skills, experiences, projects)
	if err != nil {
		return nil, err
	}
	searchHits, err := a.store.SearchEvidence(a.appContext(), analysis.SearchTerms, 50)
	if err != nil {
		return nil, err
	}
	analysis, err = domain.AnalyzeCareerEvidenceWithSearch(input, profile.Skills, experiences, projects, searchHits)
	if err != nil {
		return nil, err
	}
	selected := make(map[string]struct{}, len(selectedFactIDs))
	for _, id := range selectedFactIDs {
		selected[id] = struct{}{}
	}
	ranking := make([]domain.EvidenceMatch, 0, len(selectedFactIDs))
	for _, evidence := range analysis.RankedEvidence {
		if _, included := selected[evidence.FactID]; included {
			ranking = append(ranking, evidence)
		}
	}
	return ranking, nil
}

// SaveResumeVersionEdit appends the current edited source as a new immutable
// snapshot, copying the base version's job, evidence, template, and ranking data.
func (a *App) SaveResumeVersionEdit(input domain.SaveResumeVersionEditInput) (domain.ResumeVersion, error) {
	if err := a.ready(); err != nil {
		return domain.ResumeVersion{}, err
	}
	return a.store.CreateEditedResumeVersion(a.appContext(), input)
}

func selectResumeEvidence(selectedFactIDs []string, experiences []domain.Experience, projects []domain.Project) ([]domain.Experience, []domain.Project, error) {
	selected := make(map[string]struct{}, len(selectedFactIDs))
	for _, id := range selectedFactIDs {
		selected[id] = struct{}{}
	}
	found := make(map[string]struct{}, len(selectedFactIDs))
	filteredExperiences := make([]domain.Experience, 0)
	for _, experience := range experiences {
		bullets := make([]domain.EvidenceBullet, 0)
		for _, bullet := range experience.Bullets {
			if _, included := selected[bullet.ID]; included {
				bullets = append(bullets, bullet)
				found[bullet.ID] = struct{}{}
			}
		}
		if len(bullets) > 0 {
			experience.Bullets = bullets
			filteredExperiences = append(filteredExperiences, experience)
		}
	}
	filteredProjects := make([]domain.Project, 0)
	for _, project := range projects {
		_, wholeProject := selected[project.ID]
		bullets := make([]domain.EvidenceBullet, 0)
		for _, bullet := range project.Bullets {
			if _, included := selected[bullet.ID]; included {
				if !project.ResumeEligible {
					return nil, nil, fmt.Errorf("project %q must be reviewed before its evidence can be selected", project.Name)
				}
				bullets = append(bullets, bullet)
				found[bullet.ID] = struct{}{}
			}
		}
		if wholeProject {
			if !project.ResumeEligible {
				return nil, nil, fmt.Errorf("project %q must be reviewed before it can be selected", project.Name)
			}
			found[project.ID] = struct{}{}
		}
		if wholeProject || len(bullets) > 0 {
			project.Bullets = bullets
			filteredProjects = append(filteredProjects, project)
		}
	}
	for _, id := range selectedFactIDs {
		if _, exists := found[id]; !exists {
			return nil, nil, fmt.Errorf("selected evidence %q was not found in the current profile", id)
		}
	}
	return filteredExperiences, filteredProjects, nil
}

// ListResumeTemplates returns read-only built-ins followed by user-owned
// templates imported or created on this device.
func (a *App) ListResumeTemplates() ([]domain.ResumeTemplate, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	custom, err := a.store.ListTemplates(a.appContext())
	if err != nil {
		return nil, err
	}
	return append(resume.BuiltinTemplates(), custom...), nil
}

func (a *App) GetSelectedResumeTemplateID() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	id, err := a.store.SelectedTemplateID(a.appContext())
	if err != nil {
		return "", err
	}
	if id == "" {
		return resume.DefaultTemplateID, nil
	}
	if _, err := a.resumeTemplate(id); err != nil {
		return resume.DefaultTemplateID, nil
	}
	return id, nil
}

func (a *App) SelectResumeTemplate(id string) (domain.ResumeTemplate, error) {
	if err := a.ready(); err != nil {
		return domain.ResumeTemplate{}, err
	}
	template, err := a.resumeTemplate(id)
	if err != nil {
		return domain.ResumeTemplate{}, err
	}
	if err := a.store.SetSelectedTemplateID(a.appContext(), template.ID); err != nil {
		return domain.ResumeTemplate{}, err
	}
	return template, nil
}

// ImportResumeTemplate copies a complete, single-file .tex document into the
// local template library. An empty ID means the native dialog was cancelled.
func (a *App) ImportResumeTemplate() (domain.ResumeTemplate, error) {
	if err := a.ready(); err != nil {
		return domain.ResumeTemplate{}, err
	}
	path, err := runtime.OpenFileDialog(a.appContext(), runtime.OpenDialogOptions{
		Title:   "Import LaTeX resume template",
		Filters: []runtime.FileFilter{{DisplayName: "LaTeX template (*.tex)", Pattern: "*.tex"}},
	})
	if err != nil {
		return domain.ResumeTemplate{}, fmt.Errorf("choose LaTeX template: %w", err)
	}
	if path == "" {
		return domain.ResumeTemplate{}, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".tex") {
		return domain.ResumeTemplate{}, fmt.Errorf("template file must use the .tex extension")
	}
	info, err := os.Stat(path)
	if err != nil {
		return domain.ResumeTemplate{}, fmt.Errorf("inspect LaTeX template: %w", err)
	}
	if info.Size() > domain.MaxTemplateSourceBytes {
		return domain.ResumeTemplate{}, fmt.Errorf("template source exceeds the 1 MiB size limit")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.ResumeTemplate{}, fmt.Errorf("read LaTeX template: %w", err)
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	template, err := (domain.ResumeTemplateInput{Name: name, Description: "Imported from " + filepath.Base(path), Source: string(data)}).Validate()
	if err != nil {
		return domain.ResumeTemplate{}, err
	}
	return a.store.SaveTemplate(a.appContext(), template)
}

// SaveResumeTemplate creates or updates a user-owned template. Built-ins can
// only be customized by saving them as a new copy with an empty ID.
func (a *App) SaveResumeTemplate(input domain.ResumeTemplateInput) (domain.ResumeTemplate, error) {
	if err := a.ready(); err != nil {
		return domain.ResumeTemplate{}, err
	}
	if _, builtIn := resume.FindBuiltinTemplate(strings.TrimSpace(input.ID)); builtIn {
		return domain.ResumeTemplate{}, fmt.Errorf("built-in templates are read-only; save an editable copy instead")
	}
	template, err := input.Validate()
	if err != nil {
		return domain.ResumeTemplate{}, err
	}
	return a.store.SaveTemplate(a.appContext(), template)
}

func (a *App) DeleteResumeTemplate(id string) error {
	if err := a.ready(); err != nil {
		return err
	}
	if _, builtIn := resume.FindBuiltinTemplate(id); builtIn {
		return fmt.Errorf("built-in templates cannot be deleted")
	}
	if err := a.store.DeleteTemplate(a.appContext(), id); err != nil {
		return err
	}
	selected, err := a.store.SelectedTemplateID(a.appContext())
	if err == nil && selected == id {
		return a.store.SetSelectedTemplateID(a.appContext(), resume.DefaultTemplateID)
	}
	return err
}

// RenderResumeTemplate replaces TailorCV markers with escaped, locally stored
// profile data. Marker-free imported documents remain unchanged.
func (a *App) RenderResumeTemplate(id string, selectedProjectIDs []string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	template, err := a.resumeTemplate(id)
	if err != nil {
		return "", err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return "", err
	}
	experiences, err := a.store.ListExperiences(a.appContext())
	if err != nil {
		return "", err
	}
	projects, err := a.store.ListProjects(a.appContext())
	if err != nil {
		return "", err
	}
	educations, err := a.store.ListEducations(a.appContext())
	if err != nil {
		return "", err
	}
	selected := make(map[string]struct{}, len(selectedProjectIDs))
	for _, projectID := range selectedProjectIDs {
		selected[projectID] = struct{}{}
	}
	filteredProjects := make([]domain.Project, 0, len(projects))
	for _, project := range projects {
		if _, included := selected[project.ID]; included && project.ResumeEligible {
			filteredProjects = append(filteredProjects, project)
		}
	}
	return resume.Render(template.Source, resume.Data{Profile: profile, Experiences: experiences, Projects: filteredProjects, Educations: educations}), nil
}

// CompileLatex runs Tectonic without a shell in an isolated temporary folder.
func (a *App) CompileLatex(source string) (domain.CompileResult, error) {
	if err := a.ready(); err != nil {
		return domain.CompileResult{}, err
	}
	result, pdf, err := a.compiler.Compile(a.appContext(), source)
	a.compileMu.Lock()
	a.lastPDF = append(a.lastPDF[:0], pdf...)
	a.compileMu.Unlock()
	if err != nil {
		return domain.CompileResult{}, err
	}
	return result, nil
}

// CompileResumeVersion compiles an immutable saved source and records its
// derived diagnostics and private PDF artifact without changing the snapshot.
func (a *App) CompileResumeVersion(versionID string) (domain.CompileResult, error) {
	if err := a.ready(); err != nil {
		return domain.CompileResult{}, err
	}
	version, err := a.store.GetResumeVersion(a.appContext(), versionID)
	if err != nil {
		return domain.CompileResult{}, err
	}
	result, pdf, err := a.compiler.Compile(a.appContext(), version.LatexSource)
	if err != nil {
		return domain.CompileResult{}, err
	}
	pdfPath := ""
	if result.Success {
		pdfPath, err = a.resumeArtifactPath(version.ID)
		if err != nil {
			return domain.CompileResult{}, err
		}
		if err := writePrivateFile(pdfPath, pdf); err != nil {
			return domain.CompileResult{}, fmt.Errorf("save compiled resume artifact: %w", err)
		}
	}
	if err := a.store.RecordResumeCompilation(a.appContext(), version.ID, pdfPath, result); err != nil {
		return domain.CompileResult{}, err
	}
	a.compileMu.Lock()
	a.lastPDF = append(a.lastPDF[:0], pdf...)
	a.compileMu.Unlock()
	return result, nil
}

func (a *App) resumeArtifactPath(versionID string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve resume artifact directory: %w", err)
	}
	directory := filepath.Join(configDir, "tailorcv", "artifacts")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("create resume artifact directory: %w", err)
	}
	return filepath.Join(directory, versionID+".pdf"), nil
}

func (a *App) ExportCompiledPDF() (domain.FileResult, error) {
	if err := a.ready(); err != nil {
		return domain.FileResult{}, err
	}
	a.compileMu.RLock()
	pdf := append([]byte(nil), a.lastPDF...)
	a.compileMu.RUnlock()
	if len(pdf) == 0 {
		return domain.FileResult{}, fmt.Errorf("compile the current LaTeX source before exporting a PDF")
	}
	path, err := runtime.SaveFileDialog(a.appContext(), runtime.SaveDialogOptions{
		Title:           "Export compiled resume",
		DefaultFilename: "resume.pdf",
		Filters:         []runtime.FileFilter{{DisplayName: "PDF document (*.pdf)", Pattern: "*.pdf"}},
	})
	if err != nil {
		return domain.FileResult{}, fmt.Errorf("choose PDF destination: %w", err)
	}
	if path == "" {
		return domain.FileResult{Cancelled: true}, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		path += ".pdf"
	}
	if err := writePrivateFile(path, pdf); err != nil {
		return domain.FileResult{}, fmt.Errorf("export PDF: %w", err)
	}
	return domain.FileResult{Path: path}, nil
}

func (a *App) ExportLatexSource(source string) (domain.FileResult, error) {
	if err := a.ready(); err != nil {
		return domain.FileResult{}, err
	}
	if len(source) == 0 || len(source) > domain.MaxTemplateSourceBytes {
		return domain.FileResult{}, fmt.Errorf("LaTeX source is empty or exceeds the 1 MiB size limit")
	}
	path, err := runtime.SaveFileDialog(a.appContext(), runtime.SaveDialogOptions{
		Title:           "Export LaTeX source",
		DefaultFilename: "resume.tex",
		Filters:         []runtime.FileFilter{{DisplayName: "LaTeX source (*.tex)", Pattern: "*.tex"}},
	})
	if err != nil {
		return domain.FileResult{}, fmt.Errorf("choose LaTeX destination: %w", err)
	}
	if path == "" {
		return domain.FileResult{Cancelled: true}, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".tex") {
		path += ".tex"
	}
	if err := writePrivateFile(path, []byte(source)); err != nil {
		return domain.FileResult{}, fmt.Errorf("export LaTeX source: %w", err)
	}
	return domain.FileResult{Path: path}, nil
}

func (a *App) resumeTemplate(id string) (domain.ResumeTemplate, error) {
	if template, found := resume.FindBuiltinTemplate(id); found {
		return template, nil
	}
	templates, err := a.store.ListTemplates(a.appContext())
	if err != nil {
		return domain.ResumeTemplate{}, err
	}
	for _, template := range templates {
		if template.ID == id {
			return template, nil
		}
	}
	return domain.ResumeTemplate{}, fmt.Errorf("resume template was not found")
}

func writePrivateFile(path string, data []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".tailorcv-export-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func (a *App) ready() error {
	if a.initErr != nil {
		return a.initErr
	}
	if a.store == nil {
		return fmt.Errorf("local profile store is not ready")
	}
	return nil
}

func (a *App) appContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
