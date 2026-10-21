package domain

import "regexp"

// SkillCoverage says whether one skill the job asks for appears in a resume.
type SkillCoverage struct {
	Skill   string `json:"skill"`
	Present bool   `json:"present"`
}

// KeywordCoverage lists the job's required and preferred skills and whether
// each appears in the resume text.
type KeywordCoverage struct {
	Required  []SkillCoverage `json:"required"`
	Preferred []SkillCoverage `json:"preferred"`
}

// LaTeX escapes such as C\# or R\&D would otherwise hide a skill from matching.
var latexEscapedCharacter = regexp.MustCompile(`\\([#$%&_{}])`)

// CheckKeywordCoverage reuses job analysis to find the job's skills, then
// looks for each one (with its aliases) in the resume source.
func CheckKeywordCoverage(job JobAnalysisInput, profileSkills []string, projects []Project, resumeSource string) (KeywordCoverage, error) {
	analysis, err := AnalyzeCareerEvidence(job, profileSkills, nil, projects)
	if err != nil {
		return KeywordCoverage{}, err
	}
	text := latexEscapedCharacter.ReplaceAllString(resumeSource, "$1")
	check := func(skills []string) []SkillCoverage {
		result := make([]SkillCoverage, 0, len(skills))
		for _, skill := range skills {
			result = append(result, SkillCoverage{Skill: skill, Present: mentionsSkill(text, skill)})
		}
		return result
	}
	return KeywordCoverage{Required: check(analysis.RequiredSkills), Preferred: check(analysis.PreferredSkills)}, nil
}
