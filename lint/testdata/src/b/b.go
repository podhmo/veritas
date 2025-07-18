package b

type Product struct { // want "field Price in rules for Product does not exist in struct"
	Name string
}
