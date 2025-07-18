package veritas

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/google/cel-go/cel"
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
	if engine.baseOpts == nil {
		t.Error("NewEngine() did not initialize base options")
	}
	if engine.programCache == nil {
		t.Error("NewEngine() did not initialize program cache")
	}
	if engine.logger == nil {
		t.Error("NewEngine() did not initialize logger")
	}
}

func TestEngine_getProgram_Caching(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	engine, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Create a test environment
	env, err := cel.NewEnv()
	if err != nil {
		t.Fatalf("cel.NewEnv() failed: %v", err)
	}

	rule := `1 < 2`

	// Cache miss
	_, err = engine.getProgram(env, rule)
	if err != nil {
		t.Fatalf("getProgram() first call failed: %v", err)
	}

	// Cache hit
	_, err = engine.getProgram(env, rule)
	if err != nil {
		t.Fatalf("getProgram() second call failed: %v", err)
	}
}
