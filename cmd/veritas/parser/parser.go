package parser

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

// PackageInfo contains the necessary information from a package for parsing.
// It is constructed in the gen package from a *codegen.Pass.
type PackageInfo struct {
	PkgPath   string
	Syntax    []*ast.File
	TypesInfo *types.Info
	Types     *types.Package
}

var shorthandCELMap = map[string]any{
	"required": map[string]string{
		"ptr": "self != null",
	},
	"nonzero": map[string]string{
		"string": `self != ""`,
		"int":    "self != 0",
		"uint":   "self != 0",
		"float":  "self != 0.0",
		"ptr":    "self != null",
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

func (p *Parser) Parse(path string) (map[string]veritas.ValidationRuleSet, []TypeInfo, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, nil, fmt.Errorf("errors occurred while loading packages")
	}

	ruleSets := make(map[string]veritas.ValidationRuleSet)
	knownTypesMap := make(map[string]TypeInfo) // Use map to handle duplicates

	for _, pkg := range pkgs {
		info := PackageInfo{
			PkgPath:   pkg.PkgPath,
			Syntax:    pkg.Syntax,
			TypesInfo: pkg.TypesInfo,
			Types:     pkg.Types,
		}
		rules, types, err := p.ParseDirectly(info)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse directly for package %s: %w", pkg.PkgPath, err)
		}
		for k, v := range rules {
			ruleSets[k] = v
		}
		for _, ti := range types {
			knownTypesMap[ti.PackagePath+"."+ti.TypeName] = ti
		}
	}

	// Convert map back to slice
	var knownTypes []TypeInfo
	for _, ti := range knownTypesMap {
		knownTypes = append(knownTypes, ti)
	}

	return ruleSets, knownTypes, nil
}

// TypeInfo holds information about a type for which rules were generated.
type TypeInfo struct {
	PackagePath string
	PackageName string
	TypeName    string
}

// ParseDirectly parses validation rules from the given package information
// without loading packages itself.
func (p *Parser) ParseDirectly(info PackageInfo) (map[string]veritas.ValidationRuleSet, []TypeInfo, error) {
	ruleSets := make(map[string]veritas.ValidationRuleSet)
	var knownTypes []TypeInfo

	for _, f := range info.Syntax {
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

				var structNameBuilder strings.Builder
				structNameBuilder.WriteString(typeSpec.Name.Name)

				if typeSpec.TypeParams != nil && len(typeSpec.TypeParams.List) > 0 {
					structNameBuilder.WriteString("[")
					for i, p := range typeSpec.TypeParams.List {
						if i > 0 {
							structNameBuilder.WriteString(", ")
						}
						// p.Names is a list of identifiers, e.g., "T" in T any
						for j, name := range p.Names {
							if j > 0 {
								structNameBuilder.WriteString(", ")
							}
							structNameBuilder.WriteString(name.Name)
						}
					}
					structNameBuilder.WriteString("]")
				}
				structName := structNameBuilder.String()

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
				p.extractRulesForStruct(info, structType, &ruleSet)

				if len(ruleSet.TypeRules) > 0 || len(ruleSet.FieldRules) > 0 {
					fullTypeName := fmt.Sprintf("%s.%s", info.PkgPath, structName)
					ruleSets[fullTypeName] = ruleSet
					// NOTE: typeSpec.Name.Name does not include generic parameters.
					knownTypes = append(knownTypes, TypeInfo{
						PackagePath: info.PkgPath,
						PackageName: info.Types.Name(),
						TypeName:    typeSpec.Name.Name,
					})
				}
			}
			return true
		})
	}

	return ruleSets, knownTypes, nil
}

func (p *Parser) extractRulesForStruct(info PackageInfo, structType *ast.StructType, ruleSet *veritas.ValidationRuleSet) {
	for _, field := range structType.Fields.List {
		// Embedded field
		if field.Names == nil {
			if embeddedStruct, ok := p.getEmbeddedStruct(info, field.Type); ok {
				p.extractRulesForStruct(info, embeddedStruct, ruleSet)
			}
			continue
		}

		fieldName := field.Names[0].Name
		if field.Tag == nil {
			continue
		}
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		validateTag, ok := tag.Lookup("validate")
		if !ok {
			continue
		}

		tv := info.TypesInfo.TypeOf(field.Type)
		if tv == nil {
			p.logger.Warn("could not determine type for field", "field", fieldName)
			continue
		}

		rawRules := strings.Split(validateTag, ",")
		celRules, err := p.processRules(rawRules, tv)
		if err != nil {
			p.logger.Warn("error processing rules", "field", fieldName, "error", err)
			continue
		}
		if len(celRules) > 0 {
			ruleSet.FieldRules[fieldName] = celRules
		}
	}
}

func (p *Parser) getEmbeddedStruct(info PackageInfo, expr ast.Expr) (*ast.StructType, bool) {
	typeOf := info.TypesInfo.TypeOf(expr)
	if typeOf == nil {
		return nil, false
	}

	named, ok := typeOf.(*types.Named)
	if !ok {
		return nil, false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() != info.Types {
		// Not defined in the same package, handling this would require finding the package and file.
		// For now, we only support embedded structs from the same package.
		return nil, false
	}

	// Find the AST node for the type definition
	for _, f := range info.Syntax {
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Name.Name == obj.Name() {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						return structType, true
					}
				}
			}
		}
	}

	return nil, false
}

