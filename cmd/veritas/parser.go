package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"
	"reflect"
	"strings"

	"github.com/podhmo/veritas"
	"golang.org/x/tools/go/packages"
)

var shorthandCELMap = map[string]any{
	"required": "self != nil",
	"nonzero": map[string]string{
		"string": `self != ""`,
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

type Parser struct {
	logger *slog.Logger
}

func NewParser(logger *slog.Logger) *Parser {
	return &Parser{logger: logger}
}

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
					ruleSet := veritas.ValidationRuleSet{
						FieldRules: make(map[string][]string),
					}

					if doc := genDecl.Doc; doc != nil {
						for _, comment := range doc.List {
							if strings.HasPrefix(comment.Text, "// @cel:") {
								rule := strings.TrimSpace(strings.TrimPrefix(comment.Text, "// @cel:"))
								ruleSet.TypeRules = append(ruleSet.TypeRules, rule)
							}
						}
					}

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
						}
					}

					if len(ruleSet.TypeRules) > 0 || len(ruleSet.FieldRules) > 0 {
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
	var celRules []string
	groups := p.splitRuleGroups(rawRules)
	for _, group := range groups {
		cel, err := p.parseRuleGroup(group, tv, "self")
		if err != nil {
			p.logger.Warn("failed to parse rule group", "group", group, "error", err)
			continue
		}
		if cel != "" {
			celRules = append(celRules, cel)
		}
	}
	return celRules
}

func (p *Parser) splitRuleGroups(rawRules []string) [][]string {
	var groups [][]string
	currentGroup := []string{}
	for _, r := range rawRules {
		trimmed := strings.TrimSpace(r)
		if (trimmed == "keys" || trimmed == "values") && len(currentGroup) > 0 {
			groups = append(groups, currentGroup)
			currentGroup = []string{}
		}
		currentGroup = append(currentGroup, trimmed)
	}
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}
	return groups
}

func (p *Parser) parseRuleGroup(group []string, tv types.Type, baseVar string) (string, error) {
	if len(group) == 0 {
		return "", nil
	}

	token := group[0]
	switch token {
	case "dive":
		slice, ok := tv.Underlying().(*types.Slice)
		if !ok {
			return "", fmt.Errorf("'dive' is only applicable to slices, but got %s", tv.String())
		}
		subCEL, err := p.parseRuleGroup(group[1:], slice.Elem(), "x")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.all(x, %s)", baseVar, subCEL), nil
	case "keys":
		m, ok := tv.Underlying().(*types.Map)
		if !ok {
			return "", fmt.Errorf("'keys' is only applicable to maps, but got %s", tv.String())
		}
		subCEL, err := p.parseRuleGroup(group[1:], m.Key(), "k")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.all(k, %s)", baseVar, subCEL), nil
	case "values":
		m, ok := tv.Underlying().(*types.Map)
		if !ok {
			return "", fmt.Errorf("'values' is only applicable to maps, but got %s", tv.String())
		}
		subCEL, err := p.parseRuleGroup(group[1:], m.Elem(), "v")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.all(v, %s)", baseVar, subCEL), nil
	default:
		// Base case: list of shorthands
		var conditions []string
		for _, shorthand := range group {
			if shorthand == "" {
				continue
			}
			cel := p.shorthandToCEL(shorthand, tv, baseVar)
			if cel != "" {
				conditions = append(conditions, cel)
			}
		}
		return strings.Join(conditions, " && "), nil
	}
}

func (p *Parser) shorthandToCEL(shorthand string, tv types.Type, varName string) string {
	if strings.HasPrefix(shorthand, "cel:") {
		return strings.ReplaceAll(strings.TrimPrefix(shorthand, "cel:"), "self", varName)
	}

	cel, ok := shorthandCELMap[shorthand]
	if !ok {
		p.logger.Warn("unsupported validation shorthand", "shorthand", shorthand)
		return ""
	}

	var exprTpl string
	switch v := cel.(type) {
	case string:
		exprTpl = v
	case map[string]string:
		typeCategory := p.categorizeType(tv)
		var tplOk bool
		exprTpl, tplOk = v[typeCategory]
		if !tplOk {
			p.logger.Warn("shorthand not applicable for type category", "shorthand", shorthand, "category", typeCategory)
			return ""
		}
	}
	return strings.ReplaceAll(exprTpl, "self", varName)
}

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
		return "ptr"
	default:
		return "other"
	}
}
