package a

// @cel-type: User
// @cel: self.Email != ""
type User struct {
	Name  string `validate:"nonzero"`
	Email string `validate:"nonzero,email"`
}
