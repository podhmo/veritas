# Analyzer Tuning

The current implementation of the `cmd/veritas/gen` analyzer is inefficient because it reloads package information that is already available in the `codegen.Pass`.

## Current Situation

1.  The `gen` package's `run` function receives a `*codegen.Pass` object. This object already contains all the necessary type and syntax information.
2.  However, the `gen` package currently ignores this information and only passes the package path to the `parser.Parse` function.
3.  The `parser.Parse` function then calls `packages.Load` to reload all the package information, which is a redundant and time-consuming operation.

## Investigation of `parser` Package Dependencies

An analysis of `cmd/veritas/parser/parser.go` shows that the parsing logic relies on the following information, which is all available within the `codegen.Pass`:

*   **Package Path**: The parser needs the import path of the package to construct fully qualified type names. This can be obtained from `pass.Pkg.Path()`.
*   **Syntax Trees**: The parser iterates through the AST (`*ast.File`) of each file in the package. This is available in `pass.Files`.
*   **Type Information**: The parser heavily uses `*types.Info` to look up the types of expressions and fields. This is available in `pass.TypesInfo`.
*   **Package Types**: The parser uses `*types.Package` to resolve type objects and check their origin. This is available in `pass.Pkg`.

## Proposed Refactoring

To eliminate the inefficiency of reloading packages, the code will be refactored to pass the necessary information directly from the `gen` package to the `parser` package.

### 1. Define the Subset Struct (`PackageInfo`)

A new struct, `PackageInfo`, will be defined in the `parser` package. This struct will act as a data transfer object, holding the subset of information required for parsing, extracted from the `codegen.Pass`.

**Definition in `cmd/veritas/parser/parser.go`:**
```go
import (
	"go/ast"
	"go/types"
)

// PackageInfo contains the necessary information from a package for parsing.
// It is constructed in the gen package from a *codegen.Pass.
type PackageInfo struct {
	PkgPath   string
	Syntax    []*ast.File
	TypesInfo *types.Info
	Types     *types.Package
}
```

This struct is a direct subset of the information needed by the parser. The mapping from `*codegen.Pass` in the `gen` package is as follows:

| `PackageInfo` Field | Source in `*codegen.Pass` |
| ------------------- | ------------------------- |
| `PkgPath`           | `pass.Pkg.Path()`         |
| `Syntax`            | `pass.Files`              |
| `TypesInfo`         | `pass.TypesInfo`          |
| `Types`             | `pass.Pkg`                |

### 2. Create a New Parsing Function (`ParseDirectly`)

A new function, `ParseDirectly`, will be added to the `parser`. It will accept a `PackageInfo` object and perform the parsing, thus avoiding the `packages.Load` call.

**Signature in `cmd/veritas/parser/parser.go`:**
```go
// ParseDirectly parses validation rules from the given package information
// without loading packages itself.
func (p *Parser) ParseDirectly(info PackageInfo) (map[string]veritas.ValidationRuleSet, error) {
    // ... existing parsing logic, but using the 'info' struct ...
}
```

### 3. Update the `gen` Package

The `run` function in `cmd/veritas/gen/analyzer.go` will be modified to instantiate and populate the `PackageInfo` struct from the `*codegen.Pass` object and then call the new `parser.ParseDirectly` function.

**Example in `cmd/veritas/gen/analyzer.go`:**
```go
func run(pass *codegen.Pass) error {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	p := parser.NewParser(logger)

	// Create the PackageInfo struct from the pass
	info := parser.PackageInfo{
		PkgPath:   pass.Pkg.Path(),
		Syntax:    pass.Files,
		TypesInfo: pass.TypesInfo,
		Types:     pass.Pkg,
	}

	// Call the new direct parsing function
	ruleSets, err := p.ParseDirectly(info)
	if err != nil {
		return fmt.Errorf("failed to parse directly: %w", err)
	}

	// ... rest of the code generation logic ...
}
```

This refactoring will remove the redundant package loading step and significantly improve the performance and efficiency of the `veritas-gen` tool by reusing the already available analysis data.
