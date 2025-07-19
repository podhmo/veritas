package required_test

import (
	"testing"

	"github.com/podhmo/veritas/lint/required"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, required.Analyzer, "d")
}
