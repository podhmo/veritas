package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/podhmo/veritas/lint"
	"github.com/podhmo/veritas/lint/required"
	"golang.org/x/tools/go/analysis/multichecker"
)

func TestMain(t *testing.T) {
	t.Run("lint", func(t *testing.T) {
		// We expect the linter to exit with a non-zero status code, which will be
		// caught as an error by the testing framework.
		var err error
		var stderr string
		func() {
			// The multichecker calls os.Exit(1) on linting errors, which we need to prevent
			// from terminating the test. We can't easily capture the exit code, but we
			// can capture the stderr output.
			defer func() {
				if r := recover(); r != nil {
					// This is a crude way to catch the os.Exit(1) call.
					// A more robust solution would involve a custom test runner.
				}
			}()
			_, stderr = captureOutput(func() {
				multichecker.Main(
					lint.Analyzer,
					required.Analyzer,
				)
			})
		}()

		// Check if the linter produced any output.
		if len(stderr) == 0 && err == nil {
			t.Errorf("expected an error or lint output, but got none")
		}

		if !strings.Contains(stderr, "required' tag can only be used with pointer types") {
			t.Errorf("expected to find a lint error, but got %q", stderr)
		}
	})

	t.Run("gen", func(t *testing.T) {
		// Build the veritas command
		cmd := exec.Command("go", "build", "-o", "veritas_test", ".")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to build veritas command: %v", err)
		}
		defer os.Remove("veritas_test")
		// Run the generator
		cmd = exec.Command("./veritas_test", "github.com/podhmo/veritas/cmd/veritas/gen/testdata/src/a")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			t.Fatalf("failed to run veritas command: %v, stderr: %s", err, stderr.String())
		}

		// check stdout has "veritas.Register"
		if !strings.Contains(stdout.String(), "veritas.Register") {
			t.Errorf("expected to find veritas.Register in stdout, but got %q", stdout.String())
		}
	})
}

func captureOutput(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	f()

	wOut.Close()
	wErr.Close()
	out, _ := io.ReadAll(rOut)
	err, _ := io.ReadAll(rErr)
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return string(out), string(err)
}
