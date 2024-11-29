package main

import (
	"fmt"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func main() {
	parser := ledger.NewJournalParser()
	fmt.Println(parser.String())
}
