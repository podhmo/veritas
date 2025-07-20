# Plan for One-File Code Injection

This document outlines the plan for adding an "inject" option to `cmd/veritas` for one-file code generation. This feature will allow rewriting a specific function within a single Go source file, rather than regenerating the entire file.

## 1. Goal

The primary goal is to enable `cmd/veritas` to modify a single Go file by replacing the body of a target function. This is particularly useful for projects contained within a single file, where regenerating the whole file would be disruptive.

The initial target for this functionality will be a function named `setupValidation`.

## 2. High-Level Plan

1.  **Introduce a new `inject` command or option:** This will signal to `veritas` that it should perform an in-place modification of a file.
2.  **Locate the target function:** The tool will parse the target Go file and find the `setupValidation` function, which will be identified by a special marker comment.
3.  **Replace the function body:** The body of the `setupValidation` function will be replaced with new, generated code.
4.  **Preserve formatting and comments:** The modification will be done carefully to preserve the original file's formatting, including comments and line breaks. This will be achieved by using the AST to identify the exact start and end positions of the function body.
5.  **Create a new example:** A new example directory, `examples/codegen-onefile`, will be created to demonstrate and test this feature.

## 3. Implementation Details

### 3.1. Command Line Interface

A new subcommand or flag will be added to `cmd/veritas`. For example:

```bash
veritas generate --inject <target-file>.go
```

### 3.2. Marker Comment

A marker comment will be used to identify the function to be replaced. This makes the injection explicit and avoids accidental modification of the wrong function.

```go
// veritas:inject setupValidation
func setupValidation(b *validator.Builder) {
    // old body
}
```

### 3.3. AST Parsing and Code Generation

The implementation will use the `go/parser` and `go/ast` packages to parse the source file and build an AST.

1.  **Parse the file:** `parser.ParseFile` will be used to get the AST for the target file.
2.  **Find the function:** The AST will be traversed to find the `FuncDecl` node corresponding to `setupValidation`. The associated comment group will be inspected to find the `veritas:inject` marker.
3.  **Generate the new function body:** The new body for the `setupValidation` function will be generated as a string.
4.  **Reconstruct the file:** The new file content will be created by combining:
    *   The original source code from the beginning of the file up to the start of the function's body.
    *   The newly generated function body.
    *   The original source code from the end of the function's body to the end of the file.

The `go/printer` package will be used to print the new AST nodes to a buffer, ensuring correct formatting. To preserve comments and layout, we will be very careful with the source positions.

The logic would be roughly:

1.  Get the `pos` and `end` of the `ast.FuncDecl`.
2.  The new source is `original[:func.Body.Pos()]` + `newBody` + `original[func.Body.End():]`.

We must be careful with braces `{` and `}`.

## 4. Testing Plan

A comprehensive suite of tests will be developed to ensure the reliability of the injection feature.

1.  **Unit Tests:**
    *   Test the function locator to ensure it correctly finds the function with the marker comment.
    *   Test the code generator to ensure it produces the correct function body.
    *   Test the file reconstruction logic to ensure it correctly assembles the new file.
2.  **Integration Tests:**
    *   Create a test case in `examples/codegen-onefile` with a `main.go` file containing a `setupValidation` function.
    *   Run `veritas generate --inject` on this file and verify that the function body is correctly replaced.
    *   The test should check that the rest of the file remains unchanged.
    *   The modified file should be compilable and runnable.
3.  **Edge Cases:**
    *   Test with a file that does not contain the target function. The tool should do nothing and not report an error.
    *   Test with a file that has the marker comment but on a non-function declaration. This should be an error.
    *   Test with different formatting styles and comment placements around the target function.

## 5. New Example: `examples/codegen-onefile`

A new directory, `examples/codegen-onefile`, will be created with the following structure:

```
examples/codegen-onefile/
├── main.go
└── main_test.go
```

*   `main.go`: A simple, single-file application that uses the `setupValidation` function. This will be the target for the code injection.
*   `main_test.go`: A test file that verifies the behavior of the `setupValidation` function. This will be used to ensure that the code injection works as expected.

This example will serve as a living demonstration of the one-file injection feature and will be used in the integration tests.

## 6. Future Work

*   Support for injecting other types of declarations (e.g., `var`, `const`, `type`).
*   Support for injecting code into multiple files at once.
*   More sophisticated marker comments to control the injection process (e.g., specifying the name of the function to inject).
