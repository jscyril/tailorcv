package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jscyril/tailorcv/internal/domain"
)

const latexDraftSetting = "latex_draft"

// LatexDraft returns the saved unsaved-edits draft, or an empty draft.
func (s *Store) LatexDraft(ctx context.Context) (domain.LatexDraft, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM app_settings WHERE key = ?`, latexDraftSetting).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.LatexDraft{}, nil
	}
	if err != nil {
		return domain.LatexDraft{}, fmt.Errorf("read LaTeX draft: %w", err)
	}
	var draft domain.LatexDraft
	if err := json.Unmarshal([]byte(value), &draft); err != nil {
		// A damaged draft is not worth blocking startup over; drop it.
		return domain.LatexDraft{}, s.ClearLatexDraft(ctx)
	}
	return draft, nil
}

func (s *Store) SaveLatexDraft(ctx context.Context, draft domain.LatexDraft) error {
	value, err := json.Marshal(draft)
	if err != nil {
		return fmt.Errorf("encode LaTeX draft: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO app_settings(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, latexDraftSetting, string(value)); err != nil {
		return fmt.Errorf("save LaTeX draft: %w", err)
	}
	return nil
}

func (s *Store) ClearLatexDraft(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM app_settings WHERE key = ?`, latexDraftSetting); err != nil {
		return fmt.Errorf("clear LaTeX draft: %w", err)
	}
	return nil
}
