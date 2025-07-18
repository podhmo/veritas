package veritas

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/veritas/testdata/sources"
)

// mockUserAdapter converts a MockUser object (or a pointer to it) to a map.
func mockUserAdapter(obj any) (map[string]any, error) {
	var user *sources.MockUser
	switch v := obj.(type) {
	case sources.MockUser:
		user = &v
	case *sources.MockUser:
		user = v
	default:
		return nil, fmt.Errorf("unsupported type for MockUser adapter: %T", obj)
	}

	return map[string]any{
		"Name":  user.Name,
		"Email": user.Email,
		"Age":   user.Age,
		"ID":    user.ID,
		"URL":   user.URL,
	}, nil
}

func boxAdapter(obj any) (map[string]any, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("boxAdapter supports only structs, got %T", obj)
	}

	field := val.FieldByName("Value")
	if !field.IsValid() {
		return nil, fmt.Errorf("no 'Value' field in %T", obj)
	}

	fieldVal := field.Interface()

	// Dereference pointers for CEL, but keep nil as is.
	rv := reflect.ValueOf(fieldVal)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		fieldVal = rv.Elem().Interface()
	}

	return map[string]any{
		"Value": fieldVal,
	}, nil
}

func itemAdapter(obj any) (map[string]any, error) {
	var item *sources.Item
	switch v := obj.(type) {
	case sources.Item:
		item = &v
	case *sources.Item:
		item = v
	default:
		return nil, fmt.Errorf("unsupported type for Item adapter: %T", obj)
	}
	return map[string]any{
		"Name": item.Name,
	}, nil
}

func embeddedUserAdapter(obj any) (map[string]any, error) {
	var u *sources.EmbeddedUser
	switch v := obj.(type) {
	case sources.EmbeddedUser:
		u = &v
	case *sources.EmbeddedUser:
		u = v
	default:
		return nil, fmt.Errorf("unsupported type for adapter: %T", obj)
	}
	return map[string]any{
		"ID":   u.ID,
		"Name": u.Name,
	}, nil
}

func complexUserAdapter(obj any) (map[string]any, error) {
	var user *sources.ComplexUser
	switch v := obj.(type) {
	case sources.ComplexUser:
		user = &v
	case *sources.ComplexUser:
		user = v
	default:
		return nil, fmt.Errorf("unsupported type for ComplexUser adapter: %T", obj)
	}

	// For simplicity in testing, we'll manually convert the map.
	// A real implementation might use reflection or other helpers.
	metadata := make(map[string]any)
	for k, v := range user.Metadata {
		metadata[k] = v
	}

	// Also handle the slice.
	scores := make([]any, len(user.Scores))
	for i, s := range user.Scores {
		scores[i] = s
	}

	return map[string]any{
		"Name":     user.Name,
		"Scores":   scores,
		"Metadata": metadata,
	}, nil
}

func profileAdapter(obj any) (map[string]any, error) {
	var p *sources.Profile
	switch v := obj.(type) {
	case sources.Profile:
		p = &v
	case *sources.Profile:
		p = v
	default:
		return nil, fmt.Errorf("unsupported type for Profile adapter: %T", obj)
	}
	return map[string]any{
		"Platform": p.Platform,
		"Handle":   p.Handle,
	}, nil
}

func userWithProfilesAdapter(obj any) (map[string]any, error) {
	var u *sources.UserWithProfiles
	switch v := obj.(type) {
	case sources.UserWithProfiles:
		u = &v
	case *sources.UserWithProfiles:
		u = v
	default:
		return nil, fmt.Errorf("unsupported type for UserWithProfiles adapter: %T", obj)
	}
	return map[string]any{
		"Name": u.Name,
		// Profiles and Contacts are not directly used in UserWithProfiles's own rules,
		// but the validator will recurse into them.
		"Profiles": u.Profiles,
		"Contacts": u.Contacts,
	}, nil
}


