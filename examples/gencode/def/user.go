package def

// veritas:
type User struct {
	Name  string `json:"name" validate:"cel:self.size() > 0"`
	Email string `json:"email" validate:"cel:custom.matches(self, '^[^@]+@[^@]+$')"`
}
