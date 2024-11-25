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
	AccountName *AccountName   `parser:"AccountDirective ' ' @@"`
}

func (*AccountDirective) value() {}

type AccountName struct {
Segments []string `parser:"@AccountNameSegment (':' @AccountNameSegment)*"`
}

func MakeJournalParser() *participle.Parser[Journal] {
	lexer := MakeJournalLexer()
	parser, err := participle.Build[Journal](
		participle.Lexer(lexer),
		participle.UseLookahead(3),
		participle.Union[Entry](&AccountDirective{}),
	)
	if err != nil {
		panic(err)
	}
	return parser
}
