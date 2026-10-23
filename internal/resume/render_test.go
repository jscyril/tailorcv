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
	for _, expected := range []string{`Ada \& Co\_100\%`, `\section{Experience}`, `\$5 \& shipped \#1`, `Jan 2024 -- Present`} {
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

func TestRenderCoverLetterEscapesTextAndKeepsStructure(t *testing.T) {
	letter := "Dear Acme hiring team,\n\nIntro with 100% & $5 \\input{x}.\nSecond line\n\nWork:\n- Built Go #1\n- Ran R&D\n\nSincerely,\nAda"
	result := RenderCoverLetter(domain.Profile{Name: "Ada"}, letter)
	for _, expected := range []string{`Intro with 100\% \& \$5 \textbackslash{}input\{x\}.\\`, "Second line", "Work:\n\\begin{itemize}\n  \\item Built Go \\#1\n  \\item Ran R\\&D\n\\end{itemize}", "Sincerely,\\\\\nAda", `\textbf{Ada}`} {
		if !strings.Contains(result, expected) {
			t.Errorf("RenderCoverLetter() missing %q in:\n%s", expected, result)
		}
	}
	if strings.Contains(result, `\input{x}`) {
		t.Fatal("letter text must not become a LaTeX command")
	}
}

func TestCommentOutPDFTeXOnly(t *testing.T) {
	if reason := unsafeSourceReason(mustCommentOut("\\input{glyphtounicode}"), "Windows"); reason != "" {
		t.Fatalf("the replacement note must pass the Windows scan: %s", reason)
	}
	source := "\\usepackage{tabularx}\n\\input{glyphtounicode}\n\\pdfgentounicode=1\n\\input{other}\n"
	got, count := CommentOutPDFTeXOnly(source)
	note := "% TailorCV removed a pdfTeX-only line (glyphtounicode setup); Tectonic does not need it."
	if count != 2 || got != "\\usepackage{tabularx}\n"+note+"\n"+note+"\n\\input{other}\n" {
		t.Fatalf("CommentOutPDFTeXOnly() = %q, %d", got, count)
	}
}

func TestCommentOutPDFTeXOnlyWithWindowsLineEndings(t *testing.T) {
	got, count := CommentOutPDFTeXOnly("\\input{glyphtounicode}\r\n\\pdfgentounicode=1\r\n")
	if count != 2 || strings.Contains(got, `\input`) || strings.Count(got, "\r\n") != 2 {
		t.Fatalf("CommentOutPDFTeXOnly(CRLF) = %q, %d", got, count)
	}
}

func mustCommentOut(source string) string {
	result, _ := CommentOutPDFTeXOnly(source)
	return result
}

func TestContactLinksAndOptionalHeadline(t *testing.T) {
	result := Render(`{{TAILORCV_HEADLINE_LINE}}|{{TAILORCV_CONTACT_LINKS}}`, Data{Profile: domain.Profile{Email: "a_b@example.com", Phone: "+1 555 0100", Location: "Remote", Website: "https://example.com/100%", GitHubUsername: "ada"}})
	want := `|\href{mailto:a\_b@example.com}{\underline{a\_b@example.com}} $|$ +1 555 0100 $|$ Remote $|$ \href{https://example.com/100\%}{\underline{example.com/100\%}} $|$ \href{https://github.com/ada}{\underline{github.com/ada}}`
	if result != want {
		t.Fatalf("Render() = %q\nwant      %q", result, want)
	}
	if got := Render(`{{TAILORCV_HEADLINE_LINE}}`, Data{Profile: domain.Profile{Headline: "Engineer"}}); got != `Engineer\\[2pt]` {
		t.Fatalf("headline line = %q", got)
	}
}
