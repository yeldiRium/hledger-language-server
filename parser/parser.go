package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func MakeLexer() *lexer.StatefulDefinition {
	return lexer.MustStateful(lexer.Rules{
		"Root": {
			{Name: "CommentIndicator", Pattern: "[;#] ", Action: lexer.Push("Comment")},
			{Name: "AccountDirective", Pattern: "account", Action: lexer.Push("AccountDirective")},
			{Name: "PayeeDirective", Pattern: "payee", Action: lexer.Push("PayeeDirective")},
			{Name: "newline", Pattern: "\\n", Action: nil},
		},
		"Comment": {
			{Name: "EOL", Pattern: `\n`, Action: lexer.Pop()},
			{Name: "CommentContent", Pattern: `[^\n]*`, Action: nil},
		},
		"InlineCommentIndicator": {
			{Name: "InlineCommentIndicator", Pattern: "  [;#] ", Action: lexer.Push("InlineComment")},
		},
		"InlineComment": {
			{Name: "InlineCommentContent", Pattern: `.*`, Action: lexer.Pop()},
		},
		"AccountDirective": {
			{Name: "EOL", Pattern: "\\n", Action: lexer.Pop()},
			lexer.Include("InlineCommentIndicator"),
			{Name: "Whitespace", Pattern: `[ ]+`, Action: nil},
			{Name: "Separator", Pattern: `:`, Action: nil},
			{Name: "Segment", Pattern: `[^:\s]+`, Action: nil},
		},
		"PayeeDirective": {
			{Name: "EOL", Pattern: "\\n", Action: lexer.Pop()},
			lexer.Include("InlineCommentIndicator"),
			{Name: "Whitespace", Pattern: `[ ]+`, Action: nil},
			{Name: "PayeeName", Pattern: `.*`, Action: nil},
		},
		"Posting": {
		},
	})
}

type Journal struct {
	Entries []Entry `parser:"@@*"`
}

type Entry interface {
	value()
}

type AccountDirective struct {
	AccountName *AccountName `parser:"AccountDirective ' ' @@"`
	Comment *InlineComment `parser:"(@@)? EOL"`
}

func (*AccountDirective) value() {}

type AccountName struct {
	Segments []*AccountNameSegment `parser:"@@ (Separator @@)*"`
}

type AccountNameSegment struct {
	String string `parser:"@(Segment (' ' Segment)*)"`
}

type PayeeDirective struct {
	PayeeName *PayeeName `parser:"PayeeDirective ' ' @@"`
	Comment *InlineComment `parser:"(@@)? EOL"`
}

func (*PayeeDirective) value() {}

type PayeeName struct {
	String string `parser:"@(PayeeName (' ' PayeeName)*)"`
}

type InlineComment struct {
	String string `parser:"InlineCommentIndicator @InlineCommentContent"`
}

type Comment struct {
	String string `parser:"CommentIndicator @CommentContent EOL"`
}

func (*Comment) value() {}

func MakeParser() *participle.Parser[Journal] {
	lexer := MakeLexer()
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
