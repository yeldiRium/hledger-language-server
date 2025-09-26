package main

import (
	"fmt"

	"github.com/yeldirium/hledger-language-server/internal/ledger"
)

func main() {
	parser := ledger.NewJournalParser()
	fmt.Println(parser.String())
}