func TestValidator_Validate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	engine, err := NewEngine(logger, DefaultFunctions()...)
	if err != nil {
		t.Fatalf("NewEngine() failed: %v", err)
	}

	provider := NewJSONRuleProvider("testdata/rules/user.json")

	// Define the adapters for the types we want to validate.
	adapters := map[string]TypeAdapter{
		"sources.MockUser":         mockUserAdapter,
		"sources.EmbeddedUser":     embeddedUserAdapter,
		"sources.ComplexUser":      complexUserAdapter,
		"sources.Profile":          profileAdapter,
		"sources.UserWithProfiles": userWithProfilesAdapter,
		"sources.Box[T]":           boxAdapter,
		"sources.Item":             itemAdapter,
	}

	// Create a new validator with the adapters.
	validator, err := NewValidator(engine, provider, logger, adapters)
	if err != nil {
		t.Fatalf("NewValidator() failed: %v", err)
	}

	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name         string
		obj          any
		wantErr      error
		isMultiError bool // Flag for multi-error checks
	}{
		{
			name: "valid object",
			obj: &sources.MockUser{
				Name:  "Gopher",
				Email: "gopher@golang.org",
				Age:   20,
				ID:    intPtr(1),
			},
			wantErr: nil,
		},
		{
			name: "object with invalid field",
			obj: &sources.MockUser{
				Name:  "Gopher",
				Email: "invalid-email",
				Age:   20,
				ID:    intPtr(1),
			},
			wantErr: errors.Join(NewValidationError("sources.MockUser", "Email", `self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`)),
		},
		{
			name: "object with multiple errors",
			obj: &sources.MockUser{
				Name:  "", // Fails nonzero
				Email: "invalid-email",
				Age:   20,
				ID:    intPtr(1),
			},
			wantErr:      errors.New("multiple errors expected"),
			isMultiError: true,
		},
		{
			name: "object with type rule violation",
			obj: &sources.MockUser{
				Name:  "Gopher",
				Email: "gopher@golang.org",
				Age:   17, // Fails the type-level rule "self.Age >= 18"
				ID:    intPtr(1),
			},
			wantErr: errors.Join(NewValidationError("sources.MockUser", "", "self.Age >= 18")),
		},
		{
			name:    "unregistered type",
			obj:     struct{ Age int }{10},
			wantErr: NewFatalError("no TypeAdapter registered for type struct { Age int } or struct { Age int }"),
		},
		{
			name: "valid embedded struct",
			obj: &sources.EmbeddedUser{
				Base: sources.Base{ID: "ab"},
				Name: "Gopher",
			},
			wantErr: nil,
		},
		{
			name: "invalid embedded struct field",
			obj: &sources.EmbeddedUser{
				Base: sources.Base{ID: "a"}, // Fails size check
				Name: "Gopher",
			},
			wantErr: NewValidationError("sources.EmbeddedUser", "ID", `self != "" && self.size() > 1`),
		},
		{
			name: "invalid own struct field with embedded",
			obj: &sources.EmbeddedUser{
				Base: sources.Base{ID: "ab"},
				Name: "", // Fails required check
			},
			wantErr: NewValidationError("sources.EmbeddedUser", "Name", `self != ""`),
		},
		{
			name: "valid complex object",
			obj: &sources.ComplexUser{
				Name:   "ComplexGopher",
				Scores: []int{10, 20, 0},
				Metadata: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantErr: nil,
		},
		{
			name: "complex object with invalid slice element",
			obj: &sources.ComplexUser{
				Name:   "ComplexGopher",
				Scores: []int{10, -5, 0}, // -5 is invalid
				Metadata: map[string]string{
					"key1": "value1",
				},
			},
			wantErr: NewValidationError("sources.ComplexUser", "Scores", `self.all(x, x >= 0)`),
		},
		{
			name: "valid nested structs in slice and map",
			obj: &sources.UserWithProfiles{
				Name: "Gopher",
				Profiles: []sources.Profile{
					{Platform: "twitter", Handle: "gopher"},
				},
				Contacts: map[string]sources.Profile{
					"work": {Platform: "github", Handle: "golang"},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid struct in slice",
			obj: &sources.UserWithProfiles{
				Name: "Gopher",
				Profiles: []sources.Profile{
					{Platform: "twitter", Handle: "gopher"},
					{Platform: "twitter", Handle: "go"}, // Invalid handle
				},
			},
			wantErr:      NewValidationError("sources.Profile", "Handle", `self != "" && self.size() > 2`),
			isMultiError: true, // It's a single error, but let's check for its presence
		},
		{
			name: "invalid struct in map",
			obj: &sources.UserWithProfiles{
				Name: "Gopher",
				Contacts: map[string]sources.Profile{
					"personal": {Platform: "", Handle: "myhandle"}, // Invalid platform
				},
			},
			wantErr:      NewValidationError("sources.Profile", "Platform", `self != ""`),
			isMultiError: true, // It's a single error, but let's check for its presence
		},
		{
			name: "multiple errors in nested structs",
			obj: &sources.UserWithProfiles{
				Name: "Gopher",
				Profiles: []sources.Profile{
					{Platform: "", Handle: "gopher"}, // Invalid platform
				},
				Contacts: map[string]sources.Profile{
					"work": {Platform: "github", Handle: "go"}, // Invalid handle
				},
			},
			isMultiError: true, // Expecting two distinct validation errors
		},
		{
			name: "valid generic struct with string",
			obj: &sources.Box[string]{
				Value: "hello",
			},
			wantErr: nil,
		},
		{
			name: "invalid generic struct with nil pointer",
			obj: &sources.Box[*string]{
				Value: nil,
			},
			wantErr: NewValidationError("sources.Box[T]", "Value", `self != null`),
		},
		{
			name: "valid generic struct with struct pointer",
			obj: &sources.Box[*sources.Item]{
				Value: &sources.Item{Name: "valid-item"},
			},
			wantErr: nil,
		},
		{
			name: "invalid generic struct with invalid nested struct",
			obj: &sources.Box[*sources.Item]{
				Value: &sources.Item{Name: ""}, // name is required
			},
			wantErr:      NewValidationError("sources.Item", "Name", `self != ""`),
			isMultiError: true,
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

				// Special handling for the multi-error tests
				switch tt.name {
				case "object with multiple errors":
					nameRuleError := NewValidationError("sources.MockUser", "Name", `self != ""`).Error()
					emailRuleError := NewValidationError("sources.MockUser", "Email", `self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`).Error()
					if !strings.Contains(errStr, nameRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", nameRuleError, errStr)
					}
					if !strings.Contains(errStr, emailRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", emailRuleError, errStr)
					}
				case "invalid struct in slice":
					handleRuleError := NewValidationError("sources.Profile", "Handle", `self != "" && self.size() > 2`).Error()
					if !strings.Contains(errStr, handleRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", handleRuleError, errStr)
					}
				case "invalid struct in map":
					platformRuleError := NewValidationError("sources.Profile", "Platform", `self != ""`).Error()
					if !strings.Contains(errStr, platformRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", platformRuleError, errStr)
					}
				case "multiple errors in nested structs":
					platformRuleError := NewValidationError("sources.Profile", "Platform", `self != ""`).Error()
					handleRuleError := NewValidationError("sources.Profile", "Handle", `self != "" && self.size() > 2`).Error()
					if !strings.Contains(errStr, platformRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", platformRuleError, errStr)
					}
					if !strings.Contains(errStr, handleRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", handleRuleError, errStr)
					}
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
