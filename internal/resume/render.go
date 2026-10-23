package resume

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jscyril/tailorcv/internal/domain"
)

type Data struct {
	Profile        domain.Profile
	Experiences    []domain.Experience
	Projects       []domain.Project
	Educations     []domain.Education
	Certifications []domain.Certification
	Achievements   []domain.Achievement
}

func Render(source string, data Data) string {
	contact := compact([]string{data.Profile.Email, data.Profile.Phone, data.Profile.Location, data.Profile.Website, data.Profile.LinkedInURL, githubURL(data.Profile.GitHubUsername)})
	for _, link := range data.Profile.ContactLinks {
		contact = append(contact, link.URL)
	}
	links := contactLinks(contact)
	for index := range contact {
		contact[index] = escape(contact[index])
	}
	replacements := map[string]string{
		"{{TAILORCV_NAME}}":                   fallback(escape(data.Profile.Name), "Your Name"),
		"{{TAILORCV_HEADLINE}}":               fallback(escape(data.Profile.Headline), "Professional Headline"),
		"{{TAILORCV_HEADLINE_LINE}}":          headlineLine(data.Profile.Headline),
		"{{TAILORCV_CONTACT_LINKS}}":          links,
		"{{TAILORCV_CONTACT}}":                strings.Join(contact, ` $\vert$ `),
		"{{TAILORCV_SUMMARY_SECTION}}":        summarySection(data.Profile.Summary),
		"{{TAILORCV_EXPERIENCE_SECTION}}":     experienceSection(data.Experiences),
		"{{TAILORCV_PROJECTS_SECTION}}":       projectSection(data.Projects),
		"{{TAILORCV_EDUCATION_SECTION}}":      educationSection(data.Educations),
		"{{TAILORCV_CERTIFICATIONS_SECTION}}": certificationSection(data.Certifications),
		"{{TAILORCV_ACHIEVEMENTS_SECTION}}":   achievementSection(data.Achievements),
		"{{TAILORCV_SKILLS_SECTION}}":         skillsSection(data.Profile.Skills),
	}
	for marker, value := range replacements {
		source = strings.ReplaceAll(source, marker, value)
	}
	return source
}

// headlineLine is the headline on its own line, or nothing when it is empty.
func headlineLine(headline string) string {
	if strings.TrimSpace(headline) == "" {
		return ""
	}
	return escape(headline) + "\\\\[2pt]"
}

// contactLinks joins contact details, linking emails and web addresses the
// way Jake's resume does. It needs hyperref, so only built-ins use it.
func contactLinks(items []string) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		label := escape(item)
		switch {
		case strings.Contains(item, "@") && !strings.ContainsAny(item, " /"):
			parts = append(parts, `\href{mailto:`+label+`}{\underline{`+label+`}}`)
		case strings.Contains(item, ".") && !strings.Contains(item, " ") && !phoneLike.MatchString(item):
			target := item
			if !strings.HasPrefix(strings.ToLower(item), "http") {
				target = "https://" + item
			}
			shown := strings.TrimPrefix(strings.TrimPrefix(item, "https://"), "http://")
			parts = append(parts, `\href{`+escape(target)+`}{\underline{`+escape(shown)+`}}`)
		default:
			parts = append(parts, label)
		}
	}
	return strings.Join(parts, ` $|$ `)
}

var phoneLike = regexp.MustCompile(`^\+?[\d\s().-]+$`)

func certificationSection(items []domain.Certification) string {
	if len(items) == 0 {
		return ""
	}
	return section("Certifications", func(output *strings.Builder) {
		for _, item := range items {
			fmt.Fprintf(output, "\\tailorcvEntry{%s}{%s}{%s}{}\n", escape(item.Name), escape(displayMonth(item.IssueDate)), escape(item.Issuer))
			if item.Description != "" {
				writeBullets(output, []domain.EvidenceBullet{{Text: item.Description}})
			}
		}
	})
}

