package resume

import "github.com/jscyril/tailorcv/internal/domain"

const (
	DefaultTemplateID = "builtin-jake-style"
	ClassicTemplateID = "builtin-classic-ats"
)

const jakeStyleSource = `%-------------------------
% Jake's Resume for TailorCV
% Layout: Jake Gutierrez's resume (MIT license), based on sb2nov/resume.
% TailorCV fills the sections from your profile through the \tailorcv
% commands below, which map onto Jake's own commands.
%------------------------
\documentclass[letterpaper,11pt]{article}

\usepackage{latexsym}
\usepackage[empty]{fullpage}
\usepackage{titlesec}
\usepackage{marvosym}
\usepackage[usenames,dvipsnames]{color}
\usepackage{verbatim}
\usepackage{enumitem}
\usepackage[hidelinks]{hyperref}
\usepackage{fancyhdr}
\usepackage[english]{babel}
\usepackage{tabularx}

\pagestyle{fancy}
\fancyhf{}
\fancyfoot{}
\renewcommand{\headrulewidth}{0pt}
\renewcommand{\footrulewidth}{0pt}

\addtolength{\oddsidemargin}{-0.5in}
\addtolength{\evensidemargin}{-0.5in}
\addtolength{\textwidth}{1in}
\addtolength{\topmargin}{-.5in}
\addtolength{\textheight}{1.0in}

\urlstyle{same}
\raggedbottom
\raggedright
\setlength{\tabcolsep}{0in}

\titleformat{\section}{
  \vspace{-4pt}\scshape\raggedright\large
}{}{0em}{}[\color{black}\titlerule \vspace{-5pt}]

\newcommand{\resumeItem}[1]{\item\small{{#1 \vspace{-2pt}}}}
\newcommand{\resumeSubheading}[4]{
  \vspace{-2pt}\item
    \begin{tabular*}{0.97\textwidth}[t]{l@{\extracolsep{\fill}}r}
      \textbf{#1} & #2 \\
      \textit{\small#3} & \textit{\small #4} \\
    \end{tabular*}\vspace{-7pt}
}
\newcommand{\resumeProjectHeading}[2]{
    \item
    \begin{tabular*}{0.97\textwidth}{l@{\extracolsep{\fill}}r}
      \small#1 & #2 \\
    \end{tabular*}\vspace{-7pt}
}
\newcommand{\resumeSubItem}[1]{\resumeItem{#1}\vspace{-4pt}}
\renewcommand\labelitemii{$\vcenter{\hbox{\tiny$\bullet$}}$}
\newcommand{\resumeSubHeadingListStart}{\begin{itemize}[leftmargin=0.15in, label={}]}
\newcommand{\resumeSubHeadingListEnd}{\end{itemize}}
\newcommand{\resumeItemListStart}{\begin{itemize}}
\newcommand{\resumeItemListEnd}{\end{itemize}\vspace{-5pt}}

% TailorCV section commands, written with Jake's commands.
\newcommand{\tailorcvListStart}{\resumeSubHeadingListStart}
\newcommand{\tailorcvListEnd}{\resumeSubHeadingListEnd}
\newcommand{\tailorcvEntry}[4]{\resumeSubheading{#1}{#2}{#3}{#4}}
\newcommand{\tailorcvProject}[3]{\resumeProjectHeading{\textbf{#1}\if\relax\detokenize{#2}\relax\else{} $|$ \emph{#2}\fi}{#3}}
\newcommand{\tailorcvBulletsStart}{\resumeItemListStart}
\newcommand{\tailorcvBulletsEnd}{\resumeItemListEnd}
\newcommand{\tailorcvBullet}[1]{\resumeItem{#1}}
\newcommand{\tailorcvSkills}[1]{\begin{itemize}[leftmargin=0.15in, label={}]\small{\item{#1}}\end{itemize}}

\begin{document}

\begin{center}
    \textbf{\Huge \scshape {{TAILORCV_NAME}}} \\ \vspace{1pt}
    {{TAILORCV_HEADLINE_LINE}}
    \small {{TAILORCV_CONTACT_LINKS}}
\end{center}

{{TAILORCV_SUMMARY_SECTION}}
{{TAILORCV_EDUCATION_SECTION}}
{{TAILORCV_EXPERIENCE_SECTION}}
{{TAILORCV_PROJECTS_SECTION}}
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
% Resumes need neither section numbers nor a page number.
\setcounter{secnumdepth}{0}
\pagestyle{empty}
\begin{document}
\begin{center}
  {\LARGE\textbf{{{TAILORCV_NAME}}}}\\[2pt]
  {{TAILORCV_HEADLINE_LINE}}
  {\small {{TAILORCV_CONTACT_LINKS}}}
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
		{ID: DefaultTemplateID, Name: "Jake's Resume", Description: "Jake Gutierrez's single-column software resume layout (MIT license), filled from your profile.", Source: jakeStyleSource, BuiltIn: true},
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
