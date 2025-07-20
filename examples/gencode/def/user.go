package def

//go:generate go run ../../../cmd/veritas -o ../validation/validator.go .

// veritas:
type User struct {
	Name  string `json:"name" validate:"cel:self.size() > 0"`
	Email string `json:"email" validate:"cel:custom.matches(self, '^[^@]+@[^@]+$')"`
}
