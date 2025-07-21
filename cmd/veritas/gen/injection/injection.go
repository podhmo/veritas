package injection

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/podhmo/veritas"
	parser_ "github.com/podhmo/veritas/cmd/veritas/parser"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

// Inject injects the validation setup function into the target file.
func Inject(
	targetFile string,
	pkgName string,
	ruleSets map[string]veritas.ValidationRuleSet,
	knownTypes []parser_.TypeInfo,
) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, targetFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", targetFile, err)
	}

	// Generate the setupValidation and GetKnownTypes functions
	var setupValidationBuf bytes.Buffer
	if err := generateSetupValidation(&setupValidationBuf, ruleSets); err != nil {
		return fmt.Errorf("failed to generate setupValidation: %w", err)
	}

	var getKnownTypesBuf bytes.Buffer
	if err := generateGetKnownTypes(&getKnownTypesBuf, pkgName, knownTypes); err != nil {
		return fmt.Errorf("failed to generate GetKnownTypes: %w", err)
	}

	// Replace or append the functions
	originalContent, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("failed to read original file %s: %w", targetFile, err)
	}

	newContent, err := replaceOrAppendFunction(
		fset,
		node,
		string(originalContent),
		"setupValidation",
		setupValidationBuf.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to replace or append setupValidation: %w", err)
	}

	// Re-parse the AST to get the correct node for the next replacement
	fset = token.NewFileSet()
	node, err = parser.ParseFile(fset, targetFile, newContent, parser.ParseComments)
	if err != nil {
		// If parsing fails, it might be because the content is now just a string.
		// Let's try parsing the string content directly.
		node, err = parser.ParseFile(fset, "", newContent, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to re-parse file after setupValidation injection: %w", err)
		}
	}

	newContent, err = replaceOrAppendFunction(
		fset,
		node,
		newContent,
		"GetKnownTypes",
		getKnownTypesBuf.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to replace or append GetKnownTypes: %w", err)
	}

	// Generate and inject init function
	var initBuf bytes.Buffer
	if err := generateInit(&initBuf); err != nil {
		return fmt.Errorf("failed to generate init: %w", err)
	}

	// Re-parse the AST to get the correct node for the next replacement
	fset = token.NewFileSet()
	node, err = parser.ParseFile(fset, "", newContent, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to re-parse file after GetKnownTypes injection: %w", err)
	}

	newContent, err = replaceOrAppendFunction(
		fset,
		node,
		newContent,
		"init",
		initBuf.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to replace or append init: %w", err)
	}

	// Use imports.Process to format and add/remove imports
	formattedContent, err := imports.Process(targetFile, []byte(newContent), nil)
	if err != nil {
		return fmt.Errorf("processing (goimports) generated code for %s: %w\nOriginal newContent was:\n%s", targetFile, err, newContent)
	}

	if err := os.WriteFile(targetFile, formattedContent, 0644); err != nil {
		return fmt.Errorf("writing modified content to %s: %w", targetFile, err)
	}

	return nil
}

func replaceOrAppendFunction(
	fset *token.FileSet,
	node *ast.File,
	originalContent string,
	funcName string,
	newFuncContent string,
) (string, error) {
	var funcNode *ast.FuncDecl
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		if fn, ok := cursor.Node().(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			funcNode = fn
			return false
		}
		return true
	}, nil)

	if funcNode == nil {
		// Append new function to the end of the file.
		var builder strings.Builder
		builder.WriteString(originalContent)
		if len(originalContent) > 0 && !strings.HasSuffix(originalContent, "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("\n") // Add an extra newline for separation
		builder.WriteString(newFuncContent)
		return builder.String(), nil
	}

	// Replace existing function.
	originalLines := strings.Split(originalContent, "\n")

	var startNode ast.Node = funcNode
	if funcNode.Doc != nil && len(funcNode.Doc.List) > 0 {
		startNode = funcNode.Doc
	}
	startLine := fset.Position(startNode.Pos()).Line
	endLine := fset.Position(funcNode.End()).Line

	var builder strings.Builder
	for i := 0; i < startLine-1; i++ {
		builder.WriteString(originalLines[i])
		builder.WriteString("\n")
	}

	builder.WriteString(newFuncContent)
	if !strings.HasSuffix(newFuncContent, "\n") {
		builder.WriteString("\n")
	}

	if endLine < len(originalLines) {
		for i := endLine; i < len(originalLines); i++ {
			builder.WriteString(originalLines[i])
			if i < len(originalLines)-1 {
				builder.WriteString("\n")
			}
		}
	}

	return builder.String(), nil
}

func generateSetupValidation(w io.Writer, ruleSets map[string]veritas.ValidationRuleSet) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "func setupValidation() {\n")
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

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format setupValidation: %w\n---\n%s", err, buf.String())
	}
	_, err = w.Write(formatted)
	return err
}

func generateInit(w io.Writer) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "func init() {\n")
	fmt.Fprintf(&buf, "\tsetupValidation()\n")
	fmt.Fprintf(&buf, "}\n")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format init: %w\n---\n%s", err, buf.String())
	}
	_, err = w.Write(formatted)
	return err
}

func generateGetKnownTypes(w io.Writer, pkgName string, knownTypes []parser_.TypeInfo) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "// GetKnownTypes returns a list of all types that have validation rules.\n")
	fmt.Fprintf(&buf, "func GetKnownTypes() []any {\n")
	fmt.Fprintf(&buf, "\treturn []any{\n")
	for _, t := range knownTypes {
		if t.PackageName == pkgName {
			fmt.Fprintf(&buf, "\t\t%s{},\n", t.TypeName)
		} else {
			fmt.Fprintf(&buf, "\t\t%s.%s{},\n", t.PackageName, t.TypeName)
		}
	}
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "}\n")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format GetKnownTypes: %w\n---\n%s", err, buf.String())
	}
	_, err = w.Write(formatted)
	return err
}
