package veritas

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/go-cmp/cmp"
)

func TestNewEngine(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	engine, err := NewEngine(logger)

	if err != nil {
		t.Fatalf("NewEngine() unexpected error: %v", err)
	}
	if engine == nil {
		t.Fatal("NewEngine() returned nil engine")
	}
	if engine.env == nil {
		t.Error("NewEngine() did not initialize CEL environment")
	}
	if engine.cache == nil {
		t.Error("NewEngine() did not initialize cache")
	}
	if engine.logger == nil {
		t.Error("NewEngine() did not initialize logger")
	}
}

func TestEngine_getProgram(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	engine, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	rule := `1 < 2`

	// 1. Cache miss
	logBuf.Reset()
	prog1, err1 := engine.getProgram(rule)
	if err1 != nil {
		t.Fatalf("getProgram() first call failed: %v", err1)
	}
	if prog1 == nil {
		t.Fatal("getProgram() first call returned nil program")
	}
	if !strings.Contains(logBuf.String(), "cache miss") {
		t.Errorf("Expected 'cache miss' log, got: %s", logBuf.String())
	}

	// 2. Cache hit
	logBuf.Reset()
	prog2, err2 := engine.getProgram(rule)
	if err2 != nil {
		t.Fatalf("getProgram() second call failed: %v", err2)
	}
	if !strings.Contains(logBuf.String(), "cache hit") {
		t.Errorf("Expected 'cache hit' log, got: %s", logBuf.String())
	}

	// To check if the programs are functionally the same, we can evaluate them
	// and compare the results.
	out1, _, err := prog1.Eval(cel.NoVars())
	if err != nil {
		t.Fatalf("prog1.Eval() failed: %v", err)
	}
	out2, _, err := prog2.Eval(cel.NoVars())
	if err != nil {
		t.Fatalf("prog2.Eval() failed: %v", err)
	}
	if diff := cmp.Diff(out1, out2); diff != "" {
		t.Errorf("Program evaluation results mismatch (-want +got):\n%s", diff)
	}

	// 3. Invalid expression
	_, err3 := engine.getProgram(`1 <`)
	if err3 == nil {
		t.Fatal("getProgram() with invalid rule did not return an error")
	}
	// We only check that an error is returned, as the exact message can change.
	if err3 == nil {
		t.Fatal("getProgram() with invalid rule did not return an error")
	}
}
