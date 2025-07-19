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
	Count    int               `validate:"nonzero"`
	IsActive bool              `validate:"nonzero"`
	Scores   []int             `validate:"nonzero"`
	Metadata map[string]string `validate:"nonzero"`
	Optional *string
}

// MockComplexData is a struct for testing advanced validation scenarios
// like slices and maps.
type MockComplexData struct {
	// Validate each string in the slice is a valid email.
	UserEmails []string `validate:"dive,email"`

	// Validate each key in the map starts with "id_" and each value is not nil.
	ResourceMap map[string]*int `validate:"keys,cel:self.startsWith('id_'),values,required"`

	// Validate a slice of pointers to MockUser, ensuring each pointer is not nil
	// and then diving into the struct's own validation.
	Users []*MockUser `validate:"dive,required"`

	// Nested dive: slice of slices of ints. Each int must not be zero.
	Matrix [][]int `validate:"dive,dive,nonzero"`
}

// MockMoreComplexData is a struct for testing more advanced nested validation.
type MockMoreComplexData struct {
	// A slice of maps. Validate that each map is not nil, each key is a valid email,
	// and each value is not an empty string.
	ListOfMaps []map[string]string `validate:"dive,nonzero,keys,email,values,nonzero"`

	// A map with slice values. Validate that each key is not empty,
	// and that each string within the nested slice is not empty.
	MapOfSlices map[string][]string `validate:"keys,nonzero,values,dive,nonzero"`
}

type Base struct {
	ID string `validate:"nonzero,cel:self.size() > 1"`
}

type EmbeddedUser struct {
	Base
	Name string `validate:"nonzero"`
}

type ComplexUser struct {
	Name     string `validate:"nonzero"`
	Scores   []int  `validate:"cel:self.all(x, x >= 0)"`
	Metadata map[string]string
}

type Profile struct {
	Platform string `validate:"nonzero"`
	Handle   string `validate:"nonzero,cel:self.size() > 2"`
}

type UserWithProfiles struct {
	Name     string             `validate:"nonzero"`
	Profiles []Profile          // Slice of structs
	Contacts map[string]Profile // Map of structs
}

// Password contains a password string to test complex regex.
type Password struct {
	Value string
}
