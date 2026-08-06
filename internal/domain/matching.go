package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const maxRankedEvidence = 20

type JobAnalysisInput struct {
	ID          string `json:"id"`
	Company     string `json:"company"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

func (input JobAnalysisInput) JobInput() JobInput {
	return JobInput{ID: input.ID, Company: input.Company, Role: input.Role, Description: input.Description}
}

type EvidenceMatch struct {
	FactID        string   `json:"factId"`
	SourceID      string   `json:"sourceId"`
	SourceType    string   `json:"sourceType"`
	SourceLabel   string   `json:"sourceLabel"`
	Text          string   `json:"text"`
	Score         int      `json:"score"`
	MatchedSkills []string `json:"matchedSkills"`
	Reasons       []string `json:"reasons"`
	Verified      bool     `json:"verified"`
	Selectable    bool     `json:"selectable"`
}

type EvidenceSearchHit struct {
	FactID string `json:"factId"`
	Score  int    `json:"score"`
}

type JobAnalysis struct {
	Job               Job             `json:"job"`
	Score             int             `json:"score"`
	MatchedSkills     []string        `json:"matchedSkills"`
	UnmentionedSkills []string        `json:"unmentionedSkills"`
	DetectedSkills    []string        `json:"detectedSkills"`
	RequiredSkills    []string        `json:"requiredSkills"`
	PreferredSkills   []string        `json:"preferredSkills"`
	Responsibilities  []string        `json:"responsibilities"`
	Keywords          []string        `json:"keywords"`
	SearchTerms       []string        `json:"searchTerms"`
	RankedEvidence    []EvidenceMatch `json:"rankedEvidence"`
	Explanation       string          `json:"explanation"`
}

// AnalyzeJobDescription preserves the small domain API used by callers that
// only have a skill inventory. The full application uses AnalyzeCareerEvidence.
func AnalyzeJobDescription(input JobAnalysisInput, profileSkills []string) (JobAnalysis, error) {
	return AnalyzeCareerEvidence(input, profileSkills, nil, nil)
}

// AnalyzeCareerEvidence compares a role with user-approved local facts. It is
// deterministic: every score is derived from visible skill and term matches.
func AnalyzeCareerEvidence(input JobAnalysisInput, profileSkills []string, experiences []Experience, projects []Project) (JobAnalysis, error) {
	return AnalyzeCareerEvidenceWithSearch(input, profileSkills, experiences, projects, nil)
}

func AnalyzeCareerEvidenceWithSearch(input JobAnalysisInput, profileSkills []string, experiences []Experience, projects []Project, searchHits []EvidenceSearchHit) (JobAnalysis, error) {
	job, err := input.JobInput().Validate()
	if err != nil {
		return JobAnalysis{}, err
	}
	description := job.Description
	analysis := JobAnalysis{
		Job:               job,
		MatchedSkills:     make([]string, 0),
		UnmentionedSkills: make([]string, 0),
		DetectedSkills:    make([]string, 0),
		RequiredSkills:    make([]string, 0),
		PreferredSkills:   make([]string, 0),
		Responsibilities:  extractResponsibilities(description),
		Keywords:          extractKeywords(description, 10),
		RankedEvidence:    make([]EvidenceMatch, 0),
	}

	normalizedProfileSkills := normalizeMatchingSkills(profileSkills)
	for _, skill := range normalizedProfileSkills {
		if mentionsSkill(description, skill) {
			analysis.MatchedSkills = append(analysis.MatchedSkills, skill)
		} else {
			analysis.UnmentionedSkills = append(analysis.UnmentionedSkills, skill)
		}
	}

	knownSkills := append([]string(nil), normalizedProfileSkills...)
	knownSkills = append(knownSkills, recognizedSkillCatalog...)
	for _, project := range projects {
		knownSkills = append(knownSkills, project.Skills...)
		for _, language := range project.DetectedLanguages {
			knownSkills = append(knownSkills, language.Name)
		}
	}
	knownSkills = normalizeMatchingSkills(knownSkills)
	for _, skill := range knownSkills {
		if mentionsSkill(description, skill) {
			analysis.DetectedSkills = appendUniqueFold(analysis.DetectedSkills, skill)
			if skillIsPreferred(description, skill) {
				analysis.PreferredSkills = appendUniqueFold(analysis.PreferredSkills, skill)
			} else {
				analysis.RequiredSkills = appendUniqueFold(analysis.RequiredSkills, skill)
			}
		}
	}

	sort.Strings(analysis.MatchedSkills)
	sort.Strings(analysis.UnmentionedSkills)
	sort.Strings(analysis.DetectedSkills)
	sort.Strings(analysis.RequiredSkills)
	sort.Strings(analysis.PreferredSkills)
	analysis.SearchTerms = append([]string(nil), analysis.DetectedSkills...)
	for _, keyword := range analysis.Keywords {
		analysis.SearchTerms = appendUniqueFold(analysis.SearchTerms, keyword)
	}
	if len(normalizedProfileSkills) > 0 {
		analysis.Score = int(float64(len(analysis.MatchedSkills))/float64(len(normalizedProfileSkills))*100 + 0.5)
	}
	analysis.Explanation = fmt.Sprintf(
		"%d of %d profile skills are explicitly requested. Evidence is ranked from exact skill overlap, shared role terms, indexed search, and verification state.",
		len(analysis.MatchedSkills), len(normalizedProfileSkills),
	)

	jobTerms := significantTerms(description)
	indexedScores := make(map[string]int, len(searchHits))
	for _, hit := range searchHits {
		if hit.Score > indexedScores[hit.FactID] {
			indexedScores[hit.FactID] = hit.Score
		}
	}
	for _, experience := range experiences {
		label := strings.TrimSpace(experience.Title + " · " + experience.Company)
		for _, bullet := range experience.Bullets {
			candidate := rankEvidence(bullet.ID, experience.ID, "experience", label, bullet.Text, bullet.Verification == VerificationVerified, false, nil, analysis.DetectedSkills, jobTerms, indexedScores[bullet.ID])
			if candidate.Score > 0 {
				analysis.RankedEvidence = append(analysis.RankedEvidence, candidate)
			}
		}
	}
	for _, project := range projects {
		projectSkills := make([]string, 0)
		for _, skill := range appendProjectLanguages(project.Skills, project.DetectedLanguages) {
			if mentionsSkill(description, skill) {
				projectSkills = appendUniqueFold(projectSkills, skill)
			}
		}
		if len(project.Bullets) == 0 {
			text := project.Description
			if text == "" {
				text = project.Name
			}
			candidate := rankEvidence(project.ID, project.ID, "project", project.Name, text, project.Verification == VerificationVerified, project.ResumeEligible, projectSkills, analysis.DetectedSkills, jobTerms, indexedScores[project.ID])
			if candidate.Score > 0 {
				analysis.RankedEvidence = append(analysis.RankedEvidence, candidate)
			}
		}
		for _, bullet := range project.Bullets {
			candidate := rankEvidence(bullet.ID, project.ID, "project", project.Name, bullet.Text, bullet.Verification == VerificationVerified, project.ResumeEligible, projectSkills, analysis.DetectedSkills, jobTerms, indexedScores[bullet.ID])
			if candidate.Score > 0 {
				analysis.RankedEvidence = append(analysis.RankedEvidence, candidate)
			}
		}
	}

	sort.SliceStable(analysis.RankedEvidence, func(left, right int) bool {
		if analysis.RankedEvidence[left].Score != analysis.RankedEvidence[right].Score {
			return analysis.RankedEvidence[left].Score > analysis.RankedEvidence[right].Score
		}
		if analysis.RankedEvidence[left].SourceLabel != analysis.RankedEvidence[right].SourceLabel {
			return analysis.RankedEvidence[left].SourceLabel < analysis.RankedEvidence[right].SourceLabel
		}
		return analysis.RankedEvidence[left].FactID < analysis.RankedEvidence[right].FactID
	})
	if len(analysis.RankedEvidence) > maxRankedEvidence {
		analysis.RankedEvidence = analysis.RankedEvidence[:maxRankedEvidence]
	}
	return analysis, nil
}

func rankEvidence(factID, sourceID, sourceType, label, text string, verified, eligible bool, sourceSkills, detectedSkills []string, jobTerms map[string]struct{}, indexedScore int) EvidenceMatch {
	matchedSkills := append([]string(nil), sourceSkills...)
	for _, skill := range detectedSkills {
		if mentionsSkill(text+" "+label, skill) {
			matchedSkills = appendUniqueFold(matchedSkills, skill)
		}
	}
	sort.Strings(matchedSkills)
	overlap := termOverlap(jobTerms, significantTerms(text))
	if len(matchedSkills) == 0 && overlap == 0 && indexedScore == 0 {
		return EvidenceMatch{}
	}

	score := min(len(matchedSkills)*22, 60) + min(overlap*4, 24) + min(indexedScore, 18)
	reasons := make([]string, 0, 4)
	if len(matchedSkills) > 0 {
		reasons = append(reasons, "Matches "+strings.Join(matchedSkills, ", "))
	}
	if overlap > 0 {
		reasons = append(reasons, fmt.Sprintf("Shares %d meaningful role terms", overlap))
	}
	if indexedScore > 0 {
		reasons = append(reasons, "Indexed evidence search match")
	}
	if verified {
		score += 10
		reasons = append(reasons, "Evidence is verified")
	}
	if eligible {
		score += 6
		reasons = append(reasons, "Project is resume eligible")
	}
	return EvidenceMatch{
		FactID: factID, SourceID: sourceID, SourceType: sourceType, SourceLabel: label,
		Text: text, Score: min(score, 100), MatchedSkills: matchedSkills, Reasons: reasons, Verified: verified, Selectable: sourceType != "project" || eligible,
	}
}

func appendProjectLanguages(skills []string, languages []RepositoryLanguage) []string {
	result := append([]string(nil), skills...)
	for _, language := range languages {
		result = append(result, language.Name)
	}
	return normalizeMatchingSkills(result)
}

var skillAliasGroups = [][]string{
	{"go", "golang"},
	{"javascript", "js", "ecmascript"},
	{"typescript", "ts"},
	{"node.js", "nodejs"},
	{"react", "react.js", "reactjs"},
	{"vue", "vue.js", "vuejs"},
	{"postgresql", "postgres", "psql"},
	{"kubernetes", "k8s"},
	{"amazon web services", "aws"},
	{"google cloud platform", "gcp"},
	{"continuous integration", "ci/cd", "ci cd"},
	{"c#", "c sharp"},
	{"c++", "c plus plus"},
	{".net", "dotnet"},
}

var recognizedSkillCatalog = []string{
	"Go", "Rust", "Java", "Python", "C#", "C++", "JavaScript", "TypeScript", "React", "Angular", "Vue", "Node.js", ".NET", "Spring", "Django", "Flask", "FastAPI", "Ruby", "Rails", "PHP", "Laravel", "Swift", "Kotlin",
	"SQL", "PostgreSQL", "MySQL", "SQLite", "MongoDB", "Redis", "Kafka", "Elasticsearch", "Docker", "Kubernetes", "Terraform", "AWS", "GCP", "Azure", "Linux", "Git", "GitHub Actions", "Jenkins", "GraphQL", "REST", "gRPC", "HTML", "CSS", "Tailwind CSS", "Next.js", "Svelte", "Prometheus", "Grafana", "Spark",
}

// DetectKnownSkills returns the technology names from TailorCV's matching
// catalog that are explicitly present in text. AI response validation uses the
// same vocabulary as deterministic matching so providers cannot introduce a
// recognized technology that is absent from their cited evidence.
func DetectKnownSkills(text string) []string {
	result := make([]string, 0)
	for _, skill := range normalizeMatchingSkills(recognizedSkillCatalog) {
		if mentionsSkill(text, skill) {
			result = appendUniqueFold(result, skill)
		}
	}
	sort.Strings(result)
	return result
}

func mentionsSkill(description, skill string) bool {
	for _, alias := range aliasesFor(skill) {
		if alias == "go" {
			pattern := `(^|[^a-zA-Z0-9])(Go|GO)([^a-zA-Z0-9]|$)`
			if regexp.MustCompile(pattern).MatchString(description) {
				return true
			}
			continue
		}
		pattern := `(?i)(^|[^a-z0-9])` + regexp.QuoteMeta(alias) + `([^a-z0-9]|$)`
		if regexp.MustCompile(pattern).MatchString(description) {
			return true
		}
	}
	return false
}

func aliasesFor(skill string) []string {
	key := strings.ToLower(strings.TrimSpace(skill))
	for _, group := range skillAliasGroups {
		for _, alias := range group {
			if key == alias {
				return group
			}
		}
	}
	return []string{key}
}

func normalizeMatchingSkills(skills []string) []string {
	result := make([]string, 0, len(skills))
	seen := make(map[string]struct{}, len(skills))
	for _, skill := range normalizeSkills(skills) {
		canonical := strings.ToLower(skill)
		for _, group := range skillAliasGroups {
			for _, alias := range group {
				if canonical == alias {
					canonical = group[0]
					break
				}
			}
		}
		if _, exists := seen[canonical]; exists {
			continue
		}
		seen[canonical] = struct{}{}
		result = append(result, skill)
	}
	return result
}

var preferredMarkers = []string{"preferred", "nice to have", "nice-to-have", "bonus", "a plus", "advantageous", "ideally"}

func skillIsPreferred(description, skill string) bool {
	preferred := false
	for _, segment := range splitRequirementSegments(description) {
		if !mentionsSkill(segment, skill) {
			continue
		}
		lower := strings.ToLower(segment)
		segmentPreferred := false
		for _, marker := range preferredMarkers {
			if strings.Contains(lower, marker) {
				segmentPreferred = true
				preferred = true
				break
			}
		}
		if !segmentPreferred {
			return false
		}
	}
	return preferred
}

var responsibilityVerb = regexp.MustCompile(`(?i)(^|[^a-z])(build|builds|design|designs|develop|develops|deliver|delivers|implement|implements|improve|improves|lead|leads|maintain|maintains|manage|manages|operate|operates|own|owns|create|creates|collaborate|collaborates|optimize|optimizes|support|supports)([^a-z]|$)`)

func extractResponsibilities(description string) []string {
	result := make([]string, 0, 6)
	seen := make(map[string]struct{})
	for _, segment := range splitRequirementSegments(description) {
		segment = strings.TrimSpace(strings.TrimLeft(segment, "-•*0123456789.) "))
		segment = strings.Join(strings.Fields(segment), " ")
		if len(segment) < 20 || len(segment) > 320 || !responsibilityVerb.MatchString(segment) {
			continue
		}
		key := strings.ToLower(segment)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, segment)
		if len(result) == 6 {
			break
		}
	}
	return result
}

var requirementSeparator = regexp.MustCompile(`[\r\n;!?]+|\.\s+`)

func splitRequirementSegments(description string) []string {
	return requirementSeparator.Split(description, -1)
}

func extractKeywords(description string, limit int) []string {
	counts := make(map[string]int)
	fields := strings.FieldsFunc(strings.ToLower(description), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	for _, field := range fields {
		if len(field) < 3 {
			continue
		}
		if _, ignored := ignoredTerms[field]; ignored {
			continue
		}
		counts[field]++
	}
	keywords := make([]string, 0, len(counts))
	for keyword := range counts {
		keywords = append(keywords, keyword)
	}
	sort.Slice(keywords, func(left, right int) bool {
		if counts[keywords[left]] != counts[keywords[right]] {
			return counts[keywords[left]] > counts[keywords[right]]
		}
		return keywords[left] < keywords[right]
	})
	if len(keywords) > limit {
		keywords = keywords[:limit]
	}
	return keywords
}

var ignoredTerms = map[string]struct{}{
	"and": {}, "are": {}, "but": {}, "for": {}, "from": {}, "have": {}, "into": {}, "our": {}, "that": {}, "the": {}, "their": {}, "this": {}, "through": {}, "using": {}, "was": {}, "will": {}, "with": {}, "you": {}, "your": {},
	"ability": {}, "about": {}, "candidate": {}, "experience": {}, "including": {}, "looking": {}, "preferred": {}, "required": {}, "requirements": {}, "responsibilities": {}, "role": {}, "team": {}, "work": {}, "years": {},
}

func significantTerms(value string) map[string]struct{} {
	terms := make(map[string]struct{})
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	for _, field := range fields {
		if len(field) < 3 {
			continue
		}
		if _, ignored := ignoredTerms[field]; ignored {
			continue
		}
		terms[field] = struct{}{}
	}
	return terms
}

func termOverlap(left, right map[string]struct{}) int {
	count := 0
	for term := range left {
		if _, found := right[term]; found {
			count++
		}
	}
	return count
}

func appendUniqueFold(values []string, value string) []string {
	for _, existing := range values {
		if strings.EqualFold(existing, value) {
			return values
		}
	}
	return append(values, value)
}
