package ledger

import (
	participleLexer "github.com/alecthomas/participle/v2/lexer"
)

func pruneMetadataFromAst(ast *Journal) {
	for _, entry := range ast.Entries {
		switch entry := entry.(type) {
		case *AccountDirective:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *RealPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *VirtualPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *VirtualBalancedPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		}
	}
}
