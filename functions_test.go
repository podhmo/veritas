package veritas

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCustomFunctions(t *testing.T) {
	t.Parallel()

	// We don't care about the logger output for this test, but Handler must not be nil.
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create an engine with our custom functions.
	engine, err := NewEngine(logger, DefaultFunctions()...)
	if err != nil {
		t.Fatalf("NewEngine() with custom functions failed: %v", err)
	}

	tests := []struct {
		name     string
		rule     string
		vars     map[string]any
		want     any
		wantErr  bool
		errEquil func(a, b error) bool
	}{
		{
			name: "strings.ToUpper success",
			rule: `strings.ToUpper(name)`,
			vars: map[string]any{"name": "gopher"},
			want: "GOPHER",
		},
		{
			name: "matches success",
			rule: `custom.matches(email, '^[^@]+@[^@]+\\.[^@]+$')`,
			vars: map[string]any{"email": "test@example.com"},
			want: true,
		},
		{
			name: "matches failure",
			rule: `custom.matches(email, '^[^@]+@[^@]+\\.[^@]+$')`,
			vars: map[string]any{"email": "not-an-email"},
			want: false,
		},
		{
			name:    "matches invalid regexp",
			rule:    `custom.matches(email, '[')`, // Invalid regexp
			vars:    map[string]any{"email": "test@example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var envOpts []cel.EnvOption
			if tt.vars != nil {
				for k := range tt.vars {
					// For this test, we know all variables are strings.
					// A more robust implementation would use reflection.
					envOpts = append(envOpts, cel.Variable(k, cel.StringType))
				}
			}

			prog, err := engine.getProgram(tt.rule, envOpts...)
			if err != nil {
				t.Fatalf("getProgram() failed: %v", err)
			}

			out, _, err := prog.Eval(tt.vars)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Eval() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				// For error cases, we might just check that an error occurred.
				// More specific error checking can be added if needed.
				return
			}

			got := out.Value()
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Eval() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
