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

	"github.com/podhmo/veritas"
	parser_ "github.com/podhmo/veritas/cmd/veritas/parser"
	"golang.org/x/tools/go/ast/astutil"
)

// Inject injects the validation setup function into the target file.
func Inject(
	targetFile string,
	pkgName string,
	ruleSets map[string]veritas.ValidationRuleSet,
	knownTypes []parser_.TypeInfo,
) error {
	fset := token.NewFileSet()
	// NOTE: The third argument is src, and it can be nil if the file is specified by filename.
	node, err := parser.ParseFile(fset, targetFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", targetFile, err)
	}

	// Generate the body of the setupValidation function
	var setupBuf bytes.Buffer
	if err := generateSetupValidationBody(&setupBuf, ruleSets, knownTypes); err != nil {
		return fmt.Errorf("failed to generate function body for setupValidation: %w", err)
	}
	setupBody, err := parseFunctionBody(fset, setupBuf.String())
	if err != nil {
		return fmt.Errorf("failed to parse generated body for setupValidation: %w", err)
	}
	upsertFunc(node, "setupValidation", setupBody)

	// Generate GetKnownTypes function
	var getKnownTypesBuf bytes.Buffer
	if err := generateGetKnownTypes(&getKnownTypesBuf, pkgName, knownTypes); err != nil {
		return fmt.Errorf("failed to generate function body for GetKnownTypes: %w", err)
	}
	getKnownTypesDecl, err := parseFuncDecl(fset, getKnownTypesBuf.String())
	if err != nil {
		return fmt.Errorf("failed to parse generated body for GetKnownTypes: %w", err)
	}
	upsertFuncDecl(node, "GetKnownTypes", getKnownTypesDecl)

	// Write the modified AST back to the file
	var outBuf bytes.Buffer
	if err := format.Node(&outBuf, fset, node); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}
	if err := os.WriteFile(targetFile, outBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", targetFile, err)
	}

	return nil
}

func parseFunctionBody(fset *token.FileSet, body string) ([]ast.Stmt, error) {
	bodySrc := fmt.Sprintf("package main\nfunc temp() {\n%s\n}", body)
	bodyFile, err := parser.ParseFile(fset, "", bodySrc, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated body: %w\n---\n%s", err, body)
	}
	return bodyFile.Decls[0].(*ast.FuncDecl).Body.List, nil
}

func upsertFunc(node *ast.File, name string, body []ast.Stmt) {
	var found bool
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		fn, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || fn.Name.Name != name {
			return true
		}
		found = true
		fn.Body.List = body
		return false // Stop searching
	}, nil)

	if !found {
		// Create the function declaration
		fn := &ast.FuncDecl{
			Name: ast.NewIdent(name),
			Type: &ast.FuncType{
				Params:  &ast.FieldList{},
				Results: nil,
			},
			Body: &ast.BlockStmt{
				List: body,
			},
		}
		node.Decls = append(node.Decls, fn)
	}
}

func parseFuncDecl(fset *token.FileSet, src string) (*ast.FuncDecl, error) {
	fileSrc := fmt.Sprintf("package main\n%s", src)
	file, err := parser.ParseFile(fset, "", fileSrc, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated decl: %w\n---\n%s", err, src)
	}
	return file.Decls[0].(*ast.FuncDecl), nil
}

func upsertFuncDecl(node *ast.File, name string, decl *ast.FuncDecl) {
	var found bool
	astutil.Apply(node, func(cursor *astutil.Cursor) bool {
		fn, ok := cursor.Node().(*ast.FuncDecl)
		if !ok || fn.Name.Name != name {
			return true
		}
		found = true
		*fn = *decl
		return false // Stop searching
	}, nil)

	if !found {
		node.Decls = append(node.Decls, decl)
	}
}

func generateGetKnownTypes(w io.Writer, pkgName string, knownTypes []parser_.TypeInfo) error {
	fmt.Fprintf(w, "// GetKnownTypes returns a list of all types that have validation rules.\n")
	fmt.Fprintf(w, "func GetKnownTypes() []any {\n")
	fmt.Fprintf(w, "return []any{\n")
	for _, t := range knownTypes {
		if t.PackageName == pkgName {
			fmt.Fprintf(w, "%s{},\n", t.TypeName)
		} else {
			fmt.Fprintf(w, "%s.%s{},\n", t.PackageName, t.TypeName)
		}
	}
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "}\n")
	return nil
}

func generateSetupValidationBody(w io.Writer, ruleSets map[string]veritas.ValidationRuleSet, knownTypes []parser_.TypeInfo) error {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(ruleSets))
	for k := range ruleSets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ruleSet := ruleSets[key]
		fmt.Fprintf(w, "veritas.Register(\"%s\", veritas.ValidationRuleSet{\n", key)
		if len(ruleSet.TypeRules) > 0 {
			fmt.Fprintf(w, "TypeRules: []string{\n")
			for _, rule := range ruleSet.TypeRules {
				fmt.Fprintf(w, "`%s`,\n", rule)
			}
			fmt.Fprintf(w, "},\n")
		}
		if len(ruleSet.FieldRules) > 0 {
			fmt.Fprintf(w, "FieldRules: map[string][]string{\n")
			fieldKeys := make([]string, 0, len(ruleSet.FieldRules))
			for fk := range ruleSet.FieldRules {
				fieldKeys = append(fieldKeys, fk)
			}
			sort.Strings(fieldKeys)
			for _, fk := range fieldKeys {
				fmt.Fprintf(w, "\"%s\": {\n", fk)
				for _, rule := range ruleSet.FieldRules[fk] {
					fmt.Fprintf(w, "`%s`,\n", rule)
				}
				fmt.Fprintf(w, "},\n")
			}
			fmt.Fprintf(w, "},\n")
		}
		fmt.Fprintf(w, "})\n")
	}
	return nil
}
