package ledger

import (
	"github.com/alecthomas/participle/v2"
)

type Journal struct {
	Entries []Entry `parser:"(@@ | Newline | Garbage)*"`
}

type Entry interface {
	value()
}

type AccountDirective struct {
	AccountName *AccountName `parser:"AccountDirective ' ' @@ Garbage? Newline"`
}

func (*AccountDirective) value() {}

type AccountName struct {
	Segments []string `parser:"@AccountNameSegment (':' @AccountNameSegment)*"`
}

type RealPosting struct {
	PostingStatus string		 `parser:"Indent (@('*' | '!') ' ')?"`
	AccountName *AccountName `parser:"@@"`
	Amount      string       `parser:"(Whitespace @Amount)? Garbage? Newline"`
}

func (*RealPosting) value() {}

type VirtualPosting struct {
	PostingStatus string		 `parser:"Indent (@('*' | '!') ' ')?"`
	AccountName *AccountName `parser:"'(' @@ ')'"`
	Amount      string       `parser:"(Whitespace @Amount)? Garbage? Newline"`
}

func (*VirtualPosting) value() {}

type VirtualBalancedPosting struct {
	PostingStatus string		 `parser:"Indent (@('*' | '!') ' ')?"`
	AccountName *AccountName `parser:"'[' @@ ']'"`
	Amount      string       `parser:"(Whitespace @Amount)? Garbage? Newline"`
}

func (*VirtualBalancedPosting) value() {}

func MakeJournalParser() *participle.Parser[Journal] {
	lexer := MakeJournalLexer()
	parser, err := participle.Build[Journal](
		participle.Lexer(lexer),
		participle.UseLookahead(3),
		participle.Union[Entry](&AccountDirective{}, &RealPosting{}, &VirtualPosting{}, &VirtualBalancedPosting{}),
	)
	if err != nil {
		panic(err)
	}
	return parser
}
