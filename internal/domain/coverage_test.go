package domain

import "testing"

func TestCheckKeywordCoverage(t *testing.T) {
	job := JobAnalysisInput{Role: "Engineer", Description: "We need Go, Kubernetes, and C# experience. PostgreSQL is a plus. Terraform is nice to have."}
	source := `\section{Skills} Golang, C\# and Postgres \\ \textbf{K8s}`
	coverage, err := CheckKeywordCoverage(job, nil, nil, source)
	if err != nil {
		t.Fatalf("CheckKeywordCoverage() error = %v", err)
	}
	got := map[string]bool{}
	for _, item := range append(coverage.Required, coverage.Preferred...) {
		got[item.Skill] = item.Present
	}
	// Aliases (Golang, K8s, Postgres) and LaTeX-escaped C\# count as present.
	want := map[string]bool{"Go": true, "Kubernetes": true, "C#": true, "PostgreSQL": true, "Terraform": false}
	for skill, present := range want {
		if value, ok := got[skill]; !ok || value != present {
			t.Errorf("%s: present=%v found=%v, want present=%v (coverage %#v)", skill, value, ok, present, coverage)
		}
	}
	if len(coverage.Preferred) != 2 {
		t.Errorf("preferred = %#v, want PostgreSQL and Terraform", coverage.Preferred)
	}
}
