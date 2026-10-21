package domain

import (
	"fmt"
	"strings"
	"time"
)

// Field limits match Google Tasks so a later sync never truncates.
const (
	maxTodoTitleLength = 1024
	maxTodoNotesLength = 8192
)

// Todo is a local task, optionally linked to a resume. CompletedAt is empty
// while the task is open.
type Todo struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Notes         string `json:"notes"`
	DueDate       string `json:"dueDate"`
	ApplicationID string `json:"applicationId"`
	CompletedAt   string `json:"completedAt"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type TodoInput struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Notes         string `json:"notes"`
	DueDate       string `json:"dueDate"`
	ApplicationID string `json:"applicationId"`
}

func (input TodoInput) Validate() (Todo, error) {
	todo := Todo{
		ID:            strings.TrimSpace(input.ID),
		Title:         strings.Join(strings.Fields(input.Title), " "),
		Notes:         strings.TrimSpace(input.Notes),
		DueDate:       strings.TrimSpace(input.DueDate),
		ApplicationID: strings.TrimSpace(input.ApplicationID),
	}
	if todo.Title == "" {
		return Todo{}, fmt.Errorf("to-do title is required")
	}
	if len(todo.Title) > maxTodoTitleLength {
		return Todo{}, fmt.Errorf("to-do title must be %d characters or fewer", maxTodoTitleLength)
	}
	if len(todo.Notes) > maxTodoNotesLength {
		return Todo{}, fmt.Errorf("to-do notes must be %d characters or fewer", maxTodoNotesLength)
	}
	if todo.DueDate != "" {
		if _, err := time.Parse("2006-01-02", todo.DueDate); err != nil {
			return Todo{}, fmt.Errorf("to-do due date must be a valid YYYY-MM-DD date")
		}
	}
	return todo, nil
}
