package ledger

import (
	"testing"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"

	"github.com/yeldirium/hledger-language-server/internal/lexing"
	lextesting "github.com/yeldirium/hledger-language-server/internal/lexing/testing"
)

func TestJournalLexer(t *testing.T) {
	t.Run("Miscellaneous", func(t *testing.T) {
		t.Run("Succeeds on an empty input.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput(""),
				lextesting.ExpectTokens([]participleLexer.Token{}),
			)
		})

		t.Run("Lexes newlines", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("\n\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("Lexes garbage and newlines.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("this is not a valid journal file\n\nheckmeck\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Garbage", Value: "this is not a valid journal file"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Garbage", Value: "heckmeck"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("Lexes newlines with whitespace.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    \n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})
	})

	t.Run("Helpers", func(t *testing.T) {
		t.Run("AcceptInlineCommentIndicator", func(t *testing.T) {
			t.Run("accepts an inline comment indicator using a semicolon and emits a token.", func(t *testing.T) {
				lexer, fileName := lexing.PrepareLexer(
					"  ; foo",
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					1,
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: fileName, Line: 1, Column: 4, Offset: 3}, lexer.Pos())
			})

			t.Run("accepts an inline comment indicator using a hash.", func(t *testing.T) {
				lexer, fileName := lexing.PrepareLexer(
					"  # foo",
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					1,
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: fileName, Line: 1, Column: 4, Offset: 3}, lexer.Pos())
			})

			t.Run("does not accept things not starting with spaces.", func(t *testing.T) {
				lexer, fileName := lexing.PrepareLexer(
					"foo",
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					1,
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.False(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: fileName, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("does not accept inline comments starting with anything else.", func(t *testing.T) {
				lexer, fileName := lexing.PrepareLexer(
					"  foo",
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					1,
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.False(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: fileName, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns an error upon encountering EOF.", func(t *testing.T) {
				lexer, _ := lexing.PrepareLexer(
					" ",
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					1,
				)

				_, _, err := AcceptInlineCommentIndicator(lexer)
				assert.ErrorIs(t, err, lexing.ErrEof)
			})
		})
	})

	t.Run("Include directive", func(t *testing.T) {
		t.Run("Lexes a file containing an include directive with a path.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("include some/path.journal\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "IncludeDirective", Value: "include"},
					{Type: "Whitespace", Value: " "},
					{Type: "IncludePath", Value: "some/path.journal"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("Lexes a file containing an include directive without a path.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("include\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "IncludeDirective", Value: "include"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Lexes a file containing an account directive with multiple segments", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("account assets:Cash:Checking\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Checking"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("Lexes a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("account assets:Cash:Che cking:Spe-ci_al\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Che cking"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Spe-ci_al"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("Stops lexing the account directive when it encounters an inline comment.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("account assets:Cash:Checking  ; inline comment\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Checking"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("Lexes a file containing multiple account directives", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput(`account assets:Cash:Checking
account expenses:Food:Groceries
`),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Checking"},
					{Type: "Newline", Value: "\n"},
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Food"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})
	})

	t.Run("Posting", func(t *testing.T) {
		t.Run("lexes a posting and its amount.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    expenses:Groceries      1,234.56 €  ; inline comment\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "Whitespace", Value: "      "},
					{Type: "Amount", Value: "1,234.56 €"},
					{Type: "InlineCommentIndicator", Value: "  ;"},
					{Type: "Garbage", Value: " inline comment"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("lexes a posting with ! status indicator.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    ! expenses:Groceries\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "PostingStatusIndicator", Value: "!"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("lexes a posting with * status indicator.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    * expenses:Groceries\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "PostingStatusIndicator", Value: "*"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("lexes a posting with a virtual account.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    (expenses:Groceries)      1,234.56 €\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "AccountNameDelimiter", Value: "("},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "AccountNameDelimiter", Value: ")"},
					{Type: "Whitespace", Value: "      "},
					{Type: "Amount", Value: "1,234.56 €"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("lexes a posting with a balanced virtual account.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    [expenses:Groceries]      1,234.56 €\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "AccountNameDelimiter", Value: "["},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "AccountNameDelimiter", Value: "]"},
					{Type: "Whitespace", Value: "      "},
					{Type: "Amount", Value: "1,234.56 €"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("lexes a posting with a status indicator and virtual account.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    ! (expenses:Groceries)      1,234.56 €\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "PostingStatusIndicator", Value: "!"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameDelimiter", Value: "("},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "AccountNameDelimiter", Value: ")"},
					{Type: "Whitespace", Value: "      "},
					{Type: "Amount", Value: "1,234.56 €"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("lexes a posting with a currency conversion that includes a double space.", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput("    expenses:Groceries   -1,234.56 BTC @  1,234.50 €\n"),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Indent", Value: "    "},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "Whitespace", Value: "   "},
					{Type: "Amount", Value: "-1,234.56 BTC @  1,234.50 €"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})

		t.Run("fails on invalid inputs.", func(t *testing.T) {
			invalidInputs := []string{
				"    !expenses:Groceries\n",
				"    *expenses:Groceries\n",
				"    )expenses:Groceries\n",
				"    ]expenses:Groceries\n",
			}

			for _, invalidInput := range invalidInputs {
				lextesting.AssertLexerFails(
					t,
					NewJournalLexer(),
					invalidInput,
				)
			}
		})

		t.Run("does not fail on valid inputs.", func(t *testing.T) {
			validInputs := []string{
				"		! (expenses:Groceries)      1,234.56 €\n",
				"  [foo]\n",
				"         expenses\n",
				"    ass:Check  1234  ; fgnqle\n",
				"  assets:Checking    = $123456\n",
			}

			for _, validInput := range validInputs {
				lextesting.AssertLexer(
					t,
					NewJournalLexer(),
					lextesting.IncludeUnexpectedSymbols(),
					lextesting.LexerInput(validInput),
				)
			}
		})
	})

	t.Run("Mixed", func(t *testing.T) {
		t.Run("Lexes a journal file containing many different directives, postings and comments", func(t *testing.T) {
			lextesting.AssertLexer(
				t,
				NewJournalLexer(),
				lextesting.LexerInput(`; This is a cool journal file
# It includes many things
; This next line is empty, but contains whitespace:
    

include someLong/Pathof/things.journal
include   also/there-are-no/inlinecomments/after.includes  ; this is part of the path

account assets:Cash:Checking
    ; indented comment
account expenses:Gro ce:ries  ; hehe

payee Some Cool Person

commodity EUR
    format 1,000.00 €

2024-11-25 ! (code) Payee | transaction reason  ; inline comment
    expenses:Groceries      1,234.56 €
    assets:Cash:Checking   -1,234.56 €  ; inline comment
`),
				lextesting.ExpectMiniTokens([]lextesting.MiniToken{
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Indent", Value: "    "},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "IncludeDirective", Value: "include"},
					{Type: "Whitespace", Value: " "},
					{Type: "IncludePath", Value: "someLong/Pathof/things.journal"},
					{Type: "Newline", Value: "\n"},
					{Type: "IncludeDirective", Value: "include"},
					{Type: "Whitespace", Value: "   "},
					{Type: "IncludePath", Value: "also/there-are-no/inlinecomments/after.includes  ; this is part of the path"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Checking"},
					{Type: "Newline", Value: "\n"},
					{Type: "Indent", Value: "    "},
					{Type: "Newline", Value: "\n"},
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Gro ce"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "ries"},
					{Type: "InlineCommentIndicator", Value: "  ;"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Indent", Value: "    "},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Newline", Value: "\n"},
					{Type: "Indent", Value: "    "},
					{Type: "AccountNameSegment", Value: "expenses"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Groceries"},
					{Type: "Whitespace", Value: "      "},
					{Type: "Amount", Value: "1,234.56 €"},
					{Type: "Newline", Value: "\n"},
					{Type: "Indent", Value: "    "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Checking"},
					{Type: "Whitespace", Value: "   "},
					{Type: "Amount", Value: "-1,234.56 €"},
					{Type: "InlineCommentIndicator", Value: "  ;"},
					{Type: "Newline", Value: "\n"},
				}),
			)
		})
	})
}
