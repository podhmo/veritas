# Plan for One-File Code Injection

This document outlines the plan for adding a `-inject` option to `cmd/veritas` for one-file code generation. This feature will allow rewriting a specific function within a single Go source file, or appending it if it doesn't exist.

## 1. Goal

The primary goal is to enable `cmd/veritas` to modify a single Go file by replacing the body of a target function, `setupValidation`. If the function does not exist, it will be appended to the end of the file. This is particularly useful for projects contained within a single file.

Additionally, the standard code generation will be updated to call this `setupValidation` function from an `init()` function.

## 2. High-Level Plan

1.  **Introduce a new `-inject` flag:** This flag will take a filename as an argument and signal to `veritas` that it should perform an in-place modification of that file.
2.  **Locate the target function:** The tool will parse the target Go file and find the `setupValidation` function.
3.  **Replace or Append:**
    *   If `setupValidation` is found, its body will be replaced with the new, generated code.
    *   If `setupValidation` is not found, the entire function will be generated and appended to the end of the file.
4.  **Preserve formatting and comments:** The modification will be done carefully to preserve the original file's formatting.
5.  **Update standard code generation:** The default code generation will be modified to include an `init()` function that calls `setupValidation()`.
6.  **Create a new example:** A new example directory, `examples/codegen-onefile`, will be created to demonstrate and test this feature.
7.  **Create `TODO.md`:** A `TODO.md` file will be created to track the implementation tasks.

## 3. Implementation Details

### 3.1. Command Line Interface

A new `-inject` flag will be added to `cmd/veritas`.

```bash
veritas generate -inject <target-file>.go
```

### 3.2. AST Parsing and Code Generation

The implementation will use the `go/parser` and `go/ast` packages.

1.  **Parse the file:** `parser.ParseFile` will be used to get the AST.
2.  **Find the function:** The AST will be traversed to find the `FuncDecl` for `setupValidation`.
3.  **Generate the function:** A new `setupValidation` function will be generated as an `ast.FuncDecl`.
4.  **Replace or Append:**
    *   **If found:** The body of the existing function will be replaced with the body of the generated function. The `go/printer` package will be used to render the modified AST. To preserve comments, we will replace the `Body` of the `FuncDecl` and print the entire AST back to the file.
    *   **If not found:** The generated `setupValidation` function will be appended to the file. A simple string append to the original source might be the most reliable way to do this to avoid breaking formatting.
5.  **Update `init()` function:** The code that generates the standard validation logic will be updated to also generate an `init()` function:
    ```go
    func init() {
        setupValidation()
    }
    ```

## 4. Testing Plan

1.  **Unit Tests:**
    *   Test the AST manipulation logic for both replacing and appending the function.
2.  **Integration Tests:**
    *   **Injection:**
        *   Create a test case in `examples/codegen-onefile` where `main.go` **has** a `setupValidation` function. Run `veritas generate -inject main.go` and verify the body is replaced.
        *   Create another test case where `main.go` **does not have** a `setupValidation` function. Run `veritas generate -inject main.go` and verify the function is appended.
    *   **Standard Generation:**
        *   Run the standard code generation and verify that the output contains the `init()` function and the call to `setupValidation()`.
3.  **Edge Cases:**
    *   Test with an empty Go file.
    *   Test with a file that has syntax errors.

## 5. New Example: `examples/codegen-onefile`

This will be created as planned, with `main.go` and `main_test.go` to serve as a testbed for the injection feature.

## 6. TODO.md

A `TODO.md` file will be created with the following tasks:

*   Implement the `-inject` flag in `cmd/veritas`.
*   Implement the AST parsing logic to find `setupValidation`.
*   Implement the function body replacement logic.
*   Implement the function appending logic.
*   Update standard code generation to include `init()` and `setupValidation()`.
*   Create the `examples/codegen-onefile` directory and files.
*   Add unit and integration tests.
*   Update documentation.

