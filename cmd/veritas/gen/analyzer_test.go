package gen_test

import (
	"bytes"
	"flag"
	"fmt"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/podhmo/veritas/cmd/veritas/gen"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

var flagUpdate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	// This is a workaround for the fact that codegentest is broken.
	buf := new(bytes.Buffer)
	gen.Generator.Output = func(pkg *types.Package) io.Writer {
		return buf
	}

	if err := doGenerate(
		analysistest.TestData()+"/src",
		gen.Generator.ToAnalyzer(),
		[]string{"./a"},
	); err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	goldenFile := filepath.Join(analysistest.TestData(), "src/a/gogen.golden")
	if flagUpdate { // update golden file if -update flag is set
		if err := os.WriteFile(goldenFile, buf.Bytes(), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	} else {
		golden, err := os.ReadFile(goldenFile)
		if err != nil {
			t.Fatalf("failed to read golden file: %v", err)
		}
		want, got := strings.TrimSpace(string(golden)), strings.TrimSpace(buf.String())
		if want != got {
			t.Errorf("output does not match golden file:\nwant:\n%s\ngot:\n%s", want, got)
		}
	}
}

// codegentest is broken, so we use individual code.
func doGenerate(dir string, an *analysis.Analyzer, patterns []string) error {
	mode := packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
		packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo |
		packages.NeedDeps | packages.NeedModule
	cfg := &packages.Config{
		Mode:  mode,
		Dir:   dir,
		Tests: true,
		Env:   os.Environ(),
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	for _, pkg := range pkgs {
		if pkg.Name == "" {
			return fmt.Errorf("failed to load %q: Errors=%v", pkg.PkgPath, pkg.Errors)
		}
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages matched %s", patterns)
	}

	_, err = checker.Analyze([]*analysis.Analyzer{an}, pkgs, nil)
	if err != nil {
		return fmt.Errorf("failed to analyze packages: %w", err)
	}

	return nil
}
