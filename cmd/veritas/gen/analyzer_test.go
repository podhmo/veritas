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