## 7. Implementation Details

The final implementation uses a hybrid approach that combines AST analysis with text-based replacement to achieve a robust and format-preserving injection mechanism.

1.  **AST Parsing:** The target file is parsed using `go/parser` to locate the `setupValidation`, `GetKnownTypes`, and `init` functions.
2.  **Code Generation:** The bodies of these functions are generated as strings.
3.  **Text-based Replacement:** The original source file is read, and the lines corresponding to the target functions (identified via their positions in the `token.FileSet`) are replaced with the newly generated code. If a function does not exist, it is appended to the end of the file.
4.  **Goimports:** After the text-based replacement, `golang.org/x/tools/imports.Process` is used to format the resulting code and automatically manage imports. This ensures that the final output is clean, correctly formatted, and has the necessary imports.

This approach successfully avoids the pitfalls of re-printing the entire AST (which can lead to undesirable formatting changes) while still leveraging the power of the AST for accurate function location.

## 8. Resolving the "Chicken and Egg" Problem

A remaining challenge is the "chicken and egg" problem where `go generate` fails if the target file calls a function that `go generate` is supposed to create. Here are several potential solutions:

### Solution 1: Two-Pass Generation (Stateful Generator)

The generator could perform a "dry run" pass first.

1.  **Pass 1 (Analysis & Stubbing):** The generator runs, but instead of generating the full function bodies, it injects empty or stubbed versions of `setupValidation` and `GetKnownTypes`. This would satisfy the Go compiler, allowing the initial analysis to pass without errors.
2.  **Pass 2 (Full Generation):** The generator runs a second time. Now that the stubs exist, the type-checker and parser will succeed, allowing the generator to correctly analyze the source and replace the stubs with the fully-generated function bodies.

*   **Pros:** Conceptually simple, works within the existing `go/analysis` framework.
*   **Cons:** Slower (runs analysis twice), feels like a workaround.

### Solution 2: Pre-emptive AST Manipulation

Before passing the code to the `go/analysis` framework, manually parse the AST and comment out or remove the problematic function calls (e.g., `veritas.WithTypes(GetKnownTypes()...)`).

1.  Read the source file into memory.
2.  Use `go/parser` to get the AST.
3.  Use `astutil.Apply` or a similar visitor to find the specific call expression and "comment it out" (e.g., by replacing it with a no-op or literally rewriting that part of the source buffer).
4.  Pass the modified source buffer to the analysis runner.
5.  After generation, the final `imports.Process` step would re-format the code, and the user's original (now valid) call would remain.

*   **Pros:** Avoids the two-pass overhead.
*   **Cons:** Very complex and brittle. Manipulating the source text accurately before analysis is difficult.

### Solution 3: Configure `packages.Load` to Ignore Errors

The underlying `packages.Load` function, used by the analysis framework, could potentially be configured to ignore certain types of errors.

1.  Investigate the `packages.Config` struct for flags that might allow the loader to proceed even with "undefined" errors.
2.  If such a configuration exists, the generator could load the package, ignore the specific error about `GetKnownTypes`, and proceed with generation.

*   **Pros:** Potentially the cleanest solution if it works.
*   **Cons:** May not be possible, or it might suppress other important errors. `go/analysis` might not expose this level of configuration.

### Solution 4: Dummy `validation.go` File

The generator could create a temporary, dummy `validation.go` file in the target package before running the main analysis.

1.  Before analysis, create a `dummy_validation.go` file in the same directory.
2.  This file would contain empty implementations of `setupValidation()` and `GetKnownTypes()`.
3.  Run the analysis. The Go toolchain will see the dummy file and the types will be defined.
4.  After the main injection is complete, delete the `dummy_validation.go` file.

*   **Pros:** Relatively simple to implement, doesn't require complex AST manipulation of the target file.
*   **Cons:** Involves creating and deleting temporary files, which can be messy and have side effects in some environments (e.g., triggering file watchers).
