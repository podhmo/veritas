package sources

import "net/url"

// @cel: self.Age >= 18
// MockUser is a test struct.
type MockUser struct {
	Name  string `validate:"nonzero"`
	Email string `validate:"nonzero,email"`
	Age   int
	ID    *int `validate:"required"` // Pointer to test nil check for required
	URL   *url.URL
}

// MockVariety is a struct with various field types for testing.
type MockVariety struct {
	Count    int      `validate:"nonzero"`
	IsActive bool     `validate:"nonzero"`
	Scores   []int    `validate:"nonzero"`
	Metadata map[string]string `validate:"nonzero"`
	Optional *string
}
