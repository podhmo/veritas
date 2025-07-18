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

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Engine is not used directly, but its creation could be part of a setup.
	_, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("NewEngine() failed: %v", err)
	}

	tests := []struct {
		name    string
		rule    string
		vars    map[string]any
		want    any
		wantErr bool
	}{
		{
			name: "strings.ToUpper success",
			rule: `strings.ToUpper(name)`,
			vars: map[string]any{"name": "gopher"},
			want: "GOPHER",
		},
		{
			name: "matches success",
			rule: `matches(email, '^[^@]+@[^@]+\\.[^@]+$')`,
			vars: map[string]any{"email": "test@example.com"},
			want: true,
		},
		{
			name: "matches failure",
			rule: `matches(email, '^[^@]+@[^@]+\\.[^@]+$')`,
			vars: map[string]any{"email": "not-an-email"},
			want: false,
		},
		{
			name:    "matches invalid regexp",
			rule:    `matches(email, '[')`,
			vars:    map[string]any{"email": "test@example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a new environment for each test, including the default functions.
			envOpts := DefaultFunctions()

			if tt.vars != nil {
				for k := range tt.vars {
					envOpts = append(envOpts, cel.Variable(k, cel.StringType))
				}
			}

			env, err := cel.NewEnv(envOpts...)
			if err != nil {
				t.Fatalf("cel.NewEnv() failed: %v", err)
			}

			ast, issues := env.Compile(tt.rule)
			if issues != nil && issues.Err() != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("Compile() failed: %v", issues.Err())
			}

			prog, err := env.Program(ast)
			if err != nil {
				t.Fatalf("env.Program() failed: %v", err)
			}

			out, _, err := prog.Eval(tt.vars)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Eval() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got := out.Value()
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Eval() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
