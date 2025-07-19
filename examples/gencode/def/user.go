package def

// veritas:
type User struct {
	Name  string `json:"name" validate:"cel:len(self) > 0"`
	Email string `json:"email" validate:"cel:custom.matches(self, '^[^@]+@[^@]+$')"`
}
