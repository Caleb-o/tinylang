package main

import (
	"fmt"
	"tiny/parser"
	"tiny/shared"
)

func main() {
	parser := parser.New(shared.ReadFile("./tests/valid/parser/function_definition_nested.tiny"))
	program := parser.Parse()

	fmt.Println(program.Body.AsSExp())
}
