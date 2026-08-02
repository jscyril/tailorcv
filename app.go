package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	backupfile "github.com/jscyril/tailorcv/internal/backup"
	"github.com/jscyril/tailorcv/internal/domain"
	githubclient "github.com/jscyril/tailorcv/internal/github"
	"github.com/jscyril/tailorcv/internal/resume"
	"github.com/jscyril/tailorcv/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-facing application service. Domain and persistence details
// stay behind this deliberately small API.
type App struct {
	ctx       context.Context
	store     *storage.Store
	github    *githubclient.Client
	compiler  *resume.Compiler
	initErr   error
	compileMu sync.RWMutex
	lastPDF   []byte
}

func NewApp() *App {
	return &App{github: githubclient.NewClient(nil), compiler: resume.NewCompiler("")}
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
	for _, project := range existingProjects {
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
		if project, found := existingByRepository[strings.ToLower(repository.HTMLURL)]; found {
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
	}
	return result, nil
}

// AnalyzeJobDescription performs the deterministic first-stage comparison.
// Later embedding and LLM providers will enrich this result, not replace it.
func (a *App) AnalyzeJobDescription(input domain.JobAnalysisInput) (domain.JobAnalysis, error) {
	if err := a.ready(); err != nil {
		return domain.JobAnalysis{}, err
	}
	profile, err := a.store.GetProfile(a.appContext())
	if err != nil {
		return domain.JobAnalysis{}, err
	}
	return domain.AnalyzeJobDescription(input, profile.Skills)
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
	if err != nil {
		return domain.CompileResult{}, err
	}
	a.compileMu.Lock()
	a.lastPDF = append(a.lastPDF[:0], pdf...)
	a.compileMu.Unlock()
	return result, nil
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
