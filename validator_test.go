package veritas

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/podhmo/veritas/testdata/sources"
)

func TestValidator_Validate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	engine, err := NewEngine(logger, DefaultFunctions()...)
	if err != nil {
		t.Fatalf("NewEngine() failed: %v", err)
	}

	provider := NewJSONRuleProvider("testdata/rules/user.json")

	// Create a new validator with the native types.
	validator, err := NewValidator(
		WithEngine(engine),
		WithRuleProvider(provider),
		WithLogger(logger),
		WithTypes(
			sources.Password{},
			sources.MockUser{},
			sources.EmbeddedUser{},
			sources.ComplexUser{},
			sources.Profile{},
			sources.UserWithProfiles{},
			sources.Item{},
			sources.Box[string]{},
			sources.Box[*string]{},
			sources.Box[*int]{},
			sources.Box[*sources.Item]{},
		),
	)
	if err != nil {
		t.Fatalf("NewValidator() failed: %v", err)
	}

	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name          string
		obj           any
		ctx           context.Context
		wantErr       error
		wantMultiError []string // For checking multiple specific errors
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
			wantErr: NewValidationError("sources.MockUser", "Email", `self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`),
		},
		{
			name: "object with multiple errors",
			obj: &sources.MockUser{
				Name:  "", // Fails nonzero
				Email: "invalid-email",
				Age:   20,
				ID:    nil, // Fails required
			},
			ctx: context.Background(),
			wantMultiError: []string{
				NewValidationError("sources.MockUser", "Name", `self != ""`).Error(),
				NewValidationError("sources.MockUser", "Email", `self != "" && self.matches('^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$')`).Error(),
				NewValidationError("sources.MockUser", "ID", "self != null").Error(),
			},
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
			wantErr: NewValidationError("sources.MockUser", "", "self.Age >= 18"),
		},
		{
			name:    "unregistered type",
			obj:     struct{ Age int }{10},
			ctx:     context.Background(),
			wantErr: nil, // No rules for this type, so no error.
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
			ctx: context.Background(),
			wantMultiError: []string{
				NewValidationError("sources.Profile", "Handle", `self != "" && self.size() > 2`).Error(),
			},
		},
		{
			name: "invalid struct in map",
			obj: &sources.UserWithProfiles{
				Name: "Gopher",
				Contacts: map[string]sources.Profile{
					"personal": {Platform: "", Handle: "myhandle"}, // Invalid platform
				},
			},
			ctx: context.Background(),
			wantMultiError: []string{
				NewValidationError("sources.Profile", "Platform", `self != ""`).Error(),
			},
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
			ctx: context.Background(),
			wantMultiError: []string{
				NewValidationError("sources.Profile", "Platform", `self != ""`).Error(),
				NewValidationError("sources.Profile", "Handle", `self != "" && self.size() > 2`).Error(),
			},
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
			ctx: context.Background(),
			wantMultiError: []string{
				NewValidationError("sources.Box[T]", "Value", "self != null").Error(),
			},
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
			ctx: context.Background(),
			wantMultiError: []string{
				NewValidationError("sources.Item", "Name", `self != ""`).Error(),
			},
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
			ctx: context.Background(),
			wantMultiError: []string{
				NewValidationError("sources.Box[T]", "Value", "self != null").Error(),
			},
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

			if len(tt.wantMultiError) > 0 {
				if gotErr == nil {
					t.Fatalf("Validate() expected errors, got nil")
				}
				errStr := gotErr.Error()
				for _, wantErrStr := range tt.wantMultiError {
					if !strings.Contains(errStr, wantErrStr) {
						t.Errorf("Validate() error missing expected content '%s' in '%s'", wantErrStr, errStr)
					}
				}
				return
			}

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

			if !strings.Contains(gotErr.Error(), tt.wantErr.Error()) {
				t.Errorf("Validate() error mismatch\nwant (to contain): %s\ngot:               %s", tt.wantErr.Error(), gotErr.Error())
			}
		})
	}
}

func TestValidator_WithGlobalRegistry(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	engine, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	type User struct {
		Name  string
		Email string
	}

	ruleKey := "veritas.User"
	ruleSet := ValidationRuleSet{
		FieldRules: map[string][]string{
			"Name": {`self != ""`},
		},
	}
	Register(ruleKey, ruleSet)
	defer UnregisterAll()

	validator, err := NewValidator(
		WithEngine(engine),
		WithLogger(logger),
		WithTypes(User{}),
	)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	validUser := User{Name: "John Doe", Email: "john.doe@example.com"}
	if err := validator.Validate(context.Background(), validUser); err != nil {
		t.Errorf("Validate() with valid user failed: %v", err)
	}

	invalidUser := User{Name: "", Email: "jane.doe@example.com"}
	err = validator.Validate(context.Background(), invalidUser)
	if err == nil {
		t.Errorf("Validate() with invalid user should have failed, but got nil")
	} else if !strings.Contains(err.Error(), NewValidationError(ruleKey, "Name", `self != ""`).Error()) {
		t.Errorf("Validate() error mismatch, got %v", err)
	}
}

func setupBenchmark(b *testing.B) (*Validator, *sources.MockUser, *sources.MockUser) {
	b.Helper()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	engine, err := NewEngine(logger, DefaultFunctions()...)
	if err != nil {
		b.Fatalf("NewEngine() failed: %v", err)
	}

	provider := NewJSONRuleProvider("testdata/rules/user.json")

	validator, err := NewValidator(
		WithEngine(engine),
		WithRuleProvider(provider),
		WithLogger(logger),
		WithTypes(sources.MockUser{}),
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
		validator.engine.programCache.Purge()
		_ = validator.Validate(ctx, validUser)
	}
}

func BenchmarkValidator_Validate_Valid_WithCache(b *testing.B) {
	validator, validUser, _ := setupBenchmark(b)
	ctx := context.Background()

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
		validator.engine.programCache.Purge()
		_ = validator.Validate(ctx, invalidUser)
	}
}

func BenchmarkValidator_Validate_Invalid_WithCache(b *testing.B) {
	validator, _, invalidUser := setupBenchmark(b)
	ctx := context.Background()

	_ = validator.Validate(ctx, invalidUser)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(ctx, invalidUser)
	}
}
