# CLI Reference

The `veritas` command-line tool has two main functions: generating validation code and linting your validation rules. The behavior is controlled by the `-lint` flag.

## Code Generation (Default)

By default, the `veritas` tool scans your Go source files for struct annotations and generates Go code containing your validation rules.

### Usage

The generator is typically run via `go generate`.

```go
//go:generate go run github.com/podhmo/veritas/cmd/veritas [flags]
```

### Flags

-   `-o <filename.go>`: **(Required)** The name of the Go file to be generated.
-   `-pkg <package>`: The package name to use for the generated file. If not specified, it defaults to the package of the directory containing the file.

### Example

```go
// file: models/user.go
package models

//go:generate go run github.com/podhmo/veritas/cmd/veritas -o veritas_gen.go

type User struct {
    // ...
}
```

Running `go generate ./...` in your project will execute this command, creating `veritas_gen.go` in the `models` package.

## Linting (`-lint` flag)

When the `-lint` flag is provided, the tool switches to linting mode. It statically checks your validation rules for common errors.

### Usage

```bash
go run github.com/podhmo/veritas/cmd/veritas -lint [packages]
```

The linter recursively finds and analyzes Go files in the specified packages. If no packages are provided, it defaults to the current directory (`.`).

### Checks

The linter performs the following checks:

1.  **Valid CEL Syntax**: Ensures that all `TypeRules` and `FieldRules` are syntactically correct CEL expressions.
2.  **Field Existence**: Verifies that every field specified in a `FieldRules` map actually exists in the corresponding Go struct.

### Example

To run the linter on your entire project:

```bash
go run github.com/podhmo/veritas/cmd/veritas -lint ./...
```
