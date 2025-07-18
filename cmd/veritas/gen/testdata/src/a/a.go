package a

// @cel-type: User
// @cel: self.Email != ""
type User struct {
	Name string `validate:"required"`
	Email string `validate:"email"`
}
