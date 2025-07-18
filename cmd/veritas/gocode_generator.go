package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log/slog"
	"sort"

	"github.com/podhmo/veritas"
)

// GoCodeGenerator generates Go code for validation rule sets.
type GoCodeGenerator struct {
	logger *slog.Logger
	buf    *bytes.Buffer
}

// NewGoCodeGenerator creates a new GoCodeGenerator.
func NewGoCodeGenerator(logger *slog.Logger) *GoCodeGenerator {
	return &GoCodeGenerator{
		logger: logger,
		buf:    new(bytes.Buffer),
	}
}

// Generate writes the Go code for the given rule sets to the writer.
func (g *GoCodeGenerator) Generate(pkgName string, ruleSets map[string]veritas.ValidationRuleSet, w io.Writer) error {
	g.buf.Reset()

	// 1. Print package and imports
	g.printf("package %s\n\n", pkgName)
	g.printf("import (\n")
	g.printf("\t\"github.com/podhmo/veritas\"\n")
	g.printf(")\n\n")

	// 2. Print init function
	g.printf("func init() {\n")
	// Sort keys for deterministic output
	keys := make([]string, 0, len(ruleSets))
	for k := range ruleSets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ruleSet := ruleSets[key]
		g.printf("\tveritas.Register(\"%s\", veritas.ValidationRuleSet{\n", key)
		if len(ruleSet.TypeRules) > 0 {
			g.printf("\t\tTypeRules: []string{\n")
			for _, rule := range ruleSet.TypeRules {
				g.printf("\t\t\t`%s`,\n", rule)
			}
			g.printf("\t\t},\n")
		}
		if len(ruleSet.FieldRules) > 0 {
			g.printf("\t\tFieldRules: map[string][]string{\n")
			fieldKeys := make([]string, 0, len(ruleSet.FieldRules))
			for fk := range ruleSet.FieldRules {
				fieldKeys = append(fieldKeys, fk)
			}
			sort.Strings(fieldKeys)
			for _, fk := range fieldKeys {
				g.printf("\t\t\t\"%s\": {\n", fk)
				for _, rule := range ruleSet.FieldRules[fk] {
					g.printf("\t\t\t\t`%s`,\n", rule)
				}
				g.printf("\t\t\t},\n")
			}
			g.printf("\t\t},\n")
		}
		g.printf("\t})\n")
	}
	g.printf("}\n")

	// 3. Format and write the output
	formatted, err := format.Source(g.buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated go code: %w\nraw code:\n%s", err, g.buf.String())
	}

	_, err = w.Write(formatted)
	return err
}

func (g *GoCodeGenerator) printf(format string, args ...interface{}) {
	fmt.Fprintf(g.buf, format, args...)
}
