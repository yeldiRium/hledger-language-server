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

type Posting struct {
	AccountName *AccountName `parser:"Indent @@"`
	Amount      string       `parser:"(Whitespace @Amount)? Garbage? Newline"`
}

func (*Posting) value() {}

func MakeJournalParser() *participle.Parser[Journal] {
	lexer := MakeJournalLexer()
	parser, err := participle.Build[Journal](
		participle.Lexer(lexer),
		participle.UseLookahead(3),
		participle.Union[Entry](&AccountDirective{}, &Posting{}),
	)
	if err != nil {
		panic(err)
	}
	return parser
}
