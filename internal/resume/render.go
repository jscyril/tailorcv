package resume

import (
	"fmt"
	"strings"

	"github.com/jscyril/tailorcv/internal/domain"
)

type Data struct {
	Profile     domain.Profile
	Experiences []domain.Experience
	Projects    []domain.Project
	Educations  []domain.Education
}

func Render(source string, data Data) string {
	contact := compact([]string{data.Profile.Email, data.Profile.Phone, data.Profile.Location, data.Profile.Website, data.Profile.LinkedInURL, githubURL(data.Profile.GitHubUsername)})
	for index := range contact {
		contact[index] = escape(contact[index])
	}
	replacements := map[string]string{
		"{{TAILORCV_NAME}}":               fallback(escape(data.Profile.Name), "Your Name"),
		"{{TAILORCV_HEADLINE}}":           fallback(escape(data.Profile.Headline), "Professional Headline"),
		"{{TAILORCV_CONTACT}}":            strings.Join(contact, ` $\vert$ `),
		"{{TAILORCV_SUMMARY_SECTION}}":    summarySection(data.Profile.Summary),
		"{{TAILORCV_EXPERIENCE_SECTION}}": experienceSection(data.Experiences),
		"{{TAILORCV_PROJECTS_SECTION}}":   projectSection(data.Projects),
		"{{TAILORCV_EDUCATION_SECTION}}":  educationSection(data.Educations),
		"{{TAILORCV_SKILLS_SECTION}}":     skillsSection(data.Profile.Skills),
	}
	for marker, value := range replacements {
		source = strings.ReplaceAll(source, marker, value)
	}
	return source
}

func escape(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\textbackslash{}`,
		`{`, `\{`,
		`}`, `\}`,
		`$`, `\$`,
		`&`, `\&`,
		`#`, `\#`,
		`%`, `\%`,
		`_`, `\_`,
		`~`, `\textasciitilde{}`,
		`^`, `\textasciicircum{}`,
	)
	return replacer.Replace(strings.TrimSpace(value))
}

func summarySection(summary string) string {
	if strings.TrimSpace(summary) == "" {
		return ""
	}
	return "\\section{Summary}\n" + escape(summary)
}

func experienceSection(experiences []domain.Experience) string {
	if len(experiences) == 0 {
		return ""
	}
	var output strings.Builder
	output.WriteString("\\section{Experience}\n")
	for _, experience := range experiences {
		fmt.Fprintf(&output, "\\textbf{%s} $\\cdot$ %s\\hfill %s\\\\\n", escape(experience.Title), escape(experience.Company), escape(dateRange(experience.StartDate, experience.EndDate, experience.Current)))
		if experience.Location != "" {
			fmt.Fprintf(&output, "\\textit{%s}\\\\\n", escape(experience.Location))
		}
		writeBullets(&output, experience.Bullets)
		output.WriteString("\\vspace{3pt}\n")
	}
	return strings.TrimSpace(output.String())
}

func projectSection(projects []domain.Project) string {
	if len(projects) == 0 {
		return ""
	}
	var output strings.Builder
	output.WriteString("\\section{Projects}\n")
	for _, project := range projects {
		fmt.Fprintf(&output, "\\textbf{%s}", escape(project.Name))
		if project.Role != "" {
			fmt.Fprintf(&output, " $\\cdot$ %s", escape(project.Role))
		}
		if project.StartDate != "" || project.EndDate != "" || project.Ongoing {
			fmt.Fprintf(&output, "\\hfill %s", escape(dateRange(project.StartDate, project.EndDate, project.Ongoing)))
		}
		output.WriteString("\\\\\n")
		if project.Description != "" {
			output.WriteString(escape(project.Description) + "\\\\\n")
		}
		writeBullets(&output, project.Bullets)
		output.WriteString("\\vspace{3pt}\n")
	}
	return strings.TrimSpace(output.String())
}

func educationSection(educations []domain.Education) string {
	if len(educations) == 0 {
		return ""
	}
	var output strings.Builder
	output.WriteString("\\section{Education}\n")
	for _, education := range educations {
		fmt.Fprintf(&output, "\\textbf{%s} $\\cdot$ %s", escape(education.Institution), escape(education.Degree))
		if education.FieldOfStudy != "" {
			fmt.Fprintf(&output, " in %s", escape(education.FieldOfStudy))
		}
		fmt.Fprintf(&output, "\\hfill %s\\\\\n", escape(dateRange(education.StartDate, education.EndDate, education.Current)))
		if education.Details != "" {
			output.WriteString(escape(education.Details) + "\\\\\n")
		}
	}
	return strings.TrimSpace(output.String())
}

func skillsSection(skills []string) string {
	if len(skills) == 0 {
		return ""
	}
	return "\\section{Skills}\n" + escape(strings.Join(skills, ", "))
}

func writeBullets(output *strings.Builder, bullets []domain.EvidenceBullet) {
	if len(bullets) == 0 {
		return
	}
	output.WriteString("\\begin{itemize}\n")
	for _, bullet := range bullets {
		fmt.Fprintf(output, "  \\item %s\n", escape(bullet.Text))
	}
	output.WriteString("\\end{itemize}\n")
}

func dateRange(start, end string, current bool) string {
	if current {
		end = "Present"
	}
	if start == "" {
		return end
	}
	if end == "" {
		return start
	}
	return start + " -- " + end
}

func githubURL(username string) string {
	if strings.TrimSpace(username) == "" {
		return ""
	}
	return "github.com/" + strings.TrimSpace(username)
}

func compact(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

func fallback(value, fallbackValue string) string {
	if value == "" {
		return fallbackValue
	}
	return value
}
