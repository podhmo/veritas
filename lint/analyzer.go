package lint

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"

	"github.com/google/cel-go/cel"
	"github.com/podhmo/veritas"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "veritas",
	Doc:  "veritas is a linter for veritas rules",
	Run:  run,
}

func loadRules(pass *analysis.Pass) (map[string]veritas.ValidationRuleSet, error) {
	if len(pass.Files) == 0 {
		return make(map[string]veritas.ValidationRuleSet), nil
	}

	// find rules.json from the directory of the first file
	dir := filepath.Dir(pass.Fset.File(pass.Files[0].Pos()).Name())

	var rulesPath string
	for {
		path := filepath.Join(dir, "rules.json")
		if _, err := os.Stat(path); err == nil {
			rulesPath = path
			break
		}
		if dir == filepath.Dir(dir) {
			// root directory
			break
		}
		dir = filepath.Dir(dir)
	}

	if rulesPath == "" {
		return make(map[string]veritas.ValidationRuleSet), nil
	}

	b, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var ruleSets map[string]veritas.ValidationRuleSet
	if err := json.Unmarshal(b, &ruleSets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}
	return ruleSets, nil
}

func run(pass *analysis.Pass) (interface{}, error) {
	rules, err := loadRules(pass)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}

	st, err := cel.NewEnv(veritas.DefaultFunctions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cel env: %w", err)
	}
	env, err := st.Extend(cel.Variable("self", cel.DynType))
	if err != nil {
		return nil, fmt.Errorf("failed to extend cel env: %w", err)
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			structType, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}

			typeName := ts.Name.Name
			ruleSet, ok := rules[typeName]
			if !ok {
				return true
			}

			// Check TypeRules
			for _, rule := range ruleSet.TypeRules {
				if _, issues := env.Compile(rule); issues != nil && issues.Err() != nil {
					pass.Reportf(ts.Pos(), "invalid type rule for %s: %s", typeName, issues.Err())
				}
			}

			// Check FieldRules (CEL syntax)
			for fieldName, fieldRules := range ruleSet.FieldRules {
				for _, rule := range fieldRules {
					if _, issues := env.Compile(rule); issues != nil && issues.Err() != nil {
						pass.Reportf(ts.Pos(), "invalid field rule for %s.%s: %s", typeName, fieldName, issues.Err())
					}
				}
			}

			// Check FieldRules (field existence)
			definedFields := make(map[string]bool)
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					definedFields[name.Name] = true
				}
			}

			for fieldName := range ruleSet.FieldRules {
				if !definedFields[fieldName] {
					pass.Reportf(ts.Pos(), "field %s in rules for %s does not exist in struct", fieldName, typeName)
				}
			}
			return true
		})
	}

	return nil, nil
}
