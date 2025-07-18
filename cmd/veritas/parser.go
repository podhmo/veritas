package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"reflect"
	"strings"

	"github.com/podhmo/veritas"
)

// Parser is responsible for parsing Go source files to extract validation rules.
type Parser struct {
	logger *slog.Logger
}

// NewParser creates a new parser.
func NewParser(logger *slog.Logger) *Parser {
	return &Parser{logger: logger}
}

// Parse scans the given path for Go source files and extracts validation
// rules from struct tags and special comments.
func (p *Parser) Parse(path string) (map[string]veritas.ValidationRuleSet, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	ruleSets := make(map[string]veritas.ValidationRuleSet)

	ast.Inspect(f, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structName := typeSpec.Name.Name
			p.logger.Debug("found struct", "name", structName)

			ruleSet := veritas.ValidationRuleSet{
				FieldRules: make(map[string][]string),
			}

			// Extract type-level rules from comments associated with the GenDecl
			if doc := genDecl.Doc; doc != nil {
				p.logger.Debug("struct doc comment", "struct", structName, "doc", doc.Text())
				for _, comment := range doc.List {
					if strings.HasPrefix(comment.Text, "// @cel:") {
						rule := strings.TrimSpace(strings.TrimPrefix(comment.Text, "// @cel:"))
						ruleSet.TypeRules = append(ruleSet.TypeRules, rule)
						p.logger.Debug("found type rule", "struct", structName, "rule", rule)
					}
				}
			}

			// Extract field-level rules from tags
			for _, field := range structType.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
				validateTag, ok := tag.Lookup("validate")
				if !ok {
					continue
				}
				fieldName := field.Names[0].Name
				rules := strings.Split(validateTag, ",")
				ruleSet.FieldRules[fieldName] = rules
				p.logger.Debug("found field rules", "struct", structName, "field", fieldName, "rules", rules)
			}

			if len(ruleSet.TypeRules) > 0 || len(ruleSet.FieldRules) > 0 {
				ruleSets[structName] = ruleSet
			}
		}

		return true
	})

	return ruleSets, nil
}
