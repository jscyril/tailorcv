package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	backupfile "github.com/jscyril/tailorcv/internal/backup"
	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-facing application service. Domain and persistence details
// stay behind this deliberately small API.
type App struct {
	ctx     context.Context
	store   *storage.Store
	initErr error
}

func NewApp() *App {
	return &App{}
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
