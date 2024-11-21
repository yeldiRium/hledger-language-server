package ledger

import (
	"github.com/alecthomas/participle/v2"
)

type Journal struct {
	Entries []Entry `parser:"(@@ | EOL)*"`
}

type Entry interface {
	value()
}

type AccountDirective struct {
	AccountName string   `parser:"AccountDirective ' ' @AccountName"`
}

func (*AccountDirective) value() {}

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
