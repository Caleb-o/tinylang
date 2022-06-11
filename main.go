package main

import (
	"fmt"
	"tiny/parser"
)

func main() {
	parser := parser.New("1 + 2 * 3 + (1 * 2)")
	program := parser.Parse()

	fmt.Println(program.Body.AsSExp())
}
