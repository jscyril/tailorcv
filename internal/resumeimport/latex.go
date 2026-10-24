// Package resumeimport reads an existing LaTeX resume into profile data.
package resumeimport

import (
	"regexp"
	"strings"
	"unicode"
)

// stripComments removes LaTeX comments (an unescaped % to the end of a line).
func stripComments(source string) string {
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	for index, line := range lines {
		for position := 0; position < len(line); position++ {
			if line[position] == '\\' {
				position++
				continue
			}
			if line[position] == '%' {
				line = line[:position]
				break
			}
		}
		lines[index] = line
	}
	return strings.Join(lines, "\n")
}

// readGroup reads one {...} argument starting at or after position, skipping
// whitespace. It returns the inner text and the index after the closing brace,
// or ok=false when no group starts there.
func readGroup(source string, position int) (string, int, bool) {
	for position < len(source) && unicode.IsSpace(rune(source[position])) {
		position++
	}
	if position >= len(source) || source[position] != '{' {
		return "", position, false
	}
	depth := 0
	for index := position; index < len(source); index++ {
		switch source[index] {
		case '\\':
			index++
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[position+1 : index], index + 1, true
			}
		}
	}
	return "", position, false
}

// readGroups reads count consecutive {...} arguments.
func readGroups(source string, position, count int) ([]string, int, bool) {
	groups := make([]string, 0, count)
	for range count {
		group, next, ok := readGroup(source, position)
		if !ok {
			return nil, position, false
		}
		groups = append(groups, group)
		position = next
	}
	return groups, position, true
}

// commandIndexes finds every \name that is not followed by another letter.
func commandIndexes(source, name string) []int {
	pattern := regexp.MustCompile(`\\` + regexp.QuoteMeta(name) + `(?:[^A-Za-z]|$)`)
	indexes := []int{}
	for _, match := range pattern.FindAllStringIndex(source, -1) {
		indexes = append(indexes, match[0])
	}
	return indexes
}

// Commands whose argument is dropped (spacing, labels) and commands whose
// argument is kept as text (formatting).
var (
	droppedArgument = map[string]bool{"vspace": true, "hspace": true, "label": true, "color": true, "includegraphics": true, "faIcon": true}
	escapedSymbols  = map[byte]string{'&': "&", '%': "%", '$': "$", '#': "#", '_': "_", '{': "{", '}': "}", ' ': " ", ',': " "}
	namedSymbols    = map[string]string{"textasciitilde": "~", "textbar": "|", "textbullet": "•", "vert": "|", "cdot": "·", "bullet": "•", "ldots": "...", "LaTeX": "LaTeX", "TeX": "TeX", "faPhone": "", "faEnvelope": "", "faLinkedin": "", "faGithub": ""}
	spaceRun        = regexp.MustCompile(`[ \t]+`)
)

// toText turns a LaTeX fragment into plain text: formatting is removed, link
// text is kept, escapes are undone, and line breaks (\\) become newlines.
func toText(source string) string {
	var output strings.Builder
	for index := 0; index < len(source); index++ {
		character := source[index]
		switch {
		case character == '\\' && index+1 < len(source) && source[index+1] == '\\':
			output.WriteByte('\n')
			index++
			// Skip an optional [length] after \\.
			if rest := source[index+1:]; strings.HasPrefix(strings.TrimLeft(rest, " "), "[") {
				if end := strings.IndexByte(rest, ']'); end >= 0 {
					index += end + 1
				}
			}
		case character == '\\' && index+1 < len(source) && !isLetter(source[index+1]):
			output.WriteString(escapedSymbols[source[index+1]])
			index++
		case character == '\\':
			end := index + 1
			for end < len(source) && isLetter(source[end]) {
				end++
			}
			name := source[index+1 : end]
			index = end - 1
			if symbol, ok := namedSymbols[name]; ok {
				output.WriteString(symbol)
				continue
			}
			switch {
			case name == "href":
				if groups, next, ok := readGroups(source, end, 2); ok {
					output.WriteString(toText(groups[1]))
					index = next - 1
				}
			case name == "textcolor":
				if groups, next, ok := readGroups(source, end, 2); ok {
					output.WriteString(toText(groups[1]))
					index = next - 1
				}
			case name == "begin" || name == "end":
				// Drop the environment name and any [options] after it.
				if _, next, ok := readGroup(source, end); ok {
					index = next - 1
					if rest := source[next:]; strings.HasPrefix(rest, "[") {
						if close := strings.IndexByte(rest, ']'); close >= 0 {
							index = next + close
						}
					}
				}
			case droppedArgument[name]:
				if _, next, ok := readGroup(source, end); ok {
					index = next - 1
				}
			}
			// Any other command is dropped; its {argument}, if any, stays as text.
		case character == '{' || character == '}' || character == '$':
		case character == '~':
			output.WriteByte(' ')
		default:
			output.WriteByte(character)
		}
	}
	text := strings.NewReplacer("---", "—", "--", "–").Replace(output.String())
	lines := strings.Split(text, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimSpace(spaceRun.ReplaceAllString(line, " "))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func isLetter(character byte) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
}

// hrefTargets returns the URL of every \href in a fragment.
func hrefTargets(source string) []string {
	urls := []string{}
	for _, index := range commandIndexes(source, "href") {
		if url, _, ok := readGroup(source, index+len(`\href`)); ok {
			urls = append(urls, strings.TrimSpace(strings.ReplaceAll(url, `\`, "")))
		}
	}
	return urls
}

// PlainText is the readable text of a LaTeX document, used to check that
// AI-extracted entries appear in the original.
func PlainText(source string) string {
	return toText(stripComments(source))
}
