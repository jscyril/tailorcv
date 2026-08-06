package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jscyril/tailorcv/internal/domain"
)

const aiSettingsKey = "ai_settings"

func (s *Store) GetAISettings(ctx context.Context) (domain.AISettings, error) {
	settings := domain.DefaultAISettings()
	var provider, endpoint, ollamaModel, geminiModel string
	err := s.db.QueryRowContext(ctx, `
		SELECT
			json_extract(value, '$.provider'),
			json_extract(value, '$.ollamaEndpoint'),
			json_extract(value, '$.ollamaModel'),
			json_extract(value, '$.geminiModel')
		FROM app_settings WHERE key = ?
	`, aiSettingsKey).Scan(&provider, &endpoint, &ollamaModel, &geminiModel)
	if errors.Is(err, sql.ErrNoRows) {
		return settings, nil
	}
	if err != nil {
		return domain.AISettings{}, fmt.Errorf("read AI settings: %w", err)
	}
	return (domain.AISettings{Provider: provider, OllamaEndpoint: endpoint, OllamaModel: ollamaModel, GeminiModel: geminiModel}).Validate()
}

func (s *Store) SaveAISettings(ctx context.Context, input domain.AISettings) (domain.AISettings, error) {
	settings, err := input.Validate()
	if err != nil {
		return domain.AISettings{}, err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO app_settings(key, value) VALUES (?, json_object(
			'provider', ?, 'ollamaEndpoint', ?, 'ollamaModel', ?, 'geminiModel', ?
		))
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, aiSettingsKey, settings.Provider, settings.OllamaEndpoint, settings.OllamaModel, settings.GeminiModel)
	if err != nil {
		return domain.AISettings{}, fmt.Errorf("save AI settings: %w", err)
	}
	return settings, nil
}
