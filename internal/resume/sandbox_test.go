package resume

import (
	"strings"
	"testing"
)

func TestUnsafeSourceReasonAllowsBuiltinsAndOrdinaryLatex(t *testing.T) {
	for _, template := range BuiltinTemplates() {
		if reason := unsafeSourceReason(Render(template.Source, Data{}), "Windows"); reason != "" {
			t.Fatalf("built-in %s refused: %s", template.ID, reason)
		}
	}
	ordinary := `\documentclass[letterpaper]{article}
\usepackage[margin=1in]{geometry}
\usepackage{hyperref,enumitem}
\begin{document}
\section*{Experience} e.g.\@ shipped things \\ \textbf{Go} \href{https://example.com}{site}
\begin{itemize}\item Reading list\end{itemize}
\end{document}`
	if reason := unsafeSourceReason(ordinary, "Windows"); reason != "" {
		t.Fatalf("ordinary LaTeX refused: %s", reason)
	}
}

func TestUnsafeSourceReasonBlocksFileAccessAndObfuscation(t *testing.T) {
	for name, source := range map[string]string{
		"input absolute":    `\input{/etc/passwd}`,
		"input macro alias": `\let\x\input \x{secret}`,
		"openout":           `\newwrite\f\immediate\openout\f=C:/Users/me/x.txt`,
		"openin":            `\openin1=secret`,
		"graphics":          `\includegraphics{/home/me/photo.png}`,
		"csname":            `\csname input\endcsname{/etc/passwd}`,
		"environment":       `\begin{input}{/etc/passwd}\end{input}`,
		"catcode":           "\\catcode`|=0 |input{x}",
		"superscript":       `^^5cinput{x}`,
		"expl3":             `\ExplSyntaxOn \use:c{input}`,
		"kernel UseName":    `\UseName{input}{x}`,
		"etoolbox":          `\csuse{input}{x}`,
		"at commands":       `\makeatletter\@@input x`,
		"scantokens":        `\scantokens{\in put}`,
		"package outside":   `\usepackage{../../evil}`,
		"class absolute":    `\documentclass{/tmp/evil}`,
		"windows drive":     `\usepackage[x]{C:/evil}`,
		"home package":      `\RequirePackage{~/evil}`,
	} {
		reason := unsafeSourceReason(source, "Windows")
		if reason == "" || !strings.Contains(reason, "Windows") {
			t.Errorf("%s: expected refusal, got %q", name, reason)
		}
	}
}
