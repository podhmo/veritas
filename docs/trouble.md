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
