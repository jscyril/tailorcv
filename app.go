package main

import (
	"context"
	"fmt"

	"github.com/jscyril/tailorcv/internal/domain"
	"github.com/jscyril/tailorcv/internal/storage"
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
