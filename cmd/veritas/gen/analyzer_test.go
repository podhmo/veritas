package gen_test

import (
	"flag"
	"os"
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

// NEED: GOWORK=off if go.work is existed in toplevel
func TestGenerator(t *testing.T) {
	rs := codegentest.Run(t, codegentest.TestData(), gen.Generator, "a")
	for _, r := range rs {
		if r.Err != nil {
			t.Errorf("failed to generate code: %v", r.Err)
		}
		if r.Output != nil {
			t.Log(r.Output.String())
		}
	}
	codegentest.Golden(t, rs, flagUpdate)
}
