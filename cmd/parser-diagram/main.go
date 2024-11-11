package main

import (
	"fmt"

	"github.com/yeldiRium/hledger-language-server/parser"
)

func main() {
	parser := parser.MakeParser()
	fmt.Println(parser.String())
}
