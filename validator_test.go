package veritas

import (
	"errors"
	"log/slog"
	"os"
	"strings"
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

	// Create a new validator, registering the MockUser type.
	validator, err := NewValidator(engine, provider, logger, MockUser{})
	if err != nil {
		t.Fatalf("NewValidator() failed: %v", err)
	}

	tests := []struct {
		name         string
		obj          any
		wantErr      error
		isMultiError bool // Flag for multi-error checks
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
			wantErr: NewValidationError("MockUser", "Email", `this.Email.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$')`),
		},
		{
			name: "object with multiple errors",
			obj: &MockUser{
				Name:  "",
				Email: "invalid-email",
				Age:   10,
			},
			// We can't compare joined errors directly due to unpredictable order.
			// So we check for the presence of each error message.
			wantErr:      errors.New("multiple errors expected"),
			isMultiError: true,
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

			if tt.isMultiError {
				if gotErr == nil {
					t.Fatalf("Validate() expected errors, got nil")
				}
				errStr := gotErr.Error()
				// Check that both expected error messages are present.
				nameRule := "this.Name.size() > 0"
				emailRule := `this.Email.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$')`
				if !strings.Contains(errStr, nameRule) {
					t.Errorf("Validate() error missing expected content '%s' in '%s'", nameRule, errStr)
				}
				if !strings.Contains(errStr, emailRule) {
					t.Errorf("Validate() error missing expected content '%s' in '%s'", emailRule, errStr)
				}
			} else {
				if diff := cmp.Diff(tt.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
					t.Errorf("Validate() error mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
