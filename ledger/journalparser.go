package ledger

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	participleLexer "github.com/alecthomas/participle/v2/lexer"
)

type Journal struct {
	Entries []Entry `parser:"(@@ | Newline | Garbage)*"`
}

type Entry interface {
	value()
}

type IncludeDirective struct {
	IncludePath string `parser:"'include' Whitespace @IncludePath Newline"`
}

func (*IncludeDirective) value() {}

type AccountDirective struct {
	AccountName *AccountName `parser:"'account' Whitespace @@ (InlineCommentIndicator Garbage)? Newline"`
}

func (*AccountDirective) value() {}

type AccountName struct {
	Pos    participleLexer.Position
	EndPos participleLexer.Position

	Segments []string `parser:"@AccountNameSegment (':' @AccountNameSegment)*"`
}

func (accountName *AccountName) String() string {
	return strings.Join(accountName.Segments, ":")
}

type RealPosting struct {
	PostingStatus string       `parser:"Indent (@('*' | '!') ' ')?"`
	AccountName   *AccountName `parser:"@@"`
	Amount        string       `parser:"(Whitespace @Amount)? (InlineCommentIndicator Garbage)? Newline"`
}

func (*RealPosting) value() {}

type VirtualPosting struct {
	PostingStatus string       `parser:"Indent (@('*' | '!') ' ')?"`
	AccountName   *AccountName `parser:"'(' @@ ')'"`
	Amount        string       `parser:"(Whitespace @Amount)? (InlineCommentIndicator Garbage)? Newline"`
}

func (*VirtualPosting) value() {}

type VirtualBalancedPosting struct {
	PostingStatus string       `parser:"Indent (@('*' | '!') ' ')?"`
	AccountName   *AccountName `parser:"'[' @@ ']'"`
	Amount        string       `parser:"(Whitespace @Amount)? (InlineCommentIndicator Garbage)? Newline"`
}

func (*VirtualBalancedPosting) value() {}

type JournalParser = participle.Parser[Journal]

func NewJournalParser() *JournalParser {
	lexer := NewJournalLexer()
	parser, err := participle.Build[Journal](
		participle.Lexer(lexer),
		participle.UseLookahead(3),
		participle.Union[Entry](&IncludeDirective{}, &AccountDirective{}, &RealPosting{}, &VirtualPosting{}, &VirtualBalancedPosting{}),
	)
	if err != nil {
		panic(err)
	}
	return parser
}
