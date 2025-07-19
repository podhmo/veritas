# Analyzer Tuning

The current implementation of the `cmd/veritas/gen` analyzer is inefficient because it reloads package information that is already available.

## Current Situation

1.  The `gen` package's `run` function receives a `*codegen.Pass` object, which contains the fully loaded `*packages.Package`.
2.  The `gen` package extracts only the package path (`pass.Pkg.Path()`) and passes it to the `parser.Parse` function.
3.  The `parser.Parse` function then calls `packages.Load` to reload the package information, which is a redundant and time-consuming operation.

## Investigation of `parser` Package Dependencies

I have analyzed the `cmd/veritas/parser/parser.go` file and determined that the following fields from `packages.Package` are used:

*   `pkg.Syntax`: Used to iterate over the AST files in the package to find type declarations and struct definitions.
*   `pkg.PkgPath`: Used to create a unique, fully qualified name for each type, which is essential for registering the validation rule sets correctly.
*   `pkg.TypesInfo`: Used to obtain detailed type information for struct fields, which is necessary for applying the correct validation rules (e.g., distinguishing between a `string` and an `int`). It's also used to resolve embedded structs.
*   `pkg.Types`: Used in conjunction with `pkg.TypesInfo` to resolve and inspect type definitions, particularly to check if an embedded struct is defined within the same package.

## Proposed Refactoring

To eliminate the inefficiency of reloading packages, I will refactor the code as follows:

1.  **Define a new data structure:** In the `parser` package, define a new struct named `PackageInfo` that encapsulates the necessary data from `*packages.Package`.

    ```go
    // PackageInfo contains the necessary information from a package for parsing.
    type PackageInfo struct {
        PkgPath   string
        Syntax    []*ast.File
        TypesInfo *types.Info
        Types     *types.Package
    }
    ```

2.  **Create a new parsing function:** Create a new function in the `parser` package, `ParseDirectly`, that accepts the `PackageInfo` struct.

    ```go
    func (p *Parser) ParseDirectly(info PackageInfo) (map[string]veritas.ValidationRuleSet, error) {
        // ... parsing logic using info ...
    }
    ```

3.  **Update the `gen` package:** Modify the `run` function in the `gen` package to populate the `PackageInfo` struct from the `*codegen.Pass` object and call the new `parser.ParseDirectly` function.

This refactoring will avoid the unnecessary `packages.Load` call, leading to a performance improvement in the `veritas-gen` tool.
