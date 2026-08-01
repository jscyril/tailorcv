package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type JobAnalysisInput struct {
	Description string `json:"description"`
}

type JobAnalysis struct {
	Score             int      `json:"score"`
	MatchedSkills     []string `json:"matchedSkills"`
	UnmentionedSkills []string `json:"unmentionedSkills"`
	Explanation       string   `json:"explanation"`
}

func AnalyzeJobDescription(input JobAnalysisInput, profileSkills []string) (JobAnalysis, error) {
	description := strings.TrimSpace(input.Description)
	if len(description) < 40 {
		return JobAnalysis{}, fmt.Errorf("job description must contain at least 40 characters")
	}

	analysis := JobAnalysis{
		MatchedSkills:     make([]string, 0),
		UnmentionedSkills: make([]string, 0),
	}
	normalizedSkills := normalizeSkills(profileSkills)
	for _, skill := range normalizedSkills {
		if mentionsSkill(description, skill) {
			analysis.MatchedSkills = append(analysis.MatchedSkills, skill)
		} else {
			analysis.UnmentionedSkills = append(analysis.UnmentionedSkills, skill)
		}
	}

	sort.Strings(analysis.MatchedSkills)
	sort.Strings(analysis.UnmentionedSkills)
	if len(normalizedSkills) > 0 {
		analysis.Score = int(float64(len(analysis.MatchedSkills))/float64(len(normalizedSkills))*100 + 0.5)
	}
	analysis.Explanation = fmt.Sprintf(
		"%d of %d profile skills are explicitly mentioned in this job description. This is a direct text comparison, not an overall suitability score.",
		len(analysis.MatchedSkills),
		len(normalizedSkills),
	)
	return analysis, nil
}

func mentionsSkill(description, skill string) bool {
	pattern := `(?i)(^|[^a-z0-9])` + regexp.QuoteMeta(skill) + `([^a-z0-9]|$)`
	return regexp.MustCompile(pattern).MatchString(description)
}
