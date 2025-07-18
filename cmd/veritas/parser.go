package main

import (
	"fmt"
	"go/ast"
	"go/token" // Imported
	"go/types"
	"log/slog"
	"reflect"
	"strings"

	"github.com/podhmo/veritas"
	"golang.org/x/tools/go/packages"
)

// shorthandCELMap defines the mapping from a shorthand validation tag to its
// corresponding CEL expression. The value can be a simple string for a general rule,
// or a map[string]string for type-specific rules.
var shorthandCELMap = map[string]any{
	"required": "self != nil", // General rule for pointers, interfaces, etc.
	"nonzero": map[string]string{
		"string": "self != \"\"",
		"int":    "self != 0",
		"uint":   "self != 0",
		"float":  "self != 0.0",
		"ptr":    "self != nil",
		"slice":  "self.size() > 0",
		"map":    "self.size() > 0",
		"bool":   "self",
	},
	"email": `self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`,
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
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("errors occurred while loading packages")
	}

	ruleSets := make(map[string]veritas.ValidationRuleSet)

	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
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
					p.logger.Debug("found struct", "name", structName, "package", pkg.PkgPath)

					ruleSet := veritas.ValidationRuleSet{
						FieldRules: make(map[string][]string),
					}

					// Extract type-level rules from comments
					if doc := genDecl.Doc; doc != nil {
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
						if field.Tag == nil || len(field.Names) == 0 {
							continue
						}
						fieldName := field.Names[0].Name
						tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
						validateTag, ok := tag.Lookup("validate")
						if !ok {
							continue
						}

						tv := pkg.TypesInfo.TypeOf(field.Type)
						if tv == nil {
							p.logger.Warn("could not determine type for field", "field", fieldName)
							continue
						}

						rawRules := strings.Split(validateTag, ",")
						celRules := p.processRules(rawRules, tv)

						if len(celRules) > 0 {
							ruleSet.FieldRules[fieldName] = celRules
							p.logger.Debug("found field rules", "struct", structName, "field", fieldName, "rules", celRules)
						}
					}

					if len(ruleSet.TypeRules) > 0 || len(ruleSet.FieldRules) > 0 {
						// Use fully qualified name for uniqueness
						fullTypeName := fmt.Sprintf("%s.%s", pkg.Name, structName)
						ruleSets[fullTypeName] = ruleSet
					}
				}
				return true
			})
		}
	}

	return ruleSets, nil
}

func (p *Parser) processRules(rawRules []string, tv types.Type) []string {
	celRules := make([]string, 0, len(rawRules))
	for _, rule := range rawRules {
		trimmedRule := strings.TrimSpace(rule)
		if trimmedRule == "" {
			continue
		}

		if strings.HasPrefix(trimmedRule, "cel:") {
			celRules = append(celRules, strings.TrimPrefix(trimmedRule, "cel:"))
			continue
		}

		cel, ok := shorthandCELMap[trimmedRule]
		if !ok {
			p.logger.Warn("unsupported validation shorthand", "shorthand", trimmedRule)
			continue
		}

		switch v := cel.(type) {
		case string:
			celRules = append(celRules, v)
		case map[string]string:
			typeCategory := p.categorizeType(tv)
			if expr, ok := v[typeCategory]; ok {
				celRules = append(celRules, expr)
			} else {
				p.logger.Warn("shorthand not applicable for type category", "shorthand", trimmedRule, "category", typeCategory)
			}
		}
	}
	return celRules
}

// categorizeType determines the general category of a type for rule mapping.
func (p *Parser) categorizeType(tv types.Type) string {
	switch t := tv.Underlying().(type) {
	case *types.Basic:
		switch {
		case t.Info()&types.IsString != 0:
			return "string"
		case t.Info()&types.IsInteger != 0:
			if t.Info()&types.IsUnsigned != 0 {
				return "uint"
			}
			return "int"
		case t.Info()&types.IsFloat != 0:
			return "float"
		case t.Info()&types.IsBoolean != 0:
			return "bool"
		default:
			return "other"
		}
	case *types.Pointer:
		return "ptr"
	case *types.Slice:
		return "slice"
	case *types.Map:
		return "map"
	case *types.Interface:
		return "ptr" // Treat interfaces like pointers for nil checks
	default:
		return "other"
	}
}
