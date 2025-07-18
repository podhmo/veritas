package lint_test

import (
	"testing"

	"github.com/podhmo/veritas/lint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, lint.Analyzer, "a", "b", "c")
}
