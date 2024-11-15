package ledger

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Caveats of this implementation:
// - Postings need to be indented by at least two spaces, since I can't get the parser to work otherwise.
func MakeJournalLexer() *lexer.StatefulDefinition {
	return lexer.MustStateful(lexer.Rules{
		"Root": {
			{Name: "Newline", Pattern: `\n`, Action: nil},
			{Name: "CommentIndicator", Pattern: `^\s*[;#]`, Action: nil},
			{Name: "InlineCommentIndicator", Pattern: `  [;#] `, Action: nil},
			{Name: "PostingIndent", Pattern: `  +`, Action: lexer.Push("Posting")},
			{Name: "Whitespace", Pattern: ` `, Action: nil},
			{Name: "AccountDirective", Pattern: `account`, Action: lexer.Push("AccountName")},
			{Name: "PayeeDirective", Pattern: `payee`, Action: nil},
			{Name: "TransactionDate", Pattern: `(\d{4}[-/.])?[01]?\d[-/.][0123]?\d`, Action: lexer.Push("Transaction")},
			{Name: "Word", Pattern: `[^\s;#]+`, Action: nil},
		},
		"AccountName": {
			{Name: "Newline", Pattern: `\n`, Action: lexer.Pop()},
			{Name: "InlineCommentIndicator", Pattern: `  [;#] `, Action: lexer.Pop()},
			{Name: "Whitespace", Pattern: ` `, Action: nil},
			{Name: "AccountNameSeparator", Pattern: `:`, Action: nil},
			{Name: "AccountNameSegment", Pattern: `[^\s:;#]+( [^\s:;#]+)*`, Action: nil},
		},
		"Transaction": {
			{Name: "Newline", Pattern: `\n`, Action: lexer.Pop()},
			{Name: "InlineCommentIndicator", Pattern: `  [;#] `, Action: lexer.Pop()},
			{Name: "Whitespace", Pattern: ` `, Action: nil},
			{Name: "CodeParentheses", Pattern: `[()]`, Action: nil},
			{Name: "PayeeSeparator", Pattern: `\|`, Action: nil},
			{Name: "TransactionStatusIndicator", Pattern: `[!*]`, Action: nil},
			{Name: "TransactionWord", Pattern: `[^\s;#()|!*]+`, Action: nil},
		},
		"Posting": {
			lexer.Include("AccountName"),
			{Name: "PostingAmount", Pattern: `([^\s]+ )?[+-]?[\d,.]+( \d,.]+)*( [^\s]+)`, Action: nil},
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
	AccountName *AccountName   `parser:"AccountDirective ' ' @@"`
	Comment     *InlineComment `parser:"(@@)? Newline"`
}

func (*AccountDirective) value() {}

type AccountName struct {
	Segments []string `parser:"@AccountNameSegment (AccountNameSeparator @AccountNameSegment)*"`
}

type PayeeDirective struct {
	PayeeName string         `parser:"PayeeDirective ' ' @(Word (' ' Word)*)"`
	Comment   *InlineComment `parser:"(@@)? Newline"`
}

func (*PayeeDirective) value() {}

type Transaction struct {
	Date        string         `parser:"@TransactionDate"`
	Status      string         `parser:"(' ' @TransactionStatusIndicator)?"`
	Code        string         `parser:"(' ' '(' @TransactionWord ')')?"`
	Payee       string         `parser:"(' ' @((TransactionWord|TransactionStatusIndicator|CodeParentheses) (' ' (TransactionWord|TransactionStatusIndicator|CodeParentheses))*) ' ' PayeeSeparator)?"`
	Description string         `parser:"(' ' @((TransactionWord|TransactionStatusIndicator|CodeParentheses|PayeeSeparator) (' ' (TransactionWord|TransactionStatusIndicator|CodeParentheses|PayeeSeparator))*))?"`
	Comment     *InlineComment `parser:"(@@)? Newline"`
	Postings    []*Posting     `parser:"(@@)*"`
}

func (*Transaction) value() {}

type Posting struct {
	AccountName *AccountName   `parser:"PostingIndent @@"`
	Amount      string         `parser:"('  '+ ' '* @PostingAmount)?"`
	Comment     *InlineComment `parser:"(@@)? Newline"`
}

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
		participle.Union[Entry](&Comment{}, &AccountDirective{}, &PayeeDirective{}, &Transaction{}),
	)
	if err != nil {
		panic(err)
	}
	return parser
}
