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

func TestRenderIncludesAdditionalProfileEvidence(t *testing.T) {
	source := `{{TAILORCV_CONTACT}}{{TAILORCV_CERTIFICATIONS_SECTION}}{{TAILORCV_ACHIEVEMENTS_SECTION}}`
	result := Render(source, Data{Profile: domain.Profile{ContactLinks: []domain.ContactLink{{Label: "Portfolio", URL: "https://example.com/work"}}}, Certifications: []domain.Certification{{Name: "Cloud Professional", Issuer: "Example Institute"}}, Achievements: []domain.Achievement{{Title: "Engineering Award", Description: "Recognized for a fictional release system."}}})
	for _, expected := range []string{"https://example.com/work", `\section{Certifications}`, "Cloud Professional", `\section{Achievements}`, "Engineering Award"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("Render() missing %q in %s", expected, result)
		}
	}
}

func TestRenderTreatsJobAndModelDerivedEvidenceAsLiteralText(t *testing.T) {
	payload := `\input{/private/secret}\immediate\write18{touch /tmp/tailorcv-pwned}%_${}`
	source := `\documentclass{article}\begin{document}{{TAILORCV_SUMMARY_SECTION}}{{TAILORCV_EXPERIENCE_SECTION}}\end{document}`
	result := Render(source, Data{
		Profile: domain.Profile{Summary: payload},
		Experiences: []domain.Experience{{
			Company: "Example", Title: "Engineer", StartDate: "2026-01",
			Bullets: []domain.EvidenceBullet{{Text: payload}},
		}},
	})
	for _, unsafe := range []string{`\input{/private/secret}`, `\immediate\write18`} {
		if strings.Contains(result, unsafe) {
			t.Fatalf("Render() preserved executable control %q in:\n%s", unsafe, result)
		}
	}
	if count := strings.Count(result, `\textbackslash{}`); count < 4 {
		t.Fatalf("Render() did not escape injected control sequences in:\n%s", result)
	}
}
