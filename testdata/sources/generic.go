package sources

// @cel: self.Value != null
type Box[T any] struct {
	Value T `validate:"required"`
}

type Item struct {
	Name string `validate:"nonzero"`
}

// @cel: self.First != null && self.Second != null
type Pair[K any, V any] struct {
	First  K `validate:"required"`
	Second V `validate:"required"`
}
