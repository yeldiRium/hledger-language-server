package ledger

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	participleLexer "github.com/alecthomas/participle/v2/lexer"
)

type Journal struct {
	Entries []Entry `parser:"(@@ | Indent? (Newline | Garbage))*"`
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

func (accountName *AccountName) Prefixes() []AccountName {
	prefixes := make([]AccountName, len(accountName.Segments))

	for i := range accountName.Segments {
		prefixes[i] = AccountName{Segments: make([]string, i+1)}
		for j := 0; j <= i; j++ {
			prefixes[i].Segments[j] = accountName.Segments[j]
		}
	}

	return prefixes
}

func (accountName AccountName) Equals(otherAccountName AccountName) bool {
	if len(accountName.Segments) != len(otherAccountName.Segments) {
		return false
	}
	for i, segment := range accountName.Segments {
		if otherAccountName.Segments[i] != segment {
			return false
		}
	}
	return true
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
		participle.Elide("Whitespace"),
	)
	if err != nil {
		panic(err)
	}
	return parser
}