func achievementSection(items []domain.Achievement) string {
	if len(items) == 0 {
		return ""
	}
	var output strings.Builder
	output.WriteString(sectionCommandDefaults + "\\section{Achievements}\n\\tailorcvBulletsStart\n")
	for _, item := range items {
		text := "\\textbf{" + escape(item.Title) + "}: " + escape(item.Description)
		if item.Date != "" {
			text += " (" + escape(displayMonth(item.Date)) + ")"
		}
		fmt.Fprintf(&output, "  \\tailorcvBullet{%s}\n", text)
	}
	output.WriteString("\\tailorcvBulletsEnd")
	return output.String()
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

// Sections are written with these commands so a template can restyle every
// entry by defining them before \begin{document}. \providecommand keeps the
// template's definition and otherwise falls back to a plain layout.
const sectionCommandDefaults = `\providecommand{\tailorcvListStart}{}
\providecommand{\tailorcvListEnd}{}
\providecommand{\tailorcvEntry}[4]{\textbf{#1}\hfill #2\par\textit{#3}\hfill\textit{#4}\par}
\providecommand{\tailorcvProject}[3]{\textbf{#1}\if\relax\detokenize{#2}\relax\else{} $|$ \emph{#2}\fi\hfill #3\par}
\providecommand{\tailorcvBulletsStart}{\begin{itemize}}
\providecommand{\tailorcvBulletsEnd}{\end{itemize}}
\providecommand{\tailorcvBullet}[1]{\item #1}
\providecommand{\tailorcvSkills}[1]{#1}
`

func section(title string, body func(*strings.Builder)) string {
	var output strings.Builder
	output.WriteString(sectionCommandDefaults)
	fmt.Fprintf(&output, "\\section{%s}\n\\tailorcvListStart\n", title)
	body(&output)
	output.WriteString("\\tailorcvListEnd")
	return output.String()
}

func experienceSection(experiences []domain.Experience) string {
	if len(experiences) == 0 {
		return ""
	}
	return section("Experience", func(output *strings.Builder) {
		for _, experience := range experiences {
			fmt.Fprintf(output, "\\tailorcvEntry{%s}{%s}{%s}{%s}\n", escape(experience.Title), escape(dateRange(experience.StartDate, experience.EndDate, experience.Current)), escape(experience.Company), escape(experience.Location))
			writeBullets(output, experience.Bullets)
		}
	})
}

func projectSection(projects []domain.Project) string {
	if len(projects) == 0 {
		return ""
	}
	return section("Projects", func(output *strings.Builder) {
		for _, project := range projects {
			detail := strings.Join(project.Skills, ", ")
			if detail == "" {
				detail = project.Role
			}
			dates := ""
			if project.StartDate != "" || project.EndDate != "" || project.Ongoing {
				dates = dateRange(project.StartDate, project.EndDate, project.Ongoing)
			}
			fmt.Fprintf(output, "\\tailorcvProject{%s}{%s}{%s}\n", escape(project.Name), escape(detail), escape(dates))
			bullets := project.Bullets
			if project.Description != "" && len(bullets) == 0 {
				bullets = []domain.EvidenceBullet{{Text: project.Description}}
			}
			writeBullets(output, bullets)
		}
	})
}

func educationSection(educations []domain.Education) string {
	if len(educations) == 0 {
		return ""
	}
	return section("Education", func(output *strings.Builder) {
		for _, education := range educations {
			degree := education.Degree
			if education.FieldOfStudy != "" {
				degree += " in " + education.FieldOfStudy
			}
			fmt.Fprintf(output, "\\tailorcvEntry{%s}{%s}{%s}{%s}\n", escape(education.Institution), escape(education.Location), escape(degree), escape(dateRange(education.StartDate, education.EndDate, education.Current)))
			if education.Details != "" {
				lines := []domain.EvidenceBullet{}
				for _, line := range strings.Split(education.Details, "\n") {
					if line = strings.TrimSpace(line); line != "" {
						lines = append(lines, domain.EvidenceBullet{Text: line})
					}
				}
				writeBullets(output, lines)
			}
		}
	})
}

func skillsSection(skills []string) string {
	if len(skills) == 0 {
		return ""
	}
	return sectionCommandDefaults + "\\section{Skills}\n\\tailorcvSkills{" + escape(strings.Join(skills, ", ")) + "}"
}

func writeBullets(output *strings.Builder, bullets []domain.EvidenceBullet) {
	if len(bullets) == 0 {
		return
	}
	output.WriteString("\\tailorcvBulletsStart\n")
	for _, bullet := range bullets {
		fmt.Fprintf(output, "  \\tailorcvBullet{%s}\n", escape(bullet.Text))
	}
	output.WriteString("\\tailorcvBulletsEnd\n")
}

var monthNames = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// displayMonth turns YYYY-MM into "Jan 2024"; anything else is kept as is.
func displayMonth(value string) string {
	var year, month int
	if _, err := fmt.Sscanf(value, "%4d-%2d", &year, &month); err != nil || month < 1 || month > 12 || len(value) != 7 {
		return value
	}
	return fmt.Sprintf("%s %d", monthNames[month-1], year)
}

func dateRange(start, end string, current bool) string {
	start, end = displayMonth(start), displayMonth(end)
	if current {
		end = "Present"
	}
	if start == "" || start == end {
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

// pdfTeX-only lines common in downloaded templates (Jake's resume among them).
// Tectonic runs XeTeX, which already writes Unicode text, so they are unneeded.
var pdfTeXOnlyLine = regexp.MustCompile(`(?m)^[ \t]*(\\input\{glyphtounicode\}|\\pdfgentounicode\s*=\s*1)[ \t]*(\r?)$`)

// CommentOutPDFTeXOnly replaces those lines with a note and reports how many
// it changed. The note holds no backslash, because the Windows safety scan
// reads comments too and would refuse a commented-out \input.
func CommentOutPDFTeXOnly(source string) (string, int) {
	count := len(pdfTeXOnlyLine.FindAllStringIndex(source, -1))
	// $2 keeps a Windows line ending (\r) when the file has one.
	return pdfTeXOnlyLine.ReplaceAllString(source, "% TailorCV removed a pdfTeX-only line (glyphtounicode setup); Tectonic does not need it.$2"), count
}
