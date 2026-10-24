package resume

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMissingBundleFilesReadsTectonicErrors(t *testing.T) {
	log := "error: resume.tex:3: ! LaTeX Error: File `tcolorbox.sty' not found.\n" +
		"error: resume.tex:130: Font TU/lmtt/m/n/10=[lmmono10-regular]:mapping=tex-text; at 10.0pt not loadable: Metric (TFM) file or installed font not found\n" +
		"error: Font \\x=wasy10 at 10pt not loadable: Metric (TFM) file not found\n" +
		"! LaTeX Error: File `/etc/passwd' not found.\n" +
		"! I can't find file `../secret'.\n" +
		"! LaTeX Error: File `resume.aux' not found.\n" +
		"warning: >> This font is mapped to a physical font \"cmr10.pfb\".\n" +
		"error: failed to open input file \"xstring.tex\"\n"
	want := []string{"tcolorbox.sty", "lmmono10-regular.otf", "lmmono10-regular.ttf", "wasy10.tfm", "cmr10.pfb", "xstring.tex"}
	if got := missingBundleFiles(log); !reflect.DeepEqual(got, want) {
		t.Fatalf("missingBundleFiles() = %v, want %v", got, want)
	}
}

func TestCopyExtrasSkipsUnsafeNames(t *testing.T) {
	extras, workspace := t.TempDir(), t.TempDir()
	for _, name := range []string{"tcolorbox.sty", "resume.tex"} {
		if err := os.WriteFile(filepath.Join(extras, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := copyExtras(extras, workspace); err != nil {
		t.Fatalf("copyExtras() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "tcolorbox.sty")); err != nil {
		t.Fatalf("extra was not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "resume.tex")); !os.IsNotExist(err) {
		t.Fatal("copyExtras() overwrote the source name")
	}
	if err := copyExtras(filepath.Join(extras, "missing"), workspace); err != nil {
		t.Fatalf("copyExtras(missing folder) error = %v", err)
	}
}

func TestCompilerFetchesMissingPackagesIntegration(t *testing.T) {
	if os.Getenv("TAILORCV_TECTONIC_INTEGRATION") == "" {
		t.Skip("set TAILORCV_TECTONIC_INTEGRATION=1 to fetch from the online bundle")
	}
	compiler := NewCompiler("")
	compiler.extras = filepath.Join(t.TempDir(), "tex-extras")
	source := `\documentclass{article}
\usepackage{marvosym,wasysym,multicol,xstring}
\begin{document}
\Letter\ \smiley\ \texttt{hello}
\end{document}`
	result, pdf, err := compiler.Compile(context.Background(), source)
	if err != nil || !result.Success || len(pdf) == 0 {
		t.Fatalf("Compile() = %#v, %v", result.Diagnostics, err)
	}
	if _, err := os.Stat(filepath.Join(compiler.extras, "wasysym.sty")); err != nil {
		t.Fatalf("wasysym.sty was not cached: %v", err)
	}
}
