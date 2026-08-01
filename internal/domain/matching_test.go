package domain

import "testing"

func TestAnalyzeJobDescriptionUsesSkillBoundaries(t *testing.T) {
	analysis, err := AnalyzeJobDescription(JobAnalysisInput{
		Description: "We need an engineer who builds reliable services in Go and PostgreSQL for distributed systems.",
	}, []string{"Go", "Rust", "PostgreSQL"})
	if err != nil {
		t.Fatalf("AnalyzeJobDescription() error = %v", err)
	}
	if analysis.Score != 67 {
		t.Fatalf("Score = %d, want 67", analysis.Score)
	}
	if len(analysis.MatchedSkills) != 2 {
		t.Fatalf("MatchedSkills = %#v", analysis.MatchedSkills)
	}
}

func TestAnalyzeJobDescriptionRejectsShortText(t *testing.T) {
	_, err := AnalyzeJobDescription(JobAnalysisInput{Description: "Go developer"}, []string{"Go"})
	if err == nil {
		t.Fatal("AnalyzeJobDescription() expected validation error")
	}
}
