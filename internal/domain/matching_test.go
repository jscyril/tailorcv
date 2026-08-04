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

func TestAnalyzeCareerEvidenceUsesAliasesAndRanksVerifiedFacts(t *testing.T) {
	experiences := []Experience{{
		ID: "experience-1", Title: "Platform Engineer", Company: "Example",
		Bullets: []EvidenceBullet{
			{ID: "fact-1", Text: "Built an audited deployment platform for production services", Verification: VerificationVerified},
			{ID: "fact-2", Text: "Organized a monthly engineering book club", Verification: VerificationVerified},
		},
	}}
	projects := []Project{{
		ID: "project-1", Name: "Release Console", Skills: []string{"Kubernetes", "PostgreSQL"}, ResumeEligible: true,
		Bullets: []EvidenceBullet{{ID: "fact-3", Text: "Reduced deployment time for distributed services", Verification: VerificationVerified}},
	}}
	analysis, err := AnalyzeCareerEvidence(JobAnalysisInput{
		Description: "We need a platform engineer to operate production services on K8s and Postgres with auditable deployments.",
	}, []string{"Kubernetes", "K8s", "PostgreSQL", "Rust"}, experiences, projects)
	if err != nil {
		t.Fatalf("AnalyzeCareerEvidence() error = %v", err)
	}
	if analysis.Score != 67 {
		t.Fatalf("Score = %d, want 67", analysis.Score)
	}
	if len(analysis.DetectedSkills) != 2 || len(analysis.RankedEvidence) != 2 {
		t.Fatalf("analysis = %#v", analysis)
	}
	if analysis.RankedEvidence[0].FactID != "fact-3" || analysis.RankedEvidence[0].Score <= analysis.RankedEvidence[1].Score {
		t.Fatalf("ranked evidence = %#v", analysis.RankedEvidence)
	}
}

func TestJobInputNormalizesAndLimitsDescription(t *testing.T) {
	job, err := (JobInput{Company: "  Example   Systems ", Role: " Platform  Engineer ", Description: "  A sufficiently long job description for deterministic matching and local storage.  "}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if job.Company != "Example Systems" || job.Role != "Platform Engineer" {
		t.Fatalf("job = %#v", job)
	}
}
