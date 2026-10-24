package resume

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/atomicfile"
)

const (
	// onlineBundleURL must match bundleSourceURL in cmd/packagetectonic so
	// fetched files come from the same TeX Live release as the offline bundle.
	onlineBundleURL = "https://data1.fullyjustified.net/tlextras-2022.0r0.tar"
	maxFetchRounds  = 30
	maxFetchedBytes = 32 << 20
	fetchTimeout    = 60 * time.Second
)

var (
	missingFilePattern = regexp.MustCompile("(?:File|find file) `([^']+)'")
	// Matches font errors such as "Font TU/lmtt/m/n/10=[lmmono10-regular]: at 10.0pt not loadable".
	missingFontPattern = regexp.MustCompile(`=\[([^\]:/]+)(?::[^\]]*)?\][^\n]*not loadable`)
	// Tectonic's own messages, e.g. `physical font "cmr10.pfb"` from the PDF
	// writer and `failed to open input file "xstring.tex"`.
	missingQuotedPattern = regexp.MustCompile(`(?:physical font|open input file) "([^"]+)"`)
	missingTFMPattern    = regexp.MustCompile(`Font \\[^=\s]+=([A-Za-z0-9._-]+)(?: at [^\n]*)? not loadable: Metric \(TFM\)`)
	bundleFileName       = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)
)

func defaultExtrasDirectory() string {
	cache, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cache, "tailorcv", "tex-extras")
}

// missingBundleFiles lists the files a failed compile could not find. Fonts
// named without an extension are tried as OpenType, then TrueType.
func missingBundleFiles(log string) []string {
	names := make([]string, 0)
	add := func(name string) {
		if bundleFileName.MatchString(name) && !strings.Contains(name, "..") && !strings.HasPrefix(name, "resume.") && !containsName(names, name) {
			names = append(names, name)
		}
	}
	for _, match := range missingFilePattern.FindAllStringSubmatch(log, -1) {
		add(match[1])
	}
	for _, match := range missingFontPattern.FindAllStringSubmatch(log, -1) {
		if filepath.Ext(match[1]) != "" {
			add(match[1])
		} else {
			add(match[1] + ".otf")
			add(match[1] + ".ttf")
		}
	}
	for _, match := range missingTFMPattern.FindAllStringSubmatch(log, -1) {
		add(match[1] + ".tfm")
	}
	for _, match := range missingQuotedPattern.FindAllStringSubmatch(log, -1) {
		add(match[1])
	}
	return names
}

func containsName(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}

// fetchBundleFile downloads one named file from the online bundle into extras.
// It runs Tectonic's bundle reader, which never processes the user's LaTeX.
func fetchBundleFile(ctx context.Context, executable, extras, name string) error {
	if !bundleFileName.MatchString(name) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid bundle file name %q", name)
	}
	project, err := os.MkdirTemp("", "tailorcv-fetch-*")
	if err != nil {
		return fmt.Errorf("create fetch folder: %w", err)
	}
	defer os.RemoveAll(project)
	manifest := "[doc]\nname = \"fetch\"\nbundle = \"" + onlineBundleURL + "\"\n\n[[output]]\nname = \"default\"\ntype = \"pdf\"\n"
	if err := os.WriteFile(filepath.Join(project, "Tectonic.toml"), []byte(manifest), 0o600); err != nil {
		return fmt.Errorf("write fetch manifest: %w", err)
	}
	fetchContext, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	command := exec.CommandContext(fetchContext, executable, "-X", "bundle", "cat", name)
	command.Dir = project
	command.Env = withEnvironment(os.Environ(), "TECTONIC_CACHE_DIR", filepath.Join(filepath.Dir(extras), "tectonic-cache"))
	stdout := &cappedBuffer{limit: maxFetchedBytes}
	stderr := &cappedBuffer{limit: 4096}
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if lines := strings.Split(message, "\n"); len(lines) > 0 {
			message = strings.TrimSpace(strings.TrimPrefix(lines[len(lines)-1], "error:"))
		}
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("%s: %s", name, message)
	}
	data := []byte(stdout.String())
	if len(data) == 0 || stdout.truncated {
		return fmt.Errorf("%s: download was empty or too large", name)
	}
	if err := os.MkdirAll(extras, 0o700); err != nil {
		return fmt.Errorf("create TeX extras folder: %w", err)
	}
	return atomicfile.Write(filepath.Join(extras, name), data)
}

// copyExtras puts previously fetched files beside the source, where TeX
// looks before the bundle.
func copyExtras(extras, workspace string) error {
	if extras == "" {
		return nil
	}
	entries, err := os.ReadDir(extras)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read TeX extras folder: %w", err)
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !bundleFileName.MatchString(entry.Name()) || strings.HasPrefix(entry.Name(), "resume.") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(extras, entry.Name()))
		if err != nil {
			return fmt.Errorf("read TeX extra %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(workspace, entry.Name()), data, 0o600); err != nil {
			return fmt.Errorf("copy TeX extra %s: %w", entry.Name(), err)
		}
	}
	return nil
}
