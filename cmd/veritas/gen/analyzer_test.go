package gen_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	dir := filepath.Join(codegentest.TestData(), "src")
	rs := codegentest.Run(t, dir, gen.Generator, "testpkg/a")

	for _, r := range rs {
		r.Dir = strings.Replace(r.Dir, "src/src/testpkg", "src", 1) // workaround for codegentest

		if r.Err != nil {
			t.Errorf("failed to generate code: %v", r.Err)
		}
		if r.Output != nil {
			t.Log(r.Output.String())
		}
	}

	codegentest.Golden(t, rs, flagUpdate)
}
