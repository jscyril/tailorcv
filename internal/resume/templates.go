package resume

import "github.com/jscyril/tailorcv/internal/domain"

const (
	DefaultTemplateID = "builtin-jake-style"
	ClassicTemplateID = "builtin-classic-ats"
)

const jakeStyleSource = `% Jake-style ATS resume for TailorCV
% Inspired by Jake Gutierrez's MIT-licensed resume layout.
\documentclass[letterpaper,11pt]{article}
\usepackage[empty]{fullpage}
\usepackage{titlesec}
\usepackage{enumitem}
\usepackage[hidelinks]{hyperref}
\usepackage{fancyhdr}
\pagestyle{fancy}
\fancyhf{}
\renewcommand{\headrulewidth}{0pt}
\renewcommand{\footrulewidth}{0pt}
\addtolength{\oddsidemargin}{-0.55in}
\addtolength{\textwidth}{1.1in}
\addtolength{\topmargin}{-0.65in}
\addtolength{\textheight}{1.25in}
\titleformat{\section}{\vspace{-4pt}\scshape\raggedright\large\bfseries}{}{0em}{}[\titlerule\vspace{-5pt}]
\setlist[itemize]{leftmargin=0.18in,itemsep=1pt,topsep=2pt}
\setlength{\parindent}{0pt}
\urlstyle{same}

\begin{document}
\begin{center}
  {\Huge\scshape {{TAILORCV_NAME}}}\\[3pt]
  {\large {{TAILORCV_HEADLINE}}}\\[3pt]
  {\small {{TAILORCV_CONTACT}}}
\end{center}

{{TAILORCV_SUMMARY_SECTION}}
{{TAILORCV_EXPERIENCE_SECTION}}
{{TAILORCV_PROJECTS_SECTION}}
{{TAILORCV_EDUCATION_SECTION}}
{{TAILORCV_CERTIFICATIONS_SECTION}}
{{TAILORCV_ACHIEVEMENTS_SECTION}}
{{TAILORCV_SKILLS_SECTION}}
\end{document}`

const classicSource = `% Classic ATS template for TailorCV
\documentclass[10pt,letterpaper]{article}
\usepackage[margin=0.62in]{geometry}
\usepackage[hidelinks]{hyperref}
\usepackage{enumitem}
\setlength{\parindent}{0pt}
\setlist[itemize]{leftmargin=0.2in,itemsep=1pt,topsep=2pt}
\begin{document}
\begin{center}
  {\LARGE\textbf{{{TAILORCV_NAME}}}}\\[2pt]
  {{TAILORCV_HEADLINE}}\\[2pt]
  {\small {{TAILORCV_CONTACT}}}
\end{center}
\hrule
\vspace{6pt}

{{TAILORCV_SUMMARY_SECTION}}
{{TAILORCV_EXPERIENCE_SECTION}}
{{TAILORCV_PROJECTS_SECTION}}
{{TAILORCV_EDUCATION_SECTION}}
{{TAILORCV_CERTIFICATIONS_SECTION}}
{{TAILORCV_ACHIEVEMENTS_SECTION}}
{{TAILORCV_SKILLS_SECTION}}
\end{document}`

func BuiltinTemplates() []domain.ResumeTemplate {
	return []domain.ResumeTemplate{
		{ID: DefaultTemplateID, Name: "Jake-style ATS", Description: "A familiar single-column software resume layout inspired by Jake's Resume.", Source: jakeStyleSource, BuiltIn: true},
		{ID: ClassicTemplateID, Name: "Classic ATS", Description: "A restrained one-page article layout with minimal dependencies.", Source: classicSource, BuiltIn: true},
	}
}

func FindBuiltinTemplate(id string) (domain.ResumeTemplate, bool) {
	for _, template := range BuiltinTemplates() {
		if template.ID == id {
			return template, true
		}
	}
	return domain.ResumeTemplate{}, false
}
