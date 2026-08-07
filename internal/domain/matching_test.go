package domain

import (
	"testing"
	"time"
)

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

func TestAnalyzeCareerEvidenceExtractsStructuredRequirements(t *testing.T) {
	analysis, err := AnalyzeCareerEvidence(JobAnalysisInput{Description: `Required: Go and Kubernetes.
Preferred: PostgreSQL experience is a plus.
Design and operate reliable production APIs for customers.`}, []string{"Go"}, nil, nil)
	if err != nil {
		t.Fatalf("AnalyzeCareerEvidence() error = %v", err)
	}
	if len(analysis.RequiredSkills) != 2 || len(analysis.PreferredSkills) != 1 || analysis.PreferredSkills[0] != "PostgreSQL" {
		t.Fatalf("skill requirements = required %#v, preferred %#v", analysis.RequiredSkills, analysis.PreferredSkills)
	}
	if len(analysis.Responsibilities) != 1 || len(analysis.Keywords) == 0 || len(analysis.SearchTerms) == 0 {
		t.Fatalf("structured analysis = %#v", analysis)
	}
}

func TestAnalyzeCareerEvidenceUsesIndexedSearchSignal(t *testing.T) {
	experiences := []Experience{{ID: "experience", Title: "Engineer", Company: "Example", Bullets: []EvidenceBullet{{ID: "indexed-fact", Text: "Automated release workflows"}}}}
	analysis, err := AnalyzeCareerEvidenceWithSearch(JobAnalysisInput{Description: "Build dependable customer systems and improve operational quality across the software platform."}, nil, experiences, nil, []EvidenceSearchHit{{FactID: "indexed-fact", Score: 18}})
	if err != nil {
		t.Fatalf("AnalyzeCareerEvidenceWithSearch() error = %v", err)
	}
	if len(analysis.RankedEvidence) != 1 || analysis.RankedEvidence[0].Score != 18 || analysis.RankedEvidence[0].Reasons[0] != "Indexed evidence search match" {
		t.Fatalf("ranked evidence = %#v", analysis.RankedEvidence)
	}
}

func TestAnalyzeCareerEvidenceRanksUserImportanceAndRecency(t *testing.T) {
	experiences := []Experience{
		{ID: "older", Title: "Engineer", Company: "Earlier", EndDate: "2018-01", Bullets: []EvidenceBullet{{ID: "older-fact", Text: "Built reliable Go deployment services", Importance: EvidenceImportanceStandard}}},
		{ID: "recent", Title: "Engineer", Company: "Recent", EndDate: "2026-01", Bullets: []EvidenceBullet{{ID: "recent-fact", Text: "Built reliable Go deployment services", Importance: EvidenceImportanceImportant}}},
	}
	analysis, err := analyzeCareerEvidenceWithSearchAt(JobAnalysisInput{Description: "Build reliable Go deployment services for a growing production software platform."}, []string{"Go"}, experiences, nil, nil, time.Date(2026, time.August, 8, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("analyzeCareerEvidenceWithSearchAt() error = %v", err)
	}
	if len(analysis.RankedEvidence) != 2 || analysis.RankedEvidence[0].FactID != "recent-fact" {
		t.Fatalf("ranked evidence = %#v", analysis.RankedEvidence)
	}
	wantReasons := map[string]bool{"Marked important by you": false, "Role ended within 2 years": false}
	for _, reason := range analysis.RankedEvidence[0].Reasons {
		if _, exists := wantReasons[reason]; exists {
			wantReasons[reason] = true
		}
	}
	for reason, found := range wantReasons {
		if !found {
			t.Fatalf("missing ranking reason %q in %#v", reason, analysis.RankedEvidence[0].Reasons)
		}
	}
}
