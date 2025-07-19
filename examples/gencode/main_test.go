package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	// capture output
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	// run main
	if err := run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// check output
	output := buf.String()
	if !strings.Contains(output, "validation ok") {
		t.Errorf("missing 'validation ok' message")
	}
	if !strings.Contains(output, "validation failed, as expected") {
		t.Errorf("missing 'validation failed, as expected' message")
	}
}
