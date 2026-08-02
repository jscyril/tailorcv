package resume

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/domain"
)

const (
	compileTimeout = 45 * time.Second
	maxCompilerLog = 256 << 10
	maxPDFSize     = 24 << 20
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
		executable, err = exec.LookPath("tectonic")
		if err != nil {
			return domain.CompileResult{}, nil, fmt.Errorf("Tectonic is not installed; install Tectonic and restart TailorCV")
		}
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
	command := exec.CommandContext(compileContext, executable, "--untrusted", "--keep-logs", "--color", "never", "--outdir", workspace, inputPath)
	command.Dir = workspace
	command.Env = withEnvironment(os.Environ(), "XDG_CACHE_HOME", compilerCacheDirectory(workspace))
	command.Stdout, command.Stderr = log, log
	started := time.Now()
	err = command.Run()
	duration := time.Since(started)
	if compileContext.Err() == context.DeadlineExceeded {
		return domain.CompileResult{}, nil, fmt.Errorf("LaTeX compilation exceeded the %s time limit", compileTimeout)
	}
	if err != nil {
		message := strings.TrimSpace(log.String())
		if message == "" {
			message = err.Error()
		}
		return domain.CompileResult{}, nil, fmt.Errorf("LaTeX compilation failed:\n%s", message)
	}
	pdf, err := os.ReadFile(filepath.Join(workspace, "resume.pdf"))
	if err != nil {
		return domain.CompileResult{}, nil, fmt.Errorf("read compiled PDF: %w", err)
	}
	if len(pdf) == 0 || len(pdf) > maxPDFSize {
		return domain.CompileResult{}, nil, fmt.Errorf("compiled PDF is empty or exceeds the 24 MiB limit")
	}
	result := domain.CompileResult{
		PDFBase64:  base64.StdEncoding.EncodeToString(pdf),
		Engine:     "Tectonic",
		DurationMS: duration.Milliseconds(),
		Log:        strings.TrimSpace(log.String()),
	}
	return result, pdf, nil
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
