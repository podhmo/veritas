// This file is used as a test source for the static analysis tool.
package user

// User represents a user in the system.
// @cel: self.Password == self.PasswordConfirm
type User struct {
	// User's full name, required and must be less than 50 characters.
	Name string `validate:"required,cel:self.size() < 50"`

	// User's email address, must be a valid email format.
	Email string `validate:"required,email"`

	// User's age, must be 18 or older.
	Age int `validate:"cel:self >= 18"`

	// User's password, required and must be at least 10 characters long.
	Password string `validate:"required,cel:self.size() >= 10"`

	// Password confirmation, must match the Password field.
	PasswordConfirm string `validate:"required"`
}
