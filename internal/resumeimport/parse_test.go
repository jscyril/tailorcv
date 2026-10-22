package resumeimport

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestParseJakeResume(t *testing.T) {
	source, err := os.ReadFile("testdata/jake.tex")
	if err != nil {
		t.Fatal(err)
	}
	result := Parse(string(source))

	profile := result.Profile
	if profile.Name != "Casey Fictional" || profile.Email != "casey@example.com" || profile.Phone != "+1 555 010 0199" || profile.LinkedInURL != "https://linkedin.com/in/casey-fictional" || profile.GitHubUsername != "caseyf" {
		t.Errorf("profile = %#v", profile)
	}
	wantSkills := []string{"Go", "Python", "C/C++", "SQL (Postgres)", "JavaScript", "React", "Flask", "FastAPI", "Git", "Docker"}
	if !reflect.DeepEqual(profile.Skills, wantSkills) {
		t.Errorf("skills = %q, want %q (duplicates removed ignoring case)", profile.Skills, wantSkills)
	}

	if len(result.Experiences) != 2 {
		t.Fatalf("experiences = %#v", result.Experiences)
	}
	first := result.Experiences[0]
	if first.Title != "Software Engineer" || first.Company != "Example Systems" || first.Location != "Remote" || first.StartDate != "2022-06" || !first.Current {
		t.Errorf("first experience = %#v", first)
	}
	if len(first.Bullets) != 2 || first.Bullets[0].Text != "Built a Go deployment service that cut release time by 40% for 30 teams" || first.Bullets[1].Text != "Migrated R&D workloads to Kubernetes and kept costs flat" {
		t.Errorf("bullets = %#v", first.Bullets)
	}
	if first.Bullets[0].Provenance != domain.ProvenanceImported || first.Bullets[0].Verification != domain.VerificationVerified {
		t.Errorf("imported bullets must be marked imported and verified: %#v", first.Bullets[0])
	}
	if intern := result.Experiences[1]; intern.StartDate != "2021-06" || intern.EndDate != "2021-06" || intern.Title != "Data Analyst Intern" {
		t.Errorf("intern = %#v", intern)
	}

	if len(result.Educations) != 1 {
		t.Fatalf("educations = %#v", result.Educations)
	}
	if education := result.Educations[0]; education.Institution != "Example State University" || education.Degree != "Bachelor of Science" || education.FieldOfStudy != "Computer Science, Minor in Mathematics" || education.StartDate != "2018-08" || education.EndDate != "2022-05" || education.Location != "Austin, TX" {
		t.Errorf("education = %#v", education)
	}

	if len(result.Projects) != 1 {
		t.Fatalf("projects = %#v", result.Projects)
	}
	if project := result.Projects[0]; project.Name != "Gitstats" || !reflect.DeepEqual(project.Skills, []string{"Python", "Flask", "React", "PostgreSQL (with Timescale)", "Docker"}) || !project.Ongoing || !project.ResumeEligible || len(project.Bullets) != 1 {
		t.Errorf("project = %#v", project)
	}

	warnings := strings.Join(result.Warnings, "\n")
	for _, expected := range []string{`Skipped "Volunteer at Food Bank": could not read the dates "Sometime"`, `The "Leadership" section was not imported`} {
		if !strings.Contains(warnings, expected) {
			t.Errorf("warnings missing %q:\n%s", expected, warnings)
		}
	}
}

func TestParseNonJakeResumeWarns(t *testing.T) {
	result := Parse(`\documentclass{article}\begin{document}\section{Work}\begin{itemize}\item Did things\end{itemize}\end{document}`)
	if len(result.Experiences) != 0 || len(result.Warnings) == 0 || !strings.Contains(result.Warnings[len(result.Warnings)-1], "No entries were found") {
		t.Fatalf("result = %#v", result)
	}
}

func TestToTextKeepsLinkTextAndUndoesEscapes(t *testing.T) {
	got := toText(`\textbf{\Huge A\&B} \href{https://x.y}{\underline{site}} 50\% $|$ \vspace{-2pt}C~D \\[2pt] next`)
	if got != "A&B site 50% | C D\nnext" {
		t.Fatalf("toText() = %q", got)
	}
}
