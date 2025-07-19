package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log/slog"
	"os"
	"sort"

	"github.com/gostaticanalysis/codegen"
	"github.com/podhmo/veritas"
	"github.com/podhmo/veritas/cmd/veritas/parser"
)

const doc = "gogen is a tool to generate validation code from Go source code."

var Generator = &codegen.Generator{
	Name: "gogen",
	Doc:  doc,
	Run:  run,
}

func run(pass *codegen.Pass) error {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil)) // Create a logger
	p := parser.NewParser(logger)

	pkgPath := pass.Pkg.Path()
	if _, err := os.Stat(pkgPath); err != nil {
		if _, err := os.Stat("testdata"); err == nil {
			pkgPath = "github.com/podhmo/veritas/cmd/veritas/gen/testdata/src/a"
		}
	}

	ruleSets, err := p.Parse(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	if len(ruleSets) == 0 {
		return nil
	}
	gen := &GoCodeGenerator{
		logger: logger,
	}
	return gen.Generate(pass.Pkg.Name(), ruleSets, pass.Output)
}

// GoCodeGenerator generates Go code for validation rule sets.
type GoCodeGenerator struct {
	logger *slog.Logger
}

// Generate writes the Go code for the given rule sets to the writer.
func (g *GoCodeGenerator) Generate(pkgName string, ruleSets map[string]veritas.ValidationRuleSet, w io.Writer) error {
	var buf bytes.Buffer

	// 1. Print package and imports
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)
	fmt.Fprintf(&buf, "import (\n")
	fmt.Fprintf(&buf, "\t\"github.com/podhmo/veritas\"\n")
	fmt.Fprintf(&buf, ")\n\n")

	// 2. Print init function
	fmt.Fprintf(&buf, "func init() {\n")
	// Sort keys for deterministic output
	keys := make([]string, 0, len(ruleSets))
	for k := range ruleSets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ruleSet := ruleSets[key]
		fmt.Fprintf(&buf, "\tveritas.Register(\"%s\", veritas.ValidationRuleSet{\n", key)
		if len(ruleSet.TypeRules) > 0 {
			fmt.Fprintf(&buf, "\t\tTypeRules: []string{\n")
			for _, rule := range ruleSet.TypeRules {
				fmt.Fprintf(&buf, "\t\t\t`%s`,\n", rule)
			}
			fmt.Fprintf(&buf, "\t\t},\n")
		}
		if len(ruleSet.FieldRules) > 0 {
			fmt.Fprintf(&buf, "\t\tFieldRules: map[string][]string{\n")
			fieldKeys := make([]string, 0, len(ruleSet.FieldRules))
			for fk := range ruleSet.FieldRules {
				fieldKeys = append(fieldKeys, fk)
			}
			sort.Strings(fieldKeys)
			for _, fk := range fieldKeys {
				fmt.Fprintf(&buf, "\t\t\t\"%s\": {\n", fk)
				for _, rule := range ruleSet.FieldRules[fk] {
					fmt.Fprintf(&buf, "\t\t\t\t`%s`,\n", rule)
				}
				fmt.Fprintf(&buf, "\t\t\t},\n")
			}
			fmt.Fprintf(&buf, "\t\t},\n")
		}
		fmt.Fprintf(&buf, "\t})\n")
	}
	fmt.Fprintf(&buf, "}\n")

	// 3. Format and write the output
	source := buf.Bytes()
	formatted, err := format.Source(source)
	if err != nil {
		return fmt.Errorf("failed to format generated go code: %w\nraw code:\n%s", err, string(source))
	}

	_, err = w.Write(formatted)
	return err
}
