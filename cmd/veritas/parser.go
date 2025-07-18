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

// ShorthandCELMap provides a mapping from common validation shorthands to their
// corresponding CEL expressions. Note that the `self` in the expression will
// be replaced by `self.FieldName` during parsing.
var ShorthandCELMap = map[string]string{
	"required": "size(self) > 0", // Works for strings, slices, maps.
	"email":    "self.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$')",
}

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
				if len(field.Names) == 0 || field.Tag == nil {
					continue
				}
				fieldName := field.Names[0].Name
				tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
				validateTag, ok := tag.Lookup("validate")
				if !ok {
					continue
				}

				rawRules := strings.Split(validateTag, ",")
				celRules := make([]string, 0, len(rawRules))
				for _, r := range rawRules {
					trimmedRule := strings.TrimSpace(r)
					celExpr, isShorthand := ShorthandCELMap[trimmedRule]
					if isShorthand {
						// For shorthands, we prepend the field name to `self` to form the expression.
						// e.g., for field "Name" and shorthand "required", CEL becomes "self.Name != nil"
						celRules = append(celRules, strings.Replace(celExpr, "self", "self."+fieldName, 1))
					} else {
						// If not a shorthand, it's assumed to be a raw CEL expression.
						celRules = append(celRules, trimmedRule)
					}
				}

				if len(celRules) > 0 {
					ruleSet.FieldRules[fieldName] = celRules
					p.logger.Debug("found field rules", "struct", structName, "field", fieldName, "rules", celRules)
				}
			}

			if len(ruleSet.TypeRules) > 0 || len(ruleSet.FieldRules) > 0 {
				ruleSets[structName] = ruleSet
			}
		}

		return true
	})

	return ruleSets, nil
}
