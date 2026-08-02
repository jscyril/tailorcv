package domain

import (
	"strings"
	"testing"
)

func TestResumeTemplateValidation(t *testing.T) {
	source := `\documentclass{article}\begin{document}Hello\end{document}`
	template, err := (ResumeTemplateInput{Name: "  My   Template ", Source: source}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if template.Name != "My Template" || template.Source != source {
		t.Fatalf("Validate() = %#v", template)
	}

	if _, err := (ResumeTemplateInput{Name: "Broken", Source: `\section{No document}`}).Validate(); err == nil {
		t.Fatal("Validate() accepted an incomplete document")
	}
	if _, err := (ResumeTemplateInput{Name: "Large", Source: strings.Repeat("x", MaxTemplateSourceBytes+1)}).Validate(); err == nil {
		t.Fatal("Validate() accepted an oversized source")
	}
}
