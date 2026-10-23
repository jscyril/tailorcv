package resumeimport

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jscyril/tailorcv/internal/domain"
)

var (
	phonePattern = regexp.MustCompile(`^\+?[\d\s().-]{7,}$`)
	emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// Parse reads a resume written with Jake's resume commands (\resumeSubheading,
// \resumeItem, \resumeProjectHeading) into profile data. Entries that fail
// the normal form validation are left out with a warning, so everything
// returned can be saved as is.
func Parse(source string) domain.ResumeImport {
	source = stripComments(source)
	if begin := strings.Index(source, `\begin{document}`); begin >= 0 {
		source = source[begin+len(`\begin{document}`):]
	}
	result := domain.ResumeImport{Experiences: []domain.ExperienceInput{}, Projects: []domain.ProjectInput{}, Educations: []domain.EducationInput{}, Warnings: []string{}}
	sections := commandIndexes(source, "section")
	header := source
	if len(sections) > 0 {
		header = source[:sections[0]]
	}
	result.Profile = parseHeader(header)

	for position, start := range sections {
		title, bodyStart, ok := readGroup(source, start+len(`\section`))
		if !ok {
			continue
		}
		end := len(source)
		if position+1 < len(sections) {
			end = sections[position+1]
		}
		body := source[bodyStart:end]
		name := strings.ToLower(toText(title))
		switch {
		case strings.Contains(name, "experience") || strings.Contains(name, "employment") || strings.Contains(name, "work"):
			result.Experiences = append(result.Experiences, parseExperiences(body, &result.Warnings)...)
		case strings.Contains(name, "education"):
			result.Educations = append(result.Educations, parseEducations(body, &result.Warnings)...)
		case strings.Contains(name, "project"):
			result.Projects = append(result.Projects, parseProjects(body, &result.Warnings)...)
		case strings.Contains(name, "skill"):
			result.Profile.Skills = append(result.Profile.Skills, parseSkills(body)...)
		default:
			result.Warnings = append(result.Warnings, fmt.Sprintf("The %q section was not imported; add it by hand if you need it.", toText(title)))
		}
	}
	result.Profile.Skills = uniqueFold(result.Profile.Skills)
	if len(result.Experiences)+len(result.Projects)+len(result.Educations) == 0 {
		result.Warnings = append(result.Warnings, `No entries were found. Import reads resumes that use Jake's resume commands (\resumeSubheading, \resumeItem, \resumeProjectHeading).`)
	}
	return result
}

func parseHeader(header string) domain.ProfileInput {
	profile := domain.ProfileInput{}
	if center := strings.Index(header, `\begin{center}`); center >= 0 {
		header = header[center:]
	}
	if index := commandIndexes(header, "textbf"); len(index) > 0 {
		if name, _, ok := readGroup(header, index[0]+len(`\textbf`)); ok {
			profile.Name = toText(name)
		}
	}
	for _, url := range hrefTargets(header) {
		lower := strings.ToLower(url)
		switch {
		case strings.HasPrefix(lower, "mailto:"):
			profile.Email = strings.TrimPrefix(url[len("mailto:"):], "//")
		case strings.HasPrefix(lower, "tel:"):
			profile.Phone = url[len("tel:"):]
		case strings.Contains(lower, "linkedin.com/"):
			profile.LinkedInURL = withScheme(url)
		case strings.Contains(lower, "github.com/"):
			profile.GitHubUsername = strings.Trim(strings.SplitN(lower[strings.Index(lower, "github.com/")+len("github.com/"):], "/", 2)[0], " ")
		case profile.Website == "" && (strings.HasPrefix(lower, "http") || strings.Contains(lower, ".")):
			profile.Website = withScheme(url)
		}
	}
	// Plain-text contact parts, for resumes that do not link every item.
	for _, line := range strings.Split(toText(header), "\n") {
		for _, part := range strings.Split(line, "|") {
			part = strings.TrimSpace(part)
			switch {
			case profile.Email == "" && emailPattern.MatchString(part):
				profile.Email = part
			case profile.Phone == "" && phonePattern.MatchString(part):
				profile.Phone = part
			}
		}
	}
	return profile
}

func withScheme(url string) string {
	if strings.HasPrefix(strings.ToLower(url), "http") {
		return url
	}
	return "https://" + url
}

// heading is one \resumeSubheading or \resumeProjectHeading and the bullets
// that follow it.
type heading struct {
	parts   []string
	bullets []string
}

func headings(body, command string, arguments int) []heading {
	starts := commandIndexes(body, command)
	found := make([]heading, 0, len(starts))
	for index, start := range starts {
		parts, next, ok := readGroups(body, start+len(command)+1, arguments)
		if !ok {
			continue
		}
		end := len(body)
		if index+1 < len(starts) {
			end = starts[index+1]
		}
		item := heading{parts: make([]string, len(parts))}
		for position, part := range parts {
			item.parts[position] = strings.ReplaceAll(toText(part), "\n", " ")
		}
		for _, bulletStart := range commandIndexes(body[next:end], "resumeItem") {
			if text, _, ok := readGroup(body[next:end], bulletStart+len(`\resumeItem`)); ok {
				if bullet := strings.ReplaceAll(toText(text), "\n", " "); bullet != "" {
					item.bullets = append(item.bullets, bullet)
				}
			}
		}
		found = append(found, item)
	}
	return found
}

func evidence(bullets []string) []domain.EvidenceBulletInput {
	result := make([]domain.EvidenceBulletInput, 0, len(bullets))
	for _, text := range bullets {
		// These are the user's own resume claims, confirmed in the review step.
		result = append(result, domain.EvidenceBulletInput{Text: text, Provenance: domain.ProvenanceImported, Verification: domain.VerificationVerified, Importance: domain.EvidenceImportanceStandard})
	}
	return result
}

func parseExperiences(body string, warnings *[]string) []domain.ExperienceInput {
	result := []domain.ExperienceInput{}
	for _, item := range headings(body, "resumeSubheading", 4) {
		// Jake's experience order is {title}{dates}{company}{location}; some
		// copies use the education order {company}{location}{title}{dates}.
		title, dates, company, location := item.parts[0], item.parts[1], item.parts[2], item.parts[3]
		if !looksLikeDate(dates) && looksLikeDate(item.parts[3]) {
			company, location, title, dates = item.parts[0], item.parts[1], item.parts[2], item.parts[3]
		}
		input := domain.ExperienceInput{Title: title, Company: company, Location: location, Bullets: evidence(item.bullets)}
		var ok bool
		input.StartDate, input.EndDate, input.Current, ok = parseRange(dates)
		label := title
		if company != "" {
			label += " at " + company
		}
		if !ok {
			*warnings = append(*warnings, fmt.Sprintf("Skipped %q: could not read the dates %q. Add it on the Experience screen.", label, dates))
			continue
		}
		if _, err := input.Validate(); err != nil {
			*warnings = append(*warnings, fmt.Sprintf("Skipped %q: %v.", label, err))
			continue
		}
		result = append(result, input)
	}
	return result
}

func parseEducations(body string, warnings *[]string) []domain.EducationInput {
	result := []domain.EducationInput{}
	for _, item := range headings(body, "resumeSubheading", 4) {
		// Jake's education order is {school}{location}{degree}{dates}.
		institution, location, degree, dates := item.parts[0], item.parts[1], item.parts[2], item.parts[3]
		if !looksLikeDate(dates) && looksLikeDate(location) {
			location, dates = dates, location
		}
		input := domain.EducationInput{Institution: institution, Location: location, Degree: degree, Details: strings.Join(item.bullets, "\n")}
		if before, after, found := strings.Cut(degree, " in "); found {
			input.Degree, input.FieldOfStudy = before, after
		}
		if start, end, current, ok := parseRange(dates); ok {
			input.StartDate, input.EndDate, input.Current = start, end, current
			if start == end {
				input.StartDate = ""
			}
		} else if dates != "" {
			*warnings = append(*warnings, fmt.Sprintf("Could not read the dates %q for %s; add them on the Education screen.", dates, institution))
		}
		if _, err := input.Validate(); err != nil {
			*warnings = append(*warnings, fmt.Sprintf("Skipped education at %q: %v.", institution, err))
			continue
		}
		result = append(result, input)
	}
	return result
}

func parseProjects(body string, warnings *[]string) []domain.ProjectInput {
	result := []domain.ProjectInput{}
	for _, item := range headings(body, "resumeProjectHeading", 2) {
		// The first argument is usually "Name | Technologies".
		name, technologies, _ := strings.Cut(item.parts[0], "|")
		input := domain.ProjectInput{Name: strings.TrimSpace(name), Skills: splitList(technologies), Provenance: domain.ProvenanceImported, Verification: domain.VerificationVerified, ResumeEligible: true, Bullets: evidence(item.bullets)}
		if start, end, current, ok := parseRange(item.parts[1]); ok {
			input.StartDate, input.EndDate, input.Ongoing = start, end, current
		} else if item.parts[1] != "" {
			*warnings = append(*warnings, fmt.Sprintf("Could not read the dates %q for project %s; add them on the Projects screen.", item.parts[1], input.Name))
		}
		if _, err := input.Validate(); err != nil {
			*warnings = append(*warnings, fmt.Sprintf("Skipped project %q: %v.", input.Name, err))
			continue
		}
		result = append(result, input)
	}
	return result
}

// parseSkills reads "Languages: Go, Python" style lines.
func parseSkills(body string) []string {
	skills := []string{}
	for _, line := range strings.Split(toText(body), "\n") {
		if _, after, found := strings.Cut(line, ":"); found {
			line = after
		}
		skills = append(skills, splitList(line)...)
	}
	return skills
}

// splitList splits on commas that are not inside parentheses, so
// "SQL (Postgres, SQLite), Go" stays two skills.
func splitList(value string) []string {
	items := []string{}
	depth, start := 0, 0
	for index, character := range value + "," {
		switch character {
		case '(':
			depth++
		case ')':
			depth--
		case ',', ';':
			if depth <= 0 {
				if item := strings.Trim(strings.TrimSpace(value[start:min(index, len(value))]), "."); item != "" {
					items = append(items, item)
				}
				start = index + 1
			}
		}
	}
	return items
}

func uniqueFold(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}
