package gen_test

import (
	"flag"
	"os"
	"path/filepath"
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
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	testdata := filepath.Join(wd, "testdata")
	t.Setenv("GOPATH", testdata)
	rs := codegentest.Run(t, testdata, gen.Generator, "a")
	if len(rs) > 0 {
		for _, r := range rs {
			if r.Err != nil {
				t.Errorf("unexpected error: %v", r.Err)
			}
			if r.Output != nil {
				t.Log(r.Output.String())
			}
		}
	}
	codegentest.Golden(t, rs, flagUpdate)
}
