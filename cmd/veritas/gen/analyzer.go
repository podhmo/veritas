package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/gostaticanalysis/codegen"
	"github.com/podhmo/veritas"
	"github.com/podhmo/veritas/cmd/veritas/parser"
)

const doc = "gogen is a tool to generate validation code from Go source code."

var (
	flagOutput string
)

var Generator = &codegen.Generator{
	Name: "gogen",
	Doc:  doc,
	Run:  run,
}

func init() {
	Generator.Flags.StringVar(&flagOutput, "o", "", "output file name")
}

func run(pass *codegen.Pass) error {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	p := parser.NewParser(logger)

	info := parser.PackageInfo{
		PkgPath:   pass.Pkg.Path(),
		Syntax:    pass.Files,
		TypesInfo: pass.TypesInfo,
		Types:     pass.Pkg,
	}

	ruleSets, typeInfos, err := p.ParseDirectly(info)
	if err != nil {
		return fmt.Errorf("failed to parse directly: %w", err)
	}

	if len(ruleSets) == 0 {
		return nil
	}
	gen := &GoCodeGenerator{
		logger: logger,
	}

	if flagOutput == "" {
		return gen.Generate(pass.Pkg.Name(), ruleSets, typeInfos, pass.Output)
	}

	f, err := os.Create(flagOutput)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()
	return gen.Generate(pass.Pkg.Name(), ruleSets, typeInfos, f)
}

type GoCodeGenerator struct {
	logger *slog.Logger
}

func (g *GoCodeGenerator) Generate(pkgName string, ruleSets map[string]veritas.ValidationRuleSet, typeInfos []parser.TypeInfo, w io.Writer) error {
	var buf bytes.Buffer

	// Group types by package path to generate imports
	imports := make(map[string]string)
	for _, ti := range typeInfos {
		if _, exists := imports[ti.PkgPath]; !exists {
			// Generate a safe alias for the package
			alias := strings.ReplaceAll(ti.PkgPath, "/", "_")
			alias = strings.ReplaceAll(alias, "-", "_")
			alias = strings.ReplaceAll(alias, ".", "_")
			imports[ti.PkgPath] = alias
		}
	}

	// 1. Print package and imports
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)
	fmt.Fprintf(&buf, "import (\n")
	fmt.Fprintf(&buf, "\t\"github.com/podhmo/veritas\"\n")
	for path, alias := range imports {
		fmt.Fprintf(&buf, "\t%s \"%s\"\n", alias, path)
	}
	fmt.Fprintf(&buf, ")\n\n")

	// 2. Print init function for registering rules
	fmt.Fprintf(&buf, "func init() {\n")
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
	fmt.Fprintf(&buf, "}\n\n")

	// 3. Print GetRegisteredTypes function
	fmt.Fprintf(&buf, "// GetRegisteredTypes returns a list of all types that have validation rules.\n")
	fmt.Fprintf(&buf, "func GetRegisteredTypes() []any {\n")
	fmt.Fprintf(&buf, "\treturn []any{\n")
	for _, ti := range typeInfos {
		alias := imports[ti.PkgPath]
		fmt.Fprintf(&buf, "\t\t%s.%s{},\n", alias, ti.Name)
	}
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "}\n")

	// 4. Format and write the output
	source := buf.Bytes()
	formatted, err := format.Source(source)
	if err != nil {
		return fmt.Errorf("failed to format generated go code: %w\nraw code:\n%s", err, string(source))
	}

	_, err = w.Write(formatted)
	return err
}
