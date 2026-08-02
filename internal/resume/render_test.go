package resume

import (
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestRenderEscapesProfileAndEvidence(t *testing.T) {
	source := `\documentclass{article}\begin{document}{{TAILORCV_NAME}}{{TAILORCV_EXPERIENCE_SECTION}}\end{document}`
	result := Render(source, Data{
		Profile: domain.Profile{Name: `Ada & Co_100%`},
		Experiences: []domain.Experience{{
			Company: "Analytical Engines", Title: "Engineer", StartDate: "2024-01", Current: true,
			Bullets: []domain.EvidenceBullet{{Text: `Reduced cost by $5 & shipped #1`}},
		}},
	})
	for _, expected := range []string{`Ada \& Co\_100\%`, `\section{Experience}`, `\$5 \& shipped \#1`, `2024-01 -- Present`} {
		if !strings.Contains(result, expected) {
			t.Fatalf("Render() missing %q in:\n%s", expected, result)
		}
	}
}

func TestBuiltinTemplatesAreComplete(t *testing.T) {
	for _, template := range BuiltinTemplates() {
		if _, err := (domain.ResumeTemplateInput{Name: template.Name, Source: template.Source}).Validate(); err != nil {
			t.Fatalf("built-in %q is invalid: %v", template.Name, err)
		}
	}
}
