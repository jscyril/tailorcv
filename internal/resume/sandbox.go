package resume

import (
	"fmt"
	"regexp"
	"strings"
)

// sandboxPolicy lists the only files a compiler process may touch: it can
// read and write inside workspace, and only read the executable and bundle.
type sandboxPolicy struct {
	workspace  string
	executable string
	bundle     string
}

// Without an OS sandbox, TeX can read and write any file the user can. These
// control sequences open files directly, or let a template build or rename
// one that does (for example \csname input\endcsname or \begin{input}).
//
// shortcut: a denylist cannot cover every TeX trick, so it only applies where
// no OS sandbox exists (Windows, Linux without Landlock); upgrade by adding a
// Windows sandbox (AppContainer or low-integrity token) and dropping this.
var blockedControlSequences = []string{
	"input", "include", "includeonly", "InputIfFileExists", "IfFileExists",
	"includegraphics", "includepdf", "lstinputlisting", "verbatiminput",
	"openin", "openout", "read", "readline", "write18", "immediate",
	"XeTeXpicfile", "XeTeXpdffile", "special",
	"csname", "ifcsname", "UseName", "ExpandArgs", "NewCommandCopy",
	"csuse", "csdef", "csgdef", "csedef", "csxdef", "cslet", "csletcs", "letcs",
	"catcode", "scantokens", "ExplSyntaxOn", "makeatletter", "endlinechar",
}

var (
	controlWordPattern   = regexp.MustCompile(`\\([A-Za-z]+)`)
	environmentPattern   = regexp.MustCompile(`\\(?:begin|end)\s*\{\s*([^}]*)\}`)
	packagePathPattern   = regexp.MustCompile(`\\(?:documentclass|usepackage|RequirePackage|LoadClass)\s*(?:\[[^\]]*\])?\s*\{([^}]*)\}`)
	absoluteOrParentPath = regexp.MustCompile(`(^|,)\s*([/\\~]|[A-Za-z]:|.*\.\.)`)
)

// unsafeSourceReason explains why source is refused when no OS sandbox is
// available, or returns "" when it may be compiled.
func unsafeSourceReason(source, platform string) string {
	blocked := make(map[string]bool, len(blockedControlSequences))
	for _, name := range blockedControlSequences {
		blocked[name] = true
	}
	refuse := func(what string) string {
		return fmt.Sprintf("This LaTeX uses %s, which TailorCV blocks on %s because it can read or write files outside the resume. Remove it and compile again.", what, platform)
	}
	if strings.Contains(source, "^^") {
		return refuse("^^ character codes")
	}
	for _, match := range controlWordPattern.FindAllStringSubmatch(source, -1) {
		if blocked[match[1]] {
			return refuse(`\` + match[1])
		}
	}
	for _, match := range environmentPattern.FindAllStringSubmatch(source, -1) {
		if name := strings.TrimSpace(match[1]); blocked[name] {
			return refuse(`the "` + name + `" environment`)
		}
	}
	for _, match := range packagePathPattern.FindAllStringSubmatch(source, -1) {
		if absoluteOrParentPath.MatchString(match[1]) {
			return refuse("a package or class path outside the bundle (" + strings.TrimSpace(match[1]) + ")")
		}
	}
	return ""
}
