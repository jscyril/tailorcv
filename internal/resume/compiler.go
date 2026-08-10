package resume

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/domain"
)

const (
	compileTimeout = 45 * time.Second
	maxCompilerLog = 256 << 10
	bundleFilename = "tectonic-resources.zip"
)

type Compiler struct {
	executable string
}

func NewCompiler(executable string) *Compiler {
	return &Compiler{executable: executable}
}

func (compiler *Compiler) Compile(ctx context.Context, source string) (domain.CompileResult, []byte, error) {
	if len(source) == 0 {
		return domain.CompileResult{}, nil, fmt.Errorf("LaTeX source is empty")
	}
	if len(source) > domain.MaxTemplateSourceBytes {
		return domain.CompileResult{}, nil, fmt.Errorf("LaTeX source exceeds the 1 MiB size limit")
	}
	if strings.IndexByte(source, 0) >= 0 {
		return domain.CompileResult{}, nil, fmt.Errorf("LaTeX source contains a null byte")
	}
	executable := compiler.executable
	if executable == "" {
		var err error
		executable, err = resolveTectonicExecutable()
		if err != nil {
			return domain.CompileResult{}, nil, err
		}
	}
	bundle, err := resolveTectonicBundle(executable)
	if err != nil {
		return domain.CompileResult{}, nil, err
	}
	workspace, err := os.MkdirTemp("", "tailorcv-compile-*")
	if err != nil {
		return domain.CompileResult{}, nil, fmt.Errorf("create isolated compile workspace: %w", err)
	}
	defer os.RemoveAll(workspace)
	inputPath := filepath.Join(workspace, "resume.tex")
	if err := os.WriteFile(inputPath, []byte(source), 0o600); err != nil {
		return domain.CompileResult{}, nil, fmt.Errorf("write isolated LaTeX source: %w", err)
	}
	compileContext, cancel := context.WithTimeout(ctx, compileTimeout)
	defer cancel()
	log := &cappedBuffer{limit: maxCompilerLog}
	command := exec.CommandContext(compileContext, executable, compilerArguments(workspace, inputPath, bundle)...)
	command.Dir = workspace
	cacheDirectory := compilerCacheDirectory(workspace)
	commandEnvironment := withEnvironment(os.Environ(), "XDG_CACHE_HOME", cacheDirectory)
	commandEnvironment = withEnvironment(commandEnvironment, "TECTONIC_CACHE_DIR", cacheDirectory)
	command.Env = withEnvironment(commandEnvironment, "TECTONIC_UNTRUSTED_MODE", "1")
	command.Stdout, command.Stderr = log, log
	started := time.Now()
	err = command.Run()
	duration := time.Since(started)
	compilerLog := strings.TrimSpace(log.String())
	if compileContext.Err() == context.DeadlineExceeded {
		return domain.CompileResult{}, nil, fmt.Errorf("LaTeX compilation exceeded the %s time limit", compileTimeout)
	}
	result := domain.CompileResult{
		Success:     err == nil,
		Engine:      "Tectonic",
		DurationMS:  duration.Milliseconds(),
		Log:         compilerLog,
		Diagnostics: parseCompilerDiagnostics(compilerLog),
	}
	if err != nil {
		if len(result.Diagnostics) == 0 {
			message := firstCompilerMessage(compilerLog)
			if message == "" {
				message = err.Error()
			}
			result.Diagnostics = []domain.CompileDiagnostic{{Severity: "error", Message: message}}
		}
		return result, nil, nil
	}
	pdf, err := os.ReadFile(filepath.Join(workspace, "resume.pdf"))
	if err != nil {
		return domain.CompileResult{}, nil, fmt.Errorf("read compiled PDF: %w", err)
	}
	if len(pdf) == 0 || len(pdf) > domain.MaxCompiledPDFBytes {
		return domain.CompileResult{}, nil, fmt.Errorf("compiled PDF is empty or exceeds the 24 MiB limit")
	}
	result.PDFBase64 = base64.StdEncoding.EncodeToString(pdf)
	return result, pdf, nil
}

