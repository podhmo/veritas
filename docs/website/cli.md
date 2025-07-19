# CLI Reference

The `veritas` command-line tool has two main functions: generating validation code and linting your validation rules.

## Code Generation (`veritas`)

This is the default mode of the `veritas` tool. It scans your Go source files for struct annotations and generates Go code containing your validation rules.

### Usage

The generator is typically run via `go generate`.

```go
//go:generate go run github.com/podhmo/veritas/cmd/veritas [flags]
```

### Flags

-   `-gen-type <TypeName>`: **(Required)** The name of the struct type to generate rules for.
-   `-out-name <filename.go>`: **(Required)** The name of the Go file to be generated.

### Example

```go
// file: models/user.go
package models

//go:generate go run github.com/podhmo/veritas/cmd/veritas -gen-type=User -out-name=veritas_gen.go

type User struct {
    // ...
}
```

Running `go generate ./...` in your project will execute this command, creating `veritas_gen.go` in the `models` package.

## Linting (`veritas -lint`)

The linter statically checks your `rules.json` files for common errors.

### Usage

```bash
veritas -lint [packages]
```

The linter recursively finds and analyzes `rules.json` files in the specified packages. If no packages are provided, it defaults to the current directory (`.`).

### Checks

The linter performs the following checks:

1.  **Valid CEL Syntax**: Ensures that all `TypeRules` and `FieldRules` are syntactically correct CEL expressions.
2.  **Field Existence**: Verifies that every field specified in a `FieldRules` map actually exists in the corresponding Go struct.

### Example

To run the linter on your entire project:

```bash
go run github.com/podhmo/veritas/cmd/veritas -lint ./...
```
