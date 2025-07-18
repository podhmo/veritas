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

func TestGenerator(t *testing.T) {
	rs := codegentest.Run(t, codegentest.TestData(), gen.Generator, "a")
	codegentest.Golden(t, rs, flagUpdate)
}
