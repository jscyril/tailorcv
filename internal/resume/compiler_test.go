package resume

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestResolveTectonicExecutableUsesConfiguredPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tectonic-test")
	if err := os.WriteFile(path, []byte("test"), 0o700); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("TAILORCV_TECTONIC", path)
	resolved, err := resolveTectonicExecutable()
	if err != nil {
		t.Fatalf("resolveTectonicExecutable() error = %v", err)
	}
	if resolved != path {
		t.Fatalf("resolveTectonicExecutable() = %q", resolved)
	}
}

func TestResolveTectonicBundleUsesConfiguredLocalFile(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), bundleFilename)
	if err := os.WriteFile(bundle, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("TAILORCV_TECTONIC_BUNDLE", bundle)

	resolved, err := resolveTectonicBundle(filepath.Join(t.TempDir(), "tectonic"))
	if err != nil {
		t.Fatalf("resolveTectonicBundle() error = %v", err)
	}
	if resolved != bundle {
		t.Fatalf("resolveTectonicBundle() = %q, want %q", resolved, bundle)
	}
}

func TestResolveTectonicBundleUsesFileBesideExecutable(t *testing.T) {
	directory := t.TempDir()
	bundle := filepath.Join(directory, bundleFilename)
	if err := os.WriteFile(bundle, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("TAILORCV_TECTONIC_BUNDLE", "")

	resolved, err := resolveTectonicBundle(filepath.Join(directory, "tectonic"))
	if err != nil {
		t.Fatalf("resolveTectonicBundle() error = %v", err)
	}
	if resolved != bundle {
		t.Fatalf("resolveTectonicBundle() = %q, want %q", resolved, bundle)
	}
}

func TestCompilerArgumentsRequireOfflineLocalBundle(t *testing.T) {
	arguments := compilerArguments("workspace", "resume.tex", "resources.zip")
	for _, expected := range []string{"--only-cached", "--untrusted", "--bundle", "resources.zip"} {
		if !slices.Contains(arguments, expected) {
			t.Fatalf("compilerArguments() = %#v, missing %q", arguments, expected)
		}
	}
	for _, forbidden := range []string{"--shell-escape", "--keep-intermediates"} {
		if slices.Contains(arguments, forbidden) {
			t.Fatalf("compilerArguments() = %#v, contains unsafe %q", arguments, forbidden)
		}
	}
}

func TestCompilerCacheIsConfinedToCompileWorkspace(t *testing.T) {
	workspace := t.TempDir()
	cache := compilerCacheDirectory(workspace)
	relative, err := filepath.Rel(workspace, cache)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		t.Fatalf("compilerCacheDirectory(%q) = %q, outside workspace", workspace, cache)
	}
	info, err := os.Stat(cache)
	if err != nil || !info.IsDir() {
		t.Fatalf("compiler cache was not created: %v", err)
	}
}

func TestLocalBundleLocatorUsesFileURLForAbsolutePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resources with spaces.zip")
	locator := localBundleLocator(path)
	if !strings.HasPrefix(locator, "file://") || !strings.Contains(locator, "resources%20with%20spaces.zip") {
		t.Fatalf("localBundleLocator(%q) = %q", path, locator)
	}
	if runtime.GOOS == "windows" && !strings.HasPrefix(locator, "file:///") {
		t.Fatalf("localBundleLocator(%q) = %q, want a Windows file URL", path, locator)
	}
}

func TestCompilerRejectsInvalidInputBeforeExecution(t *testing.T) {
	compiler := NewCompiler("/path/that/does/not/exist")
	if _, _, err := compiler.Compile(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("Compile(empty) error = %v", err)
	}
	if _, _, err := compiler.Compile(context.Background(), "bad\x00source"); err == nil || !strings.Contains(err.Error(), "null byte") {
		t.Fatalf("Compile(null) error = %v", err)
	}
}

func TestCompilerReportsUnavailableExecutableAndBundle(t *testing.T) {
	t.Setenv("TAILORCV_TECTONIC", filepath.Join(t.TempDir(), "missing-tectonic"))
	if _, _, err := NewCompiler("").Compile(context.Background(), `\documentclass{article}`); err == nil || !strings.Contains(err.Error(), "does not point") {
		t.Fatalf("Compile() unavailable executable error = %v", err)
	}

	executable := filepath.Join(t.TempDir(), "tectonic")
	if err := os.WriteFile(executable, []byte("fixture"), 0o700); err != nil {
		t.Fatalf("WriteFile(executable) error = %v", err)
	}
	t.Setenv("TAILORCV_TECTONIC", executable)
	t.Setenv("TAILORCV_TECTONIC_BUNDLE", filepath.Join(t.TempDir(), "missing-bundle.zip"))
	if _, _, err := NewCompiler("").Compile(context.Background(), `\documentclass{article}`); err == nil || !strings.Contains(err.Error(), "does not point") {
		t.Fatalf("Compile() unavailable bundle error = %v", err)
	}
}

