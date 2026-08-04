package domain

import (
	"fmt"
	"strings"
)

const MaxTemplateSourceBytes = 1 << 20

type ResumeTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	BuiltIn     bool   `json:"builtIn"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ResumeTemplateInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

func (input ResumeTemplateInput) Validate() (ResumeTemplate, error) {
	template := ResumeTemplate{
		ID:          strings.TrimSpace(input.ID),
		Name:        strings.Join(strings.Fields(input.Name), " "),
		Description: strings.TrimSpace(input.Description),
		Source:      strings.TrimSpace(input.Source),
	}
	if template.Name == "" {
		return ResumeTemplate{}, fmt.Errorf("template name is required")
	}
	if len(template.Name) > 120 {
		return ResumeTemplate{}, fmt.Errorf("template name must be 120 characters or fewer")
	}
	if len(template.Description) > 500 {
		return ResumeTemplate{}, fmt.Errorf("template description must be 500 characters or fewer")
	}
	if template.Source == "" {
		return ResumeTemplate{}, fmt.Errorf("template source is required")
	}
	if len(template.Source) > MaxTemplateSourceBytes {
		return ResumeTemplate{}, fmt.Errorf("template source exceeds the 1 MiB size limit")
	}
	if !strings.Contains(template.Source, `\documentclass`) || !strings.Contains(template.Source, `\begin{document}`) || !strings.Contains(template.Source, `\end{document}`) {
		return ResumeTemplate{}, fmt.Errorf("template must be a complete LaTeX document")
	}
	return template, nil
}

type CompileResult struct {
	Success     bool                `json:"success"`
	PDFBase64   string              `json:"pdfBase64"`
	Engine      string              `json:"engine"`
	DurationMS  int64               `json:"durationMs"`
	Log         string              `json:"log"`
	Diagnostics []CompileDiagnostic `json:"diagnostics"`
}

type CompileDiagnostic struct {
	Line     int    `json:"line"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type FileResult struct {
	Path      string `json:"path"`
	Cancelled bool   `json:"cancelled"`
}
