package veritas

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// MockUser is a test struct.
type MockUser struct {
	Name  string
	Email string
	Age   int
}

func TestValidator_Validate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine, err := NewEngine(logger, DefaultFunctions()...)
	if err != nil {
		t.Fatalf("NewEngine() failed: %v", err)
	}

	provider := NewJSONRuleProvider("testdata/rules/user.json")

	validator, err := NewValidator(engine, provider, logger)
	if err != nil {
		t.Fatalf("NewValidator() failed: %v", err)
	}

	tests := []struct {
		name    string
		obj     any
		wantErr error
	}{
		{
			name: "valid object",
			obj: &MockUser{
				Name:  "Gopher",
				Email: "gopher@golang.org",
				Age:   10,
			},
			wantErr: nil,
		},
		{
			name: "object with invalid field",
			obj: &MockUser{
				Name:  "Gopher",
				Email: "invalid-email",
				Age:   10,
			},
			wantErr: NewValidationError("MockUser", "Email", `this.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$')`),
		},
		{
			name: "object with multiple errors",
			obj: &MockUser{
				Name:  "",
				Email: "invalid-email",
				Age:   10,
			},
			wantErr: errors.Join(
				NewValidationError("MockUser", "Name", "this.size() > 0"),
				NewValidationError("MockUser", "Email", `this.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$')`),
			),
		},
		{
			name: "object with type rule violation",
			obj: &MockUser{
				Name:  "Gopher",
				Email: "gopher@golang.org",
				Age:   99, // Fails the type-level rule "this.Age < 50"
			},
			wantErr: NewValidationError("MockUser", "", "this.Age < 50"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Validate(tt.obj)

			if diff := cmp.Diff(tt.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
