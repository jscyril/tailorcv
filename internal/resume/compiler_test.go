package resume

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestCompilerRejectsInvalidInputBeforeExecution(t *testing.T) {
	compiler := NewCompiler("/path/that/does/not/exist")
	if _, _, err := compiler.Compile(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("Compile(empty) error = %v", err)
	}
	if _, _, err := compiler.Compile(context.Background(), "bad\x00source"); err == nil || !strings.Contains(err.Error(), "null byte") {
		t.Fatalf("Compile(null) error = %v", err)
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
			if result.Engine != "Tectonic" || result.PDFBase64 == "" || len(pdf) < 100 {
				t.Fatalf("Compile() returned incomplete output: %#v, PDF bytes = %d", result, len(pdf))
			}
		})
	}
}
