package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func MakeLexer() *lexer.StatefulDefinition {
	return lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Word", Pattern: `[^:\s]+`},
		{Name: "Separator", Pattern: `:`},
		{Name: "Whitespace", Pattern: `[ ]+`},
	})
}

type AccountDirective struct {
	AccountName *AccountName `parser:"'account' ' ' @@"`
}
type AccountName struct {
	Segments []*AccountNameSegment `parser:"@@ (Separator @@)*"`
}
type AccountNameSegment struct {
	String string `parser:"@(Word (' ' Word)*)"`
}

func MakeParser() *participle.Parser[AccountDirective] {
	lexer := MakeLexer()
	parser, err := participle.Build[AccountDirective](participle.Lexer(lexer), participle.UseLookahead(3))
	if err != nil {
		panic(err)
	}
	return parser
}
