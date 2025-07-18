package a // want "invalid type rule for User: ERROR: <input>:1:4: Syntax error: mismatched input '<EOF>' expecting .*"

type User struct { // want "field NonExistentField in rules for User does not exist in struct"
	Name  string
	Email string
}
