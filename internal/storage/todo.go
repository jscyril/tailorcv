package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jscyril/tailorcv/internal/domain"
)

const todosEnabledSetting = "todos_enabled"

const todoColumns = `id,title,notes,due_date,COALESCE(application_id,''),completed_at,created_at,updated_at`

func scanTodo(row interface{ Scan(...any) error }) (domain.Todo, error) {
	var item domain.Todo
	err := row.Scan(&item.ID, &item.Title, &item.Notes, &item.DueDate, &item.ApplicationID, &item.CompletedAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

// ListTodos returns open tasks first (soonest due, undated last), then
// completed tasks, most recently completed first.
func (s *Store) ListTodos(ctx context.Context) ([]domain.Todo, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+todoColumns+` FROM todos
		ORDER BY completed_at != '', due_date = '', due_date, completed_at DESC, created_at`)
	if err != nil {
		return nil, fmt.Errorf("list to-dos: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Todo, 0)
	for rows.Next() {
		item, err := scanTodo(rows)
		if err != nil {
			return nil, fmt.Errorf("scan to-do: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate to-dos: %w", err)
	}
	return items, nil
}

// SaveTodo creates or edits a task; completion is changed only by SetTodoDone.
func (s *Store) SaveTodo(ctx context.Context, item domain.Todo) (domain.Todo, error) {
	if item.ID == "" {
		item.ID = uuid.NewString()
	} else if _, err := uuid.Parse(item.ID); err != nil {
		return domain.Todo{}, fmt.Errorf("to-do ID is not valid")
	}
	if item.ApplicationID != "" {
		if err := s.db.QueryRowContext(ctx, `SELECT id FROM applications WHERE id = ?`, item.ApplicationID).Scan(new(string)); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.Todo{}, fmt.Errorf("linked resume was not found")
			}
			return domain.Todo{}, fmt.Errorf("read linked resume: %w", err)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `INSERT INTO todos(id,title,notes,due_date,application_id,created_at,updated_at) VALUES(?,?,?,?,NULLIF(?,''),?,?)
		ON CONFLICT(id) DO UPDATE SET title=excluded.title,notes=excluded.notes,due_date=excluded.due_date,application_id=excluded.application_id,updated_at=excluded.updated_at`,
		item.ID, item.Title, item.Notes, item.DueDate, item.ApplicationID, now, now)
	if err != nil {
		return domain.Todo{}, fmt.Errorf("write to-do: %w", err)
	}
	return s.getTodo(ctx, item.ID)
}

func (s *Store) SetTodoDone(ctx context.Context, id string, done bool) (domain.Todo, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	completedAt := ""
	if done {
		completedAt = now
	}
	result, err := s.db.ExecContext(ctx, `UPDATE todos SET completed_at = ?, updated_at = ? WHERE id = ?`, completedAt, now, id)
	if err != nil {
		return domain.Todo{}, fmt.Errorf("update to-do: %w", err)
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return domain.Todo{}, fmt.Errorf("to-do was not found")
	}
	return s.getTodo(ctx, id)
}

func (s *Store) DeleteTodo(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete to-do: %w", err)
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return fmt.Errorf("to-do was not found")
	}
	return nil
}

func (s *Store) getTodo(ctx context.Context, id string) (domain.Todo, error) {
	item, err := scanTodo(s.db.QueryRowContext(ctx, `SELECT `+todoColumns+` FROM todos WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Todo{}, fmt.Errorf("to-do was not found")
	}
	if err != nil {
		return domain.Todo{}, fmt.Errorf("read to-do: %w", err)
	}
	return item, nil
}

// TodosEnabled reports the Settings toggle; the list is on until turned off.
func (s *Store) TodosEnabled(ctx context.Context) (bool, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM app_settings WHERE key = ?`, todosEnabledSetting).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read to-do setting: %w", err)
	}
	return value != "false", nil
}

func (s *Store) SetTodosEnabled(ctx context.Context, enabled bool) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO app_settings(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, todosEnabledSetting, fmt.Sprint(enabled))
	if err != nil {
		return fmt.Errorf("save to-do setting: %w", err)
	}
	return nil
}
