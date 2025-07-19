package lint

import (
	"github.com/podhmo/veritas/lint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func Main() {
	singlechecker.Main(lint.Analyzer)
}
