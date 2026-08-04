package main

import (
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestSelectResumeEvidenceFiltersFactsAndRejectsLockedProjects(t *testing.T) {
	experiences := []domain.Experience{{ID: "experience", Bullets: []domain.EvidenceBullet{{ID: "experience-fact-1", Text: "Selected"}, {ID: "experience-fact-2", Text: "Not selected"}}}}
	projects := []domain.Project{
		{ID: "eligible", Name: "Eligible", ResumeEligible: true, Bullets: []domain.EvidenceBullet{{ID: "project-fact", Text: "Selected project fact"}}},
		{ID: "locked", Name: "Locked", ResumeEligible: false, Bullets: []domain.EvidenceBullet{{ID: "locked-fact", Text: "Locked project fact"}}},
	}
	selectedExperiences, selectedProjects, err := selectResumeEvidence([]string{"experience-fact-1", "project-fact"}, experiences, projects)
	if err != nil {
		t.Fatalf("selectResumeEvidence() error = %v", err)
	}
	if len(selectedExperiences) != 1 || len(selectedExperiences[0].Bullets) != 1 || selectedExperiences[0].Bullets[0].ID != "experience-fact-1" {
		t.Fatalf("selected experiences = %#v", selectedExperiences)
	}
	if len(selectedProjects) != 1 || len(selectedProjects[0].Bullets) != 1 || selectedProjects[0].Bullets[0].ID != "project-fact" {
		t.Fatalf("selected projects = %#v", selectedProjects)
	}
	if _, _, err := selectResumeEvidence([]string{"locked-fact"}, experiences, projects); err == nil {
		t.Fatal("selectResumeEvidence() expected locked project error")
	}
	if _, _, err := selectResumeEvidence([]string{"missing"}, experiences, projects); err == nil {
		t.Fatal("selectResumeEvidence() expected unknown evidence error")
	}
}
