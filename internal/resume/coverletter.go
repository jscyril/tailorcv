package resume

import (
	"strings"

	"github.com/jscyril/tailorcv/internal/domain"
)

const coverLetterSource = `% TailorCV cover letter
\documentclass[11pt,letterpaper]{article}
\usepackage[margin=1in]{geometry}
\usepackage[hidelinks]{hyperref}
\usepackage{enumitem}
\setlength{\parindent}{0pt}
\setlength{\parskip}{10pt}
\setlist[itemize]{leftmargin=0.2in,itemsep=2pt,topsep=0pt}
\pagestyle{empty}
\begin{document}
{\Large\textbf{{{TAILORCV_NAME}}}}\\
{\small {{TAILORCV_CONTACT}}}

{{TAILORCV_LETTER_BODY}}
\end{document}
`

// RenderCoverLetter turns plain letter text into LaTeX. Blank lines separate
// paragraphs, lines starting with "- " become a list, and other line breaks
// are kept. All text is escaped.
func RenderCoverLetter(profile domain.Profile, text string) string {
	var body strings.Builder
	for _, paragraph := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n") {
		lines := strings.Split(strings.TrimSpace(paragraph), "\n")
		if lines[0] == "" {
			continue
		}
		inList := false
		var rendered []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			item, isItem := strings.CutPrefix(line, "- ")
			switch {
			case isItem && !inList:
				rendered = append(rendered, "\\begin{itemize}", "  \\item "+escape(item))
				inList = true
			case isItem:
				rendered = append(rendered, "  \\item "+escape(item))
			default:
				if inList {
					rendered = append(rendered, "\\end{itemize}")
					inList = false
				}
				if len(rendered) > 0 && !strings.HasPrefix(rendered[len(rendered)-1], "\\end{itemize}") {
					rendered[len(rendered)-1] += "\\\\"
				}
				rendered = append(rendered, escape(line))
			}
		}
		if inList {
			rendered = append(rendered, "\\end{itemize}")
		}
		body.WriteString(strings.Join(rendered, "\n") + "\n\n")
	}
	source := strings.ReplaceAll(coverLetterSource, "{{TAILORCV_LETTER_BODY}}", strings.TrimSpace(body.String()))
	return Render(source, Data{Profile: profile})
}
