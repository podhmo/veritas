package main

type User struct {
	Name string `validate:"nonzero"`
}
