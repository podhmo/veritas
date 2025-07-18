package c

type Order struct { // want "invalid field rule for Order.Amount: ERROR: <input>:1:7: Syntax error: mismatched input '<EOF>' expecting .*"
	Amount int
}
