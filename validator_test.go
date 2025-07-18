package veritas

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// MockUser is a test struct.
type MockUser struct {
	Name  string
	Email string
	Age   int
}

// MockAddress is a nested struct.
type MockAddress struct {
	Street string
	City   string
}

// MockProfile has a nested struct.
type MockProfile struct {
	Title   string
	Address *MockAddress
}

// MockUserWithProfile contains a nested struct with validation rules.
type MockUserWithProfile struct {
	Name    string
	Email   string
	Age     int
	Profile *MockProfile
}

// mockUserAdapter converts a MockUser object (or a pointer to it) to a map.
func mockUserAdapter(obj any) (map[string]any, error) {
	var user *MockUser
	switch v := obj.(type) {
	case MockUser:
		user = &v
	case *MockUser:
		user = v
	default:
		return nil, fmt.Errorf("unsupported type for MockUser adapter: %T", obj)
	}

	return map[string]any{
		"Name":  user.Name,
		"Email": user.Email,
		"Age":   user.Age,
	}, nil
}

// mockUserWithProfileAdapter converts a MockUserWithProfile to a map.
func mockUserWithProfileAdapter(obj any) (map[string]any, error) {
	var user *MockUserWithProfile
	switch v := obj.(type) {
	case MockUserWithProfile:
		user = &v
	case *MockUserWithProfile:
		user = v
	default:
		return nil, fmt.Errorf("unsupported type for adapter: %T", obj)
	}
	return map[string]any{
		"Name":    user.Name,
		"Email":   user.Email,
		"Age":     user.Age,
		"Profile": user.Profile, // Keep the nested struct as is for now
	}, nil
}

// mockProfileAdapter converts a MockProfile to a map.
func mockProfileAdapter(obj any) (map[string]any, error) {
	var p *MockProfile
	switch v := obj.(type) {
	case MockProfile:
		p = &v
	case *MockProfile:
		p = v
	default:
		return nil, fmt.Errorf("unsupported type for adapter: %T", obj)
	}
	return map[string]any{
		"Title":   p.Title,
		"Address": p.Address,
	}, nil
}

// mockAddressAdapter converts a MockAddress to a map.
func mockAddressAdapter(obj any) (map[string]any, error) {
	var a *MockAddress
	switch v := obj.(type) {
	case MockAddress:
		a = &v
	case *MockAddress:
		a = v
	default:
		return nil, fmt.Errorf("unsupported type for adapter: %T", obj)
	}
	return map[string]any{
		"Street": a.Street,
		"City":   a.City,
	}, nil
}

func TestValidator_Validate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine, err := NewEngine(logger, DefaultFunctions()...)
	if err != nil {
		t.Fatalf("NewEngine() failed: %v", err)
	}

	provider := NewJSONRuleProvider("testdata/rules/user.json")

	// Define the adapters for the types we want to validate.
	adapters := map[string]TypeAdapter{
		"MockUser":              mockUserAdapter,
		"MockUserWithProfile":   mockUserWithProfileAdapter,
		"MockProfile":           mockProfileAdapter,
		"MockAddress":           mockAddressAdapter,
		// Adapter for the unregistered type test case.
		"struct { Name string }": func(obj any) (map[string]any, error) {
			return map[string]any{}, nil
		},
	}

	// Create a new validator with the adapters.
	validator, err := NewValidator(engine, provider, logger, adapters)
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
			wantErr: errors.Join(NewValidationError("MockUser", "Email", `this.Email.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$')`)),
		},
		{
			name: "object with multiple errors",
			obj: &MockUser{
				Name:  "",
				Email: "invalid-email",
				Age:   10,
			},
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
			wantErr: errors.Join(NewValidationError("MockUser", "", "this.Age < 50")),
		},
		{
			name:    "unregistered type",
			obj:     struct{ Age int }{10},
			wantErr: NewFatalError("no TypeAdapter registered for type struct { Age int }"),
		},
		{
			name: "valid nested struct",
			obj: &MockUserWithProfile{
				Name:  "Gopher",
				Email: "gopher@golang.org",
				Age:   10,
				Profile: &MockProfile{
					Title: "Engineer",
					Address: &MockAddress{
						Street: "123 Go Street",
						City:   "Gopherville",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid nested struct field",
			obj: &MockUserWithProfile{
				Name:  "Gopher",
				Email: "gopher@golang.org",
				Age:   10,
				Profile: &MockProfile{
					Title: "", // Fails validation rule
					Address: &MockAddress{
						Street: "123 Go Street",
						City:   "Gopherville",
					},
				},
			},
			wantErr: NewValidationError("MockProfile", "Title", "this.Title.size() > 0"),
		},
		{
			name: "invalid deep nested struct field",
			obj: &MockUserWithProfile{
				Name:  "Gopher",
				Email: "gopher@golang.org",
				Age:   10,
				Profile: &MockProfile{
					Title: "Engineer",
					Address: &MockAddress{
						Street: "", // Fails validation rule
						City:   "Gopherville",
					},
				},
			},
			wantErr: NewValidationError("MockAddress", "Street", "this.Street.size() > 0"),
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
				nameRuleError := NewValidationError("MockUser", "Name", "this.Name.size() > 0").Error()
				emailRuleError := NewValidationError("MockUser", "Email", `this.Email.matches('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$')`).Error()

				if !strings.Contains(errStr, nameRuleError) {
					t.Errorf("Validate() error missing expected content '%s' in '%s'", nameRuleError, errStr)
				}
				if !strings.Contains(errStr, emailRuleError) {
					t.Errorf("Validate() error missing expected content '%s' in '%s'", emailRuleError, errStr)
				}
				return
			}

			// Handle nil and non-nil error cases separately.
			if tt.wantErr == nil {
				if gotErr != nil {
					t.Errorf("Validate() got error = %v, want nil", gotErr)
				}
				return
			}

			if gotErr == nil {
				t.Errorf("Validate() got nil, want error = %v", tt.wantErr)
				return
			}

			if tt.wantErr.Error() != gotErr.Error() {
				t.Errorf("Validate() error mismatch\nwant: %s\ngot:  %s", tt.wantErr.Error(), gotErr.Error())
			}
		})
	}
}
