- ## `go test` with multiple modules

- date: 2025-07-18
- author: jules

### problem

When running `go test ./...` in a repository with multiple modules, only the tests for the main module are executed. `go test all` or `go test ./... all` can be used to run tests for all modules in the workspace, but this can be slow and may include unnecessary external dependencies.

### solution

The most reliable way to test all modules in a workspace is to use `go.work` and run `go test` on each module individually. This can be done by adding a separate step for each module in the CI workflow.

For example, in `.github/workflows/ci.yml`:

```yaml
    - name: Test
      run: go test ./...
    - name: Test examples
      run: go test -C ./examples/http-server ./...
```

This ensures that each module is tested in its own context, with the correct dependencies.

- ## `go generate` with `singlegenerator`

- date: 2025-07-18
- author: jules

### problem

When trying to use `go:generate` with a command that uses the `singlegenerator` framework, the arguments are not parsed correctly. The `singlegenerator` framework does not seem to support passing arguments directly from the `go:generate` directive.

### solution

The solution is to create a separate `run.go` file that calls the `gen.Generator.Run` function directly. This allows you to bypass the `singlegenerator` framework and pass the arguments to the generator correctly.

The `run.go` file should look something like this:

```go
package main

import (
	"flag"
	"log"
	"os"

	"github.com/gostaticanalysis/codegen"
	"github.com/podhmo/veritas/cmd/veritas/gen"
	"golang.org/x/tools/go/packages"
)

func main() {
	var (
		pkgPath    string
		outputFile string
	)
	flag.StringVar(&pkgPath, "pkg", "./def", "package path to parse")
	flag.StringVar(&outputFile, "o", "", "output file path")
	flag.Parse()

	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		log.Fatalf("failed to load packages: %v", err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	var output *os.File
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}
		defer f.Close()
		output = f
	} else {
		output = os.Stdout
	}

	pass := &codegen.Pass{
		Pkg:    pkgs[0].Types,
		Output: output,
	}

	if err := gen.Generator.Run(pass); err != nil {
		log.Fatalf("failed to run generator: %v", err)
	}
}
```

The `go:generate` directive should then be updated to use the `run.go` file:

```go
//go:generate go run ./run.go -pkg=./def -o=./validation/validator.go
```
