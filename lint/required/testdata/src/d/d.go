package d

type Input struct {
	Name string `validate:"required"` // want "'required' tag can only be used with pointer types"
}
