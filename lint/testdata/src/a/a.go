package a

type User struct { // want "invalid type rule for User: ERROR: <input>:1:4: Syntax error: mismatched input '<EOF>' expecting .*" "field NonExistentField in rules for User does not exist in struct"
	Name  string
	Email string
}
