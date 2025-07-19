package sources

// @cel: self.Value != null
type Box[T any] struct {
	Value T `validate:"required"`
}

type Item struct {
	Name string `validate:"nonzero"`
}
