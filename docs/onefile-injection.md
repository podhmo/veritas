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
