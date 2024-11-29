package ledger_test

import (
	participleLexer "github.com/alecthomas/participle/v2/lexer"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func pruneMetadataFromAst(ast *ledger.Journal) {
	for _, entry := range ast.Entries {
		switch entry := entry.(type) {
		case *ledger.AccountDirective:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *ledger.RealPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *ledger.VirtualPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *ledger.VirtualBalancedPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		}
	}
}