func resolveTectonicExecutable() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("TAILORCV_TECTONIC")); configured != "" {
		if info, err := os.Stat(configured); err == nil && !info.IsDir() {
			return filepath.Clean(configured), nil
		}
		return "", fmt.Errorf("TAILORCV_TECTONIC does not point to a Tectonic executable")
	}
	if current, err := os.Executable(); err == nil {
		name := "tectonic"
		if strings.EqualFold(filepath.Ext(current), ".exe") {
			name += ".exe"
		}
		for _, candidate := range []string{
			filepath.Join(filepath.Dir(current), "bin", name),
			filepath.Join(filepath.Dir(current), name),
		} {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	if executable, err := exec.LookPath("tectonic"); err == nil {
		return executable, nil
	}
	return "", fmt.Errorf("Tectonic is unavailable; package it in TailorCV's bin directory, set TAILORCV_TECTONIC, or install it on PATH")
}

func resolveTectonicBundle(executable string) (string, error) {
	if configured := strings.TrimSpace(os.Getenv("TAILORCV_TECTONIC_BUNDLE")); configured != "" {
		if isRegularFile(configured) {
			return filepath.Clean(configured), nil
		}
		return "", fmt.Errorf("TAILORCV_TECTONIC_BUNDLE does not point to a local Tectonic resource bundle")
	}

	candidate := filepath.Join(filepath.Dir(executable), bundleFilename)
	if isRegularFile(candidate) {
		return candidate, nil
	}
	return "", fmt.Errorf("Tectonic's offline resource bundle is unavailable; package %s beside Tectonic or set TAILORCV_TECTONIC_BUNDLE", bundleFilename)
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func compilerArguments(workspace, inputPath, bundle string) []string {
	return []string{
		"--only-cached",
		"--untrusted",
		"--keep-logs",
		"--color", "never",
		"--bundle", localBundleLocator(bundle),
		"--outdir", workspace,
		inputPath,
	}
}

func localBundleLocator(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	slashed := filepath.ToSlash(path)
	if runtime.GOOS == "windows" && filepath.VolumeName(path) != "" && !strings.HasPrefix(slashed, "/") {
		slashed = "/" + slashed
	}
	return (&url.URL{Scheme: "file", Path: slashed}).String()
}

var (
	structuredDiagnosticPattern = regexp.MustCompile(`(?i)^(error|warning):\s*(?:(?:[^:]+):(\d+):\s*)?(.+)$`)
	texLinePattern              = regexp.MustCompile(`^l\.(\d+)\s*(.*)$`)
	locationPattern             = regexp.MustCompile(`(?:-->|at)?\s*[^:\s]+:(\d+)(?::\d+)?`)
)

func parseCompilerDiagnostics(output string) []domain.CompileDiagnostic {
	diagnostics := make([]domain.CompileDiagnostic, 0)
	var pending *domain.CompileDiagnostic
	appendDiagnostic := func(diagnostic domain.CompileDiagnostic) {
		diagnostic.Message = strings.Join(strings.Fields(diagnostic.Message), " ")
		if diagnostic.Message == "" {
			return
		}
		for _, existing := range diagnostics {
			if existing.Line == diagnostic.Line && existing.Severity == diagnostic.Severity && existing.Message == diagnostic.Message {
				return
			}
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	flushPending := func() {
		if pending != nil {
			appendDiagnostic(*pending)
			pending = nil
		}
	}

	for _, sourceLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(sourceLine)
		if line == "" {
			continue
		}
		if match := structuredDiagnosticPattern.FindStringSubmatch(line); match != nil {
			flushPending()
			diagnostic := domain.CompileDiagnostic{Severity: strings.ToLower(match[1]), Message: match[3]}
			if match[2] != "" {
				diagnostic.Line, _ = strconv.Atoi(match[2])
				appendDiagnostic(diagnostic)
			} else {
				pending = &diagnostic
			}
			continue
		}
		if strings.HasPrefix(line, "!") {
			flushPending()
			pending = &domain.CompileDiagnostic{Severity: "error", Message: strings.TrimSpace(strings.TrimPrefix(line, "!"))}
			continue
		}
		if pending != nil {
			if match := texLinePattern.FindStringSubmatch(line); match != nil {
				pending.Line, _ = strconv.Atoi(match[1])
				if context := strings.TrimSpace(match[2]); context != "" {
					pending.Message += " — " + context
				}
				flushPending()
				continue
			}
			if match := locationPattern.FindStringSubmatch(line); match != nil {
				pending.Line, _ = strconv.Atoi(match[1])
				flushPending()
			}
		}
	}
	flushPending()
	return diagnostics
}

func firstCompilerMessage(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(strings.ToLower(line), "note:") {
			return line
		}
	}
	return ""
}

func compilerCacheDirectory(fallback string) string {
	if root, err := os.UserCacheDir(); err == nil {
		path := filepath.Join(root, "tailorcv")
		if err := os.MkdirAll(path, 0o700); err == nil {
			return path
		}
	}
	path := filepath.Join(fallback, "cache")
	_ = os.MkdirAll(path, 0o700)
	return path
}

func withEnvironment(environment []string, key, value string) []string {
	prefix := key + "="
	result := make([]string, 0, len(environment)+1)
	for _, item := range environment {
		if !strings.HasPrefix(item, prefix) {
			result = append(result, item)
		}
	}
	return append(result, prefix+value)
}

type cappedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (buffer *cappedBuffer) Write(data []byte) (int, error) {
	originalLength := len(data)
	remaining := buffer.limit - buffer.buffer.Len()
	if remaining > 0 {
		if len(data) > remaining {
			data = data[:remaining]
			buffer.truncated = true
		}
		_, _ = buffer.buffer.Write(data)
	} else {
		buffer.truncated = true
	}
	return originalLength, nil
}

func (buffer *cappedBuffer) String() string {
	if buffer.truncated {
		return buffer.buffer.String() + "\n[compiler output truncated]"
	}
	return buffer.buffer.String()
}
