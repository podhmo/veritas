package sources

// @cel: self.Age >= 18
// MockUser is a test struct.
type MockUser struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int
}
