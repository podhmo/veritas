package veritas

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

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

	return map[string]any{
		"Value": field.Interface(),
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

func passwordAdapter(obj any) (map[string]any, error) {
	var p *sources.Password
	switch v := obj.(type) {
	case sources.Password:
		p = &v
	case *sources.Password:
		p = v
	default:
		return nil, fmt.Errorf("unsupported type for Password adapter: %T", obj)
	}
	return map[string]any{
		"Value": p.Value,
	}, nil
}

func anotherUserAdapter(obj any) (map[string]any, error) {
	var user *sources.AnotherUser
	switch v := obj.(type) {
	case sources.AnotherUser:
		user = &v
	case *sources.AnotherUser:
		user = v
	default:
		return nil, fmt.Errorf("unsupported type for AnotherUser adapter: %T", obj)
	}

	return map[string]any{
		"Name":  user.Username, // Note the field name mapping
		"Email": user.Email,
		"Age":   99, // Mocking a default value
		"ID":    -1, // Mocking a default value
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
	adapters := map[reflect.Type]TypeAdapterTarget{
		reflect.TypeOf(sources.Password{}):         {TargetName: "sources.Password", Adapter: passwordAdapter},
		reflect.TypeOf(&sources.Password{}):        {TargetName: "sources.Password", Adapter: passwordAdapter},
		reflect.TypeOf(sources.MockUser{}):         {TargetName: "sources.MockUser", Adapter: mockUserAdapter},
		reflect.TypeOf(&sources.MockUser{}):        {TargetName: "sources.MockUser", Adapter: mockUserAdapter},
		reflect.TypeOf(sources.EmbeddedUser{}):     {TargetName: "sources.EmbeddedUser", Adapter: embeddedUserAdapter},
		reflect.TypeOf(&sources.EmbeddedUser{}):    {TargetName: "sources.EmbeddedUser", Adapter: embeddedUserAdapter},
		reflect.TypeOf(sources.ComplexUser{}):      {TargetName: "sources.ComplexUser", Adapter: complexUserAdapter},
		reflect.TypeOf(&sources.ComplexUser{}):     {TargetName: "sources.ComplexUser", Adapter: complexUserAdapter},
		reflect.TypeOf(sources.Profile{}):          {TargetName: "sources.Profile", Adapter: profileAdapter},
		reflect.TypeOf(&sources.Profile{}):         {TargetName: "sources.Profile", Adapter: profileAdapter},
		reflect.TypeOf(sources.UserWithProfiles{}): {TargetName: "sources.UserWithProfiles", Adapter: userWithProfilesAdapter},
		reflect.TypeOf(&sources.UserWithProfiles{}): {TargetName: "sources.UserWithProfiles", Adapter: userWithProfilesAdapter},
		reflect.TypeOf(sources.Item{}):             {TargetName: "sources.Item", Adapter: itemAdapter},
		reflect.TypeOf(&sources.Item{}):            {TargetName: "sources.Item", Adapter: itemAdapter},

		// generic types
		reflect.TypeOf(sources.Box[string]{}):    {TargetName: "sources.Box[T]", Adapter: boxAdapter},
		reflect.TypeOf(&sources.Box[string]{}):   {TargetName: "sources.Box[T]", Adapter: boxAdapter},
		reflect.TypeOf(sources.Box[*string]{}):   {TargetName: "sources.Box[T]", Adapter: boxAdapter},
		reflect.TypeOf(&sources.Box[*string]{}):  {TargetName: "sources.Box[T]", Adapter: boxAdapter},
		reflect.TypeOf(sources.Box[*int]{}):      {TargetName: "sources.Box[T]", Adapter: boxAdapter},
		reflect.TypeOf(&sources.Box[*int]{}):     {TargetName: "sources.Box[T]", Adapter: boxAdapter},
		reflect.TypeOf(sources.Box[*sources.Item]{}):   {TargetName: "sources.Box[T]", Adapter: boxAdapter},
		reflect.TypeOf(&sources.Box[*sources.Item]{}):  {TargetName: "sources.Box[T]", Adapter: boxAdapter},

		// type mapping
		reflect.TypeOf(sources.AnotherUser{}): {TargetName: "sources.MockUser", Adapter: anotherUserAdapter},
		reflect.TypeOf(&sources.AnotherUser{}): {TargetName: "sources.MockUser", Adapter: anotherUserAdapter},
	}

	// Create a new validator with the adapters.
	validator, err := NewValidator(
		WithEngine(engine),
		WithRuleProvider(provider),
		WithLogger(logger),
		WithTypeAdapters(adapters),
	)
	if err != nil {
		t.Fatalf("NewValidator() failed: %v", err)
	}

	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name         string
		obj          any
		ctx          context.Context // Add context to test cases
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
			ctx:     context.Background(),
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
			ctx:     context.Background(),
			wantErr: errors.Join(NewValidationError("sources.MockUser", "Email", `self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`)),
		},
		{
			name: "object with multiple errors",
			obj: &sources.MockUser{
				Name:  "", // Fails nonzero
				Email: "invalid-email",
				Age:   20,
				ID:    nil, // Fails required
			},
			ctx:          context.Background(),
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
			ctx:     context.Background(),
			wantErr: errors.Join(NewValidationError("sources.MockUser", "", "self.Age >= 18")),
		},
		{
			name:    "unregistered type",
			obj:     struct{ Age int }{10},
			ctx:     context.Background(),
			wantErr: nil, // No adapter means no CEL validation, but recursion should still happen. No rules, so no error.
		},
		{
			name: "type mapping with adapter",
			obj: &sources.AnotherUser{ // This type is not in the rules.json
				Username: "gopher-alias",
				Email:    "gopher-alias@example.com",
			},
			ctx:     context.Background(),
			wantErr: nil,
		},
		{
			name: "invalid type mapping with adapter",
			obj: &sources.AnotherUser{
				Username: "gopher-alias",
				Email:    "invalid-email", // This should fail MockUser's email validation
			},
			ctx:     context.Background(),
			wantErr: NewValidationError("sources.MockUser", "Email", `self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`),
		},
		{
			name: "valid embedded struct",
			obj: &sources.EmbeddedUser{
				Base: sources.Base{ID: "ab"},
				Name: "Gopher",
			},
			ctx:     context.Background(),
			wantErr: nil,
		},
		{
			name: "invalid embedded struct field",
			obj: &sources.EmbeddedUser{
				Base: sources.Base{ID: "a"}, // Fails size check
				Name: "Gopher",
			},
			ctx:     context.Background(),
			wantErr: NewValidationError("sources.EmbeddedUser", "ID", `self != "" && self.size() > 1`),
		},
		{
			name: "invalid own struct field with embedded",
			obj: &sources.EmbeddedUser{
				Base: sources.Base{ID: "ab"},
				Name: "", // Fails nonzero check
			},
			ctx:     context.Background(),
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
			ctx:     context.Background(),
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
			ctx:     context.Background(),
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
			ctx:     context.Background(),
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
			ctx:          context.Background(),
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
			ctx:          context.Background(),
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
			ctx:          context.Background(),
			isMultiError: true, // Expecting two distinct validation errors
		},
		{
			name: "valid generic struct with string",
			obj: &sources.Box[string]{
				Value: "hello",
			},
			ctx:     context.Background(),
			wantErr: nil,
		},
		{
			name: "invalid generic struct with nil pointer",
			obj: &sources.Box[*string]{
				Value: nil,
			},
			ctx:          context.Background(),
			wantErr:      NewValidationError("sources.Box[T]", "Value", `self != null`),
			isMultiError: true, // Expect both type and field errors
		},
		{
			name: "valid generic struct with struct pointer",
			obj: &sources.Box[*sources.Item]{
				Value: &sources.Item{Name: "valid-item"},
			},
			ctx:     context.Background(),
			wantErr: nil,
		},
		{
			name: "invalid generic struct with invalid nested struct",
			obj: &sources.Box[*sources.Item]{
				Value: &sources.Item{Name: ""}, // name is required
			},
			ctx:          context.Background(),
			wantErr:      NewValidationError("sources.Item", "Name", `self != ""`),
			isMultiError: true,
		},
		{
			name: "valid generic struct with int pointer",
			obj: &sources.Box[*int]{
				Value: intPtr(123),
			},
			ctx:     context.Background(),
			wantErr: nil,
		},
		{
			name: "invalid generic struct with nil int pointer",
			obj: &sources.Box[*int]{
				Value: nil,
			},
			ctx:          context.Background(),
			wantErr:      NewValidationError("sources.Box[T]", "Value", `self != null`),
			isMultiError: true, // Expect both type and field errors
		},
		{
			name: "context cancelled",
			obj: &sources.MockUser{
				Name:  "Gopher",
				Email: "gopher@golang.org",
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
		{
			name: "context timeout",
			obj: &sources.MockUser{
				Name:  "Gopher",
				Email: "gopher@golang.org",
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(2 * time.Nanosecond) // Ensure timeout
				cancel()
				return ctx
			}(),
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "valid simple password",
			obj: &sources.Password{
				Value: "password123",
			},
			ctx:     context.Background(),
			wantErr: nil,
		},
		{
			name: "invalid simple password",
			obj: &sources.Password{
				Value: "weak",
			},
			ctx:     context.Background(),
			wantErr: NewValidationError("sources.Password", "Value", `self.matches('^[a-zA-Z0-9]{8,}$')`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Validate(tt.ctx, tt.obj)

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
					idRuleError := NewValidationError("sources.MockUser", "ID", `self != null`).Error()
					if !strings.Contains(errStr, nameRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", nameRuleError, errStr)
					}
					if !strings.Contains(errStr, emailRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", emailRuleError, errStr)
					}
					if !strings.Contains(errStr, idRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", idRuleError, errStr)
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
				case "invalid generic struct with nil pointer", "invalid generic struct with nil int pointer":
					typeRuleError := NewValidationError("sources.Box[T]", "", "self.Value != null").Error()
					fieldRuleError := NewValidationError("sources.Box[T]", "Value", "self != null").Error()
					if !strings.Contains(errStr, typeRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", typeRuleError, errStr)
					}
					if !strings.Contains(errStr, fieldRuleError) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", fieldRuleError, errStr)
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

func TestValidator_WithGlobalRegistry(t *testing.T) {
	// A simple logger for testing.
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create a new engine.
	engine, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Define a sample struct and its adapter.
	type User struct {
		Name  string
		Email string
	}
	userAdapter := func(obj any) (map[string]any, error) {
		u := obj.(User)
		return map[string]any{
			"Name":  u.Name,
			"Email": u.Email,
		}, nil
	}
	adapters := map[reflect.Type]TypeAdapterTarget{
		reflect.TypeOf(User{}): {TargetName: "veritas.User", Adapter: userAdapter},
	}

	// Register a rule set in the global registry.
	ruleSet := ValidationRuleSet{
		FieldRules: map[string][]string{
			"Name": {`self != ""`},
		},
	}
	Register("veritas.User", ruleSet)
	defer UnregisterAll() // Clean up the registry after the test.

	// Create a new validator using the global registry (provider is nil).
	validator, err := NewValidator(
		WithEngine(engine),
		WithLogger(logger),
		WithTypeAdapters(adapters),
	)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Test case 1: Valid user
	validUser := User{Name: "John Doe", Email: "john.doe@example.com"}
	if err := validator.Validate(context.Background(), validUser); err != nil {
		t.Errorf("Validate() with valid user failed: %v", err)
	}

	// Test case 2: Invalid user
	invalidUser := User{Name: "", Email: "jane.doe@example.com"}
	err = validator.Validate(context.Background(), invalidUser)
	if err == nil {
		t.Errorf("Validate() with invalid user should have failed, but got nil")
	}
}

// setupBenchmark creates a standard validator setup for benchmarking.
func setupBenchmark(b *testing.B) (*Validator, *sources.MockUser, *sources.MockUser) {
	b.Helper()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})) // Use Warn to reduce noise
	engine, err := NewEngine(logger, DefaultFunctions()...)
	if err != nil {
		b.Fatalf("NewEngine() failed: %v", err)
	}

	provider := NewJSONRuleProvider("testdata/rules/user.json")

	adapters := map[reflect.Type]TypeAdapterTarget{
		reflect.TypeOf(&sources.MockUser{}): {TargetName: "sources.MockUser", Adapter: mockUserAdapter},
	}

	validator, err := NewValidator(
		WithEngine(engine),
		WithRuleProvider(provider),
		WithLogger(logger),
		WithTypeAdapters(adapters),
	)
	if err != nil {
		b.Fatalf("NewValidator() failed: %v", err)
	}

	intPtr := func(i int) *int { return &i }
	validUser := &sources.MockUser{
		Name:  "Gopher",
		Email: "gopher@golang.org",
		Age:   20,
		ID:    intPtr(1),
	}
	invalidUser := &sources.MockUser{
		Name:  "",
		Email: "invalid",
		Age:   15,
		ID:    nil,
	}

	return validator, validUser, invalidUser
}

func BenchmarkValidator_Validate_Valid_NoCache(b *testing.B) {
	validator, validUser, _ := setupBenchmark(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Invalidate cache on each run
		validator.engine.programCache.Purge()
		_ = validator.Validate(ctx, validUser)
	}
}

func BenchmarkValidator_Validate_Valid_WithCache(b *testing.B) {
	validator, validUser, _ := setupBenchmark(b)
	ctx := context.Background()

	// Prime the cache
	_ = validator.Validate(ctx, validUser)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(ctx, validUser)
	}
}

func BenchmarkValidator_Validate_Invalid_NoCache(b *testing.B) {
	validator, _, invalidUser := setupBenchmark(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Invalidate cache on each run
		validator.engine.programCache.Purge()
		_ = validator.Validate(ctx, invalidUser)
	}
}

func BenchmarkValidator_Validate_Invalid_WithCache(b *testing.B) {
	validator, _, invalidUser := setupBenchmark(b)
	ctx := context.Background()

	// Prime the cache
	_ = validator.Validate(ctx, invalidUser)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(ctx, invalidUser)
	}
}
