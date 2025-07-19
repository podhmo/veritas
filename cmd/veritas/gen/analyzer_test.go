package gen_test

import (
	"bytes"
	"flag"
	"fmt"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gostaticanalysis/codegen/codegentest"
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
	dir := filepath.Join(codegentest.TestData(), "src")

	// workaround for testdata/src/go.mod
	// {
	// 	cwd, err := os.Getwd()
	// 	if err != nil {
	// 		t.Fatalf("failed to get cwd: %v", err)
	// 	}
	// 	os.Chdir(dir)
	// 	t.Cleanup(func() {
	// 		os.Chdir(cwd)
	// 	})
	// }

	os.Setenv("PWD", dir)

	// rs := codegentest.Run(t, dir, gen.Generator, "testpkg/a")
	// for _, r := range rs {
	// 	r.Dir = strings.Replace(r.Dir, "src/src/testpkg", "src", 1) // workaround for codegentest

	// 	if r.Err != nil {
	// 		t.Errorf("failed to generate code: %v", r.Err)
	// 	}
	// 	if r.Output != nil {
	// 		t.Log(r.Output.String())
	// 	}
	// }
	// codegentest.Golden(t, rs, flagUpdate)

	{
		mode := packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedDeps | packages.NeedModule
		gowork := "off"
		env := []string{"GO111MODULE=on", "GOPROXY=off", "GOWORK=" + gowork} // module mode
		cfg := &packages.Config{
			Mode:  mode,
			Dir:   dir,
			Tests: true,
			Env:   append(os.Environ(), env...),
		}
		pkgs, err := packages.Load(cfg, "testpkg/a")

		if err != nil {
			t.Errorf("failed to load packages: %v", err)
		}

		for _, pkg := range pkgs {
			t.Logf("package: %s, files: %s", pkg.PkgPath, pkg.GoFiles)
			t.Logf("error? %+v", pkg.Errors)
		}
		packages.PrintErrors(pkgs)

		g, err := checker.Analyze([]*analysis.Analyzer{
			gen.Generator.ToAnalyzer(),
		}, pkgs, nil)

		if err != nil {
			t.Errorf("failed to create analyzer: %v", err)
		}
		if g == nil {
			t.Fatal("analyzer is nil")
		}

		for act := range g.All() {
			if err := act.Err; err != nil {
				t.Errorf("failed to generate code: %v", err)
			}
		}
	}

	fmt.Println("````````````````````````````````````````")
	{
		outputs := map[*types.Package]*bytes.Buffer{}
		g := gen.Generator

		g.Output = func(pkg *types.Package) io.Writer {
			t.Logf("loading package: %s", pkg.Path())
			var buf bytes.Buffer
			outputs[pkg] = &buf
			return &buf
		}

		rs := analysistest.Run(t, dir, gen.Generator.ToAnalyzer(), "testpkg/a")
		for _, r := range rs {
			if r.Err != nil {
				t.Errorf("failed to generate code: %v", r.Err)
			}

			output := outputs[r.Pass.Pkg]
			if output != nil {
				t.Log(output.String())
			}
		}
	}

}