func TestCappedBufferLimitsCompilerOutput(t *testing.T) {
	buffer := &cappedBuffer{limit: 4}
	if count, err := buffer.Write([]byte("abcdef")); err != nil || count != 6 {
		t.Fatalf("Write() = %d, %v", count, err)
	}
	if got := buffer.String(); got != "abcd\n[compiler output truncated]" {
		t.Fatalf("String() = %q", got)
	}
}

func TestParseCompilerDiagnostics(t *testing.T) {
	output := `error: resume.tex:12: Undefined control sequence
warning: resume.tex:4: Overfull \hbox
! Missing $ inserted.
l.20 total = $value`

	diagnostics := parseCompilerDiagnostics(output)
	if len(diagnostics) != 3 {
		t.Fatalf("parseCompilerDiagnostics() returned %d diagnostics: %#v", len(diagnostics), diagnostics)
	}
	want := []struct {
		line     int
		severity string
		message  string
	}{
		{line: 12, severity: "error", message: "Undefined control sequence"},
		{line: 4, severity: "warning", message: `Overfull \hbox`},
		{line: 20, severity: "error", message: "Missing $ inserted. — total = $value"},
	}
	for index, expected := range want {
		actual := diagnostics[index]
		if actual.Line != expected.line || actual.Severity != expected.severity || actual.Message != expected.message {
			t.Errorf("diagnostic %d = %#v, want line=%d severity=%q message=%q", index, actual, expected.line, expected.severity, expected.message)
		}
	}
}

func TestFirstCompilerMessageSkipsNotesAndBlankLines(t *testing.T) {
	if got := firstCompilerMessage("\n note: checking source\n error: broken source\n"); got != "error: broken source" {
		t.Fatalf("firstCompilerMessage() = %q", got)
	}
}

func TestCompilerIntegration(t *testing.T) {
	if os.Getenv("TAILORCV_TECTONIC_INTEGRATION") == "" {
		t.Skip("set TAILORCV_TECTONIC_INTEGRATION=1 to exercise the local Tectonic executable")
	}
	compiler := NewCompiler("")
	for _, template := range BuiltinTemplates() {
		t.Run(template.ID, func(t *testing.T) {
			result, pdf, err := compiler.Compile(context.Background(), Render(template.Source, Data{}))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			if !result.Success || result.Engine != "Tectonic" || result.PDFBase64 == "" || len(pdf) < 100 {
				t.Fatalf("Compile() returned incomplete output: %#v, PDF bytes = %d", result, len(pdf))
			}
		})
	}

	t.Run("untrusted-input-cannot-read-an-arbitrary-file", func(t *testing.T) {
		outside := t.TempDir()
		secretToken := "TAILORCV-UNTRUSTED-SECRET-74A5"
		secretPath := filepath.Join(outside, "secret.tex")
		if err := os.WriteFile(secretPath, []byte(`\errmessage{`+secretToken+`}`), 0o600); err != nil {
			t.Fatalf("WriteFile(secret) error = %v", err)
		}
		source := `\documentclass{article}\begin{document}\input{` + filepath.ToSlash(secretPath) + `}\end{document}`
		result, _, err := compiler.Compile(context.Background(), source)
		if err != nil {
			t.Fatalf("Compile(read attempt) setup error = %v", err)
		}
		if strings.Contains(result.Log, secretToken) {
			t.Fatalf("untrusted compilation read arbitrary input; log contains secret token")
		}
	})

	t.Run("untrusted-input-cannot-write-or-run-shell-outside-workspace", func(t *testing.T) {
		outside := t.TempDir()
		texWritePath := filepath.Join(outside, "tex-write.txt")
		shellWritePath := filepath.Join(outside, "shell-write.txt")
		source := `\documentclass{article}
\begin{document}
\newwrite\outside
\immediate\openout\outside=` + filepath.ToSlash(texWritePath) + `
\immediate\write\outside{compromised}
\immediate\closeout\outside
\immediate\write18{echo compromised > "` + filepath.ToSlash(shellWritePath) + `"}
Safe output
\end{document}`
		if _, _, err := compiler.Compile(context.Background(), source); err != nil {
			t.Fatalf("Compile(write attempt) setup error = %v", err)
		}
		for _, path := range []string{texWritePath, shellWritePath} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("untrusted compilation wrote outside its workspace: %s", path)
			}
		}
	})
}
