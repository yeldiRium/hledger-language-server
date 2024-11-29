package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func main() {
	parser := ledger.NewJournalParser()
	tokens, err := parser.Lex("stdin", os.Stdin)
	if err != nil {
		fmt.Printf("An error occured: %v\n\n", err)
		fmt.Printf("Maybe the parser still found something? Here are the tokens:\n\n")
	}

    output, err := json.Marshal(tokens)
    if err != nil {
        fmt.Printf("Failed to convert AST to json. Here is the raw AST:\n\n")
        fmt.Printf("%#v\n", tokens)
    }
	fmt.Printf("%s\n", output)
}