func (p *Parser) processRules(rawRules []string, tv types.Type) ([]string, error) {
	var celRules []string
	remaining := rawRules
	for len(remaining) > 0 {
		rule, nextRemaining, err := p.parseRule(remaining, tv)
		if err != nil {
			return nil, err
		}
		cel, err := rule.ToCEL()
		if err != nil {
			return nil, err
		}
		if cel != "" {
			celRules = append(celRules, cel)
		}
		remaining = nextRemaining
	}

	// Post-process to join non-directive rules
	var finalRules []string
	var simpleConditions []string

	for _, cel := range celRules {
		// This is a heuristic: directives like `all()` are assumed to be standalone.
		if strings.Contains(cel, ".all(") {
			if len(simpleConditions) > 0 {
				finalRules = append(finalRules, strings.Join(simpleConditions, " && "))
				simpleConditions = nil
			}
			finalRules = append(finalRules, cel)
		} else {
			simpleConditions = append(simpleConditions, cel)
		}
	}
	if len(simpleConditions) > 0 {
		finalRules = append(finalRules, strings.Join(simpleConditions, " && "))
	}

	return finalRules, nil
}

type Rule struct {
	TV        types.Type
	BaseVar   string
	Directive string
	SubRules  []string
	Nested    []*Rule
	parser    *Parser
}

func (r *Rule) ToCEL() (string, error) {
	if r.Directive == "" {
		var conditions []string
		for _, shorthand := range r.SubRules {
			if shorthand == "" {
				continue
			}
			cel := r.parser.shorthandToCEL(shorthand, r.TV, r.BaseVar)
			if cel != "" {
				conditions = append(conditions, cel)
			}
		}
		return strings.Join(conditions, " && "), nil
	}

	var varName string
	switch r.Directive {
	case "dive":
		varName = "x"
	case "keys":
		varName = "k"
	case "values":
		varName = "v"
	}

	var nestedCELs []string
	for _, nestedRule := range r.Nested {
		nestedRule.BaseVar = varName
		cel, err := nestedRule.ToCEL()
		if err != nil {
			return "", err
		}
		nestedCELs = append(nestedCELs, cel)
	}

	return fmt.Sprintf("%s.all(%s, %s)", r.BaseVar, varName, strings.Join(nestedCELs, " && ")), nil
}

func (p *Parser) parseRule(rawRules []string, tv types.Type) (*Rule, []string, error) {
	if len(rawRules) == 0 {
		return nil, nil, nil
	}
	token := strings.TrimSpace(rawRules[0])

	rule := &Rule{TV: tv, BaseVar: "self", parser: p}

	switch token {
	case "dive":
		slice, ok := tv.Underlying().(*types.Slice)
		if !ok {
			return nil, nil, fmt.Errorf("'dive' on non-slice type: %s", tv.String())
		}
		rule.Directive = "dive"

		var nestedRules []*Rule
		remaining := rawRules[1:]
		for len(remaining) > 0 {
			var nested *Rule
			var err error
			nested, remaining, err = p.parseRule(remaining, slice.Elem())
			if err != nil {
				return nil, nil, err
			}
			nestedRules = append(nestedRules, nested)
		}
		rule.Nested = nestedRules
		return rule, remaining, nil
	case "keys":
		m, ok := tv.Underlying().(*types.Map)
		if !ok {
			return nil, nil, fmt.Errorf("'keys' on non-map type: %s", tv.String())
		}
		rule.Directive = "keys"
		nested, remaining, err := p.parseRule(rawRules[1:], m.Key())
		if err != nil {
			return nil, nil, err
		}
		rule.Nested = []*Rule{nested}
		return rule, remaining, nil
	case "values":
		m, ok := tv.Underlying().(*types.Map)
		if !ok {
			return nil, nil, fmt.Errorf("'values' on non-map type: %s", tv.String())
		}
		rule.Directive = "values"
		nested, remaining, err := p.parseRule(rawRules[1:], m.Elem())
		if err != nil {
			return nil, nil, err
		}
		rule.Nested = []*Rule{nested}
		return rule, remaining, nil
	default:
		// Check if the first token starts a CEL expression.
		if strings.HasPrefix(token, "cel:") {
			// Find where the CEL expression ends. It might span multiple "tokens"
			// if there are commas within the CEL expression itself.
			// This is a simplification; a truly robust solution would need a more
			// sophisticated parser. For now, we assume CEL expressions don't contain
			// the 'dive', 'keys', or 'values' keywords and that they are the last rule.
			var celExprBuilder strings.Builder
			celExprBuilder.WriteString(strings.TrimPrefix(token, "cel:"))

			remaining := rawRules[1:]
			end := 0
			for i, t := range remaining {
				trimmed := strings.TrimSpace(t)
				if trimmed == "dive" || trimmed == "keys" || trimmed == "values" {
					break
				}
				celExprBuilder.WriteString(",")
				celExprBuilder.WriteString(t)
				end = i + 1
			}
			rule.SubRules = []string{"cel:" + celExprBuilder.String()}
			return rule, remaining[end:], nil
		}

		// Find end of shorthands if not a CEL expression
		end := 0
		for i, t := range rawRules {
			trimmed := strings.TrimSpace(t)
			if trimmed == "dive" || trimmed == "keys" || trimmed == "values" || strings.HasPrefix(trimmed, "cel:") {
				break
			}
			end = i + 1
		}
		rule.SubRules = rawRules[:end]
		return rule, rawRules[end:], nil
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
	// Keep resolving named types until we get to the underlying type.
	for {
		named, ok := tv.(*types.Named)
		if !ok {
			break
		}
		tv = named.Underlying()
	}

	switch t := tv.(type) {
	case *types.TypeParam:
		// For a generic type parameter `T`, we can't know the concrete type at parse time.
		// We'll treat it like a pointer for shorthands like `required` -> `!= null`.
		// This is a reasonable default for `any` constraints.
		return "ptr"
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
