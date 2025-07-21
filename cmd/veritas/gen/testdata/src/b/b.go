package b

//go:generate go run ../../../../../ main.go -pkg b -o b.gen.go .

import (
	"fmt"

	"github.com/podhmo/veritas"
)

type Person struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"nonzero"`
}

func main() {
	v, err := veritas.NewValidator(
		veritas.WithTypes(GetKnownTypes()...),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
