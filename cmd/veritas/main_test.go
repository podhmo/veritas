package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	// Build the veritas command
	cmd := exec.Command("go", "build", "-o", "veritas_test", ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build veritas command: %v", err)
	}
	defer os.Remove("veritas_test")

	t.Run("lint", func(t *testing.T) {
		cmd := exec.Command("./veritas_test", "-lint", "github.com/podhmo/veritas/lint/testdata/src/a")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()

		if err == nil {
			t.Errorf("expected an error, but got none")
		}

		if !strings.Contains(stderr.String(), "invalid type rule for User: ERROR: <input>:1:4: Syntax error: mismatched input '<EOF>' expecting {'[', '{', '(', '.', '-', '!', 'true', 'false', 'null', NUM_FLOAT, NUM_INT, NUM_UINT, STRING, BYTES, IDENTIFIER}") {
			t.Errorf("expected to find a lint error, but got %q", stderr.String())
		}
	})

	t.Run("gen", func(t *testing.T) {
		// Run the generator
		cmd := exec.Command("./veritas_test", "github.com/podhmo/veritas/cmd/veritas/gen/testdata/src/a")
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
