package gen_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gostaticanalysis/codegen/codegentest"
	"github.com/podhmo/veritas/cmd/veritas/gen"
)

var flagUpdate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	cases := []struct {
		Name    string
		Args    []string
		PkgPath string
		Golden  string
	}{
		{
			Name:    "default",
			PkgPath: "testpkg/a",
			Golden:  "testdata/src/a/gogen.golden",
		},
		{
			Name:    "with-pkg-flag",
			Args:    []string{"-pkg=myvalidation"},
			PkgPath: "testpkg/a",
			Golden:  "testdata/src/a/gogen.golden.pkg",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			gen.Generator.Flags.Parse(c.Args)

			dir := filepath.Join(codegentest.TestData(), "src")
			rs := codegentest.Run(t, dir, gen.Generator, c.PkgPath)

			if len(rs) != 1 {
				t.Fatalf("unexpected number of results: %d", len(rs))
			}
			r := rs[0]
			r.Dir = strings.Replace(r.Dir, "src/src/testpkg", "src", 1) // workaround for codegentest

			if r.Err != nil {
				t.Errorf("failed to generate code: %v", r.Err)
			}

			if flagUpdate {
				if err := os.WriteFile(c.Golden, r.Output.Bytes(), 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				return
			}

			golden, err := os.ReadFile(c.Golden)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}
			if diff := cmp.Diff(string(golden), r.Output.String()); diff != "" {
				t.Errorf("output differs from golden file:\n%s", diff)
			}
		})
	}
}
