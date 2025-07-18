package a

type User struct { // want "field NonExistentField in rules for User does not exist in struct"
	Name  string
	Email string
}

// want "invalid type rule for User: ERROR: <input>:1:9: Syntax error: mismatched input 'cel' expecting <EOF>"
