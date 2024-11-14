package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func MakeJournalLexer() *lexer.StatefulDefinition {
	return lexer.MustStateful(lexer.Rules{
		"Root": {
			{Name: "Newline", Pattern: `\n`, Action: nil},
			{Name: "CommentIndicator", Pattern: `^\s*[;#]`, Action: nil},
			{Name: "InlineCommentIndicator", Pattern: `  [;#] `, Action: nil},
			{Name: "Whitespace", Pattern: ` `, Action: nil},
			{Name: "AccountDirective", Pattern: `account`, Action: nil},
			{Name: "PayeeDirective", Pattern: `payee`, Action: nil},
			{Name: "Word", Pattern: `[^\s;#]+`, Action: nil},
		},
	})
}

type Journal struct {
	Entries []Entry `parser:"(@@ | Newline)*"`
}

type Entry interface {
	value()
}

type AccountDirective struct {
	AccountName string         `parser:"AccountDirective ' ' @(Word (' ' Word)*)"`
	Comment     *InlineComment `parser:"(@@)? Newline"`
}

func (*AccountDirective) value() {}

type PayeeDirective struct {
	PayeeName string         `parser:"PayeeDirective ' ' @(Word (' ' Word)*)"`
	Comment   *InlineComment `parser:"(@@)? Newline"`
}

func (*PayeeDirective) value() {}

type InlineComment struct {
	String string `parser:"InlineCommentIndicator @(Word (Whitespace+ Word)*)"`
}

type Comment struct {
	String string `parser:"CommentIndicator Whitespace* @((Whitespace* (Word | CommentIndicator))*) Newline"`
}

func (*Comment) value() {}

func MakeJournalParser() *participle.Parser[Journal] {
	lexer := MakeJournalLexer()
	parser, err := participle.Build[Journal](
		participle.Lexer(lexer),
		participle.UseLookahead(3),
		participle.Union[Entry](&Comment{}, &AccountDirective{}, &PayeeDirective{}),
	)
	if err != nil {
		panic(err)
	}
	return parser
}
