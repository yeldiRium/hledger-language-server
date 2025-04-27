package ledger

import (
	"testing"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"

	"github.com/yeldiRium/hledger-language-server/internal/lexing"
	lextesting "github.com/yeldiRium/hledger-language-server/internal/lexing/testing"
)

func TestJournalLexer(t *testing.T) {
	fileName := "testFile"
	runLexer := func(input string) (*lexing.Lexer, []participleLexer.Token, error) {
		lexerDefinition := NewJournalLexer()
		return lextesting.RunLexerWithFileName(lexerDefinition, input, fileName)
	}
	makeTokens := func(lexer *lexing.Lexer, miniTokens []lextesting.MiniToken) []participleLexer.Token {
		return lextesting.MakeTokens(lexer, fileName, miniTokens)
	}

	t.Run("Miscellaneous", func(t *testing.T) {
		t.Run("Succeeds on an empty input.", func(t *testing.T) {
			_, tokens, err := runLexer("")
			assert.NoError(t, err)
			assert.Equal(t, []participleLexer.Token{}, tokens)
		})

		t.Run("Lexes newlines", func(t *testing.T) {
			lexer, tokens, err := runLexer("\n\n")
			assert.NoError(t, err)
			assert.Equal(t, makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
			}), tokens)
		})

		t.Run("Lexes garbage and newlines.", func(t *testing.T) {
			lexer, tokens, err := runLexer("this is not a valid journal file\n\nheckmeck\n")
			assert.NoError(t, err)
			assert.Equal(t, makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Garbage", Value: "this is not a valid journal file"},
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "heckmeck"},
				{Type: "Newline", Value: "\n"},
			}), tokens)
		})

		t.Run("Lexes newlines with whitespace.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    \n")
			assert.NoError(t, err)
			assert.Equal(t, makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Indent", Value: "    "},
				{Type: "Newline", Value: "\n"},
			}), tokens)
		})
	})

	t.Run("Helpers", func(t *testing.T) {
		t.Run("AcceptInlineCommentIndicator", func(t *testing.T) {
			t.Run("accepts an inline comment indicator using a semicolon and emits a token.", func(t *testing.T) {
				filename := fileName
				lexerDefinition := lexing.NewLexerDefinition(
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
				)
				lexer := lexing.NewLexer(
					filename,
					lexerDefinition,
					"  ; foo",
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan participleLexer.Token, 1),
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, lexer.Pos())
			})

			t.Run("accepts an inline comment indicator using a hash.", func(t *testing.T) {
				filename := fileName
				lexerDefinition := lexing.NewLexerDefinition(
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
				)
				lexer := lexing.NewLexer(
					filename,
					lexerDefinition,
					"  # foo",
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan participleLexer.Token, 1),
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, lexer.Pos())
			})

			t.Run("does not accept things not starting with spaces.", func(t *testing.T) {
				filename := fileName
				lexerDefinition := lexing.NewLexerDefinition(
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
				)
				lexer := lexing.NewLexer(
					filename,
					lexerDefinition,
					"foo",
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan participleLexer.Token, 1),
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.False(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("does not accept inline comments starting with anything else.", func(t *testing.T) {
				filename := fileName
				lexerDefinition := lexing.NewLexerDefinition(
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
				)
				lexer := lexing.NewLexer(
					filename,
					lexerDefinition,
					"  foo",
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan participleLexer.Token, 1),
				)

				ok, _, err := AcceptInlineCommentIndicator(lexer)
				assert.NoError(t, err)
				assert.False(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns an error upon encountering EOF.", func(t *testing.T) {
				filename := fileName
				lexerDefinition := lexing.NewLexerDefinition(
					func(lexer *lexing.Lexer) lexing.StateFn { return nil },
					[]string{
						"Char",
						"InlineCommentIndicator",
					},
				)
				lexer := lexing.NewLexer(
					filename,
					lexerDefinition,
					" ",
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan participleLexer.Token, 1),
				)

				_, _, err := AcceptInlineCommentIndicator(lexer)
				assert.ErrorIs(t, err, lexing.ErrEof)
			})
		})
	})

	t.Run("Include directive", func(t *testing.T) {
		t.Run("Lexes a file containing an include directive with a path.", func(t *testing.T) {
			lexer, tokens, err := runLexer("include some/path.journal\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				makeTokens(lexer, []lextesting.MiniToken{
					{Type: "IncludeDirective", Value: "include"},
					{Type: "Whitespace", Value: " "},
					{Type: "IncludePath", Value: "some/path.journal"},
					{Type: "Newline", Value: "\n"},
				}),
				tokens,
			)
		})

		t.Run("Lexes a file containing an include directive without a path.", func(t *testing.T) {
			lexer, tokens, err := runLexer("include\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				makeTokens(lexer, []lextesting.MiniToken{
					{Type: "IncludeDirective", Value: "include"},
					{Type: "Newline", Value: "\n"},
				}),
				tokens,
			)
		})
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Lexes a file containing an account directive with multiple segments", func(t *testing.T) {
			lexer, tokens, err := runLexer("account assets:Cash:Checking\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				makeTokens(lexer, []lextesting.MiniToken{
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Checking"},
					{Type: "Newline", Value: "\n"},
				}),
				tokens,
			)
		})

		t.Run("Lexes a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			lexer, tokens, err := runLexer("account assets:Cash:Che cking:Spe-ci_al\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				makeTokens(lexer, []lextesting.MiniToken{
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
				tokens,
			)
		})

		t.Run("Stops lexing the account directive when it encounters an inline comment.", func(t *testing.T) {
			lexer, tokens, err := runLexer("account assets:Cash:Checking  ; inline comment\n")
			assert.NoError(t, err)
			assert.Greater(t, len(tokens), 7)
			assert.Equal(
				t,
				makeTokens(lexer, []lextesting.MiniToken{
					{Type: "AccountDirective", Value: "account"},
					{Type: "Whitespace", Value: " "},
					{Type: "AccountNameSegment", Value: "assets"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Cash"},
					{Type: "AccountNameSeparator", Value: ":"},
					{Type: "AccountNameSegment", Value: "Checking"},
				}),
				tokens[0:7],
			)
		})

		t.Run("Lexes a file containing multiple account directives", func(t *testing.T) {
			lexer, tokens, err := runLexer(`account assets:Cash:Checking
account expenses:Food:Groceries
`)
			assert.NoError(t, err)
			assert.Equal(
				t,
				[]participleLexer.Token{
					{Type: lexer.Symbol("AccountDirective"), Value: "account", Pos: participleLexer.Position{Filename: fileName, Offset: 0, Line: 1, Column: 1}},
					{Type: lexer.Symbol("Whitespace"), Value: " ", Pos: participleLexer.Position{Filename: fileName, Offset: 7, Line: 1, Column: 8}},
					{Type: lexer.Symbol("AccountNameSegment"), Value: "assets", Pos: participleLexer.Position{Filename: fileName, Offset: 8, Line: 1, Column: 9}},
					{Type: lexer.Symbol("AccountNameSeparator"), Value: ":", Pos: participleLexer.Position{Filename: fileName, Offset: 14, Line: 1, Column: 15}},
					{Type: lexer.Symbol("AccountNameSegment"), Value: "Cash", Pos: participleLexer.Position{Filename: fileName, Offset: 15, Line: 1, Column: 16}},
					{Type: lexer.Symbol("AccountNameSeparator"), Value: ":", Pos: participleLexer.Position{Filename: fileName, Offset: 19, Line: 1, Column: 20}},
					{Type: lexer.Symbol("AccountNameSegment"), Value: "Checking", Pos: participleLexer.Position{Filename: fileName, Offset: 20, Line: 1, Column: 21}},
					{Type: lexer.Symbol("Newline"), Value: "\n", Pos: participleLexer.Position{Filename: fileName, Offset: 28, Line: 1, Column: 29}},
					{Type: lexer.Symbol("AccountDirective"), Value: "account", Pos: participleLexer.Position{Filename: fileName, Offset: 29, Line: 2, Column: 1}},
					{Type: lexer.Symbol("Whitespace"), Value: " ", Pos: participleLexer.Position{Filename: fileName, Offset: 36, Line: 2, Column: 8}},
					{Type: lexer.Symbol("AccountNameSegment"), Value: "expenses", Pos: participleLexer.Position{Filename: fileName, Offset: 37, Line: 2, Column: 9}},
					{Type: lexer.Symbol("AccountNameSeparator"), Value: ":", Pos: participleLexer.Position{Filename: fileName, Offset: 45, Line: 2, Column: 17}},
					{Type: lexer.Symbol("AccountNameSegment"), Value: "Food", Pos: participleLexer.Position{Filename: fileName, Offset: 46, Line: 2, Column: 18}},
					{Type: lexer.Symbol("AccountNameSeparator"), Value: ":", Pos: participleLexer.Position{Filename: fileName, Offset: 50, Line: 2, Column: 22}},
					{Type: lexer.Symbol("AccountNameSegment"), Value: "Groceries", Pos: participleLexer.Position{Filename: fileName, Offset: 51, Line: 2, Column: 23}},
					{Type: lexer.Symbol("Newline"), Value: "\n", Pos: participleLexer.Position{Filename: fileName, Offset: 60, Line: 2, Column: 32}},
				},
				tokens,
			)
		})
	})

	t.Run("Posting", func(t *testing.T) {
		t.Run("lexes a posting and its amount.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    expenses:Groceries      1,234.56 €  ; inline comment\n")

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Indent", Value: "    "},
				{Type: "AccountNameSegment", Value: "expenses"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "Groceries"},
				{Type: "Whitespace", Value: "      "},
				{Type: "Amount", Value: "1,234.56 €"},
				{Type: "InlineCommentIndicator", Value: "  ;"},
				{Type: "Garbage", Value: " inline comment"},
				{Type: "Newline", Value: "\n"},
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, tokens)
		})

		t.Run("lexes a posting with ! status indicator.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    ! expenses:Groceries\n")

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Indent", Value: "    "},
				{Type: "PostingStatusIndicator", Value: "!"},
				{Type: "Whitespace", Value: " "},
				{Type: "AccountNameSegment", Value: "expenses"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "Groceries"},
				{Type: "Newline", Value: "\n"},
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, tokens)
		})

		t.Run("lexes a posting with * status indicator.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    * expenses:Groceries\n")

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Indent", Value: "    "},
				{Type: "PostingStatusIndicator", Value: "*"},
				{Type: "Whitespace", Value: " "},
				{Type: "AccountNameSegment", Value: "expenses"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "Groceries"},
				{Type: "Newline", Value: "\n"},
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, tokens)
		})

		t.Run("lexes a posting with a virtual account.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    (expenses:Groceries)      1,234.56 €\n")

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Indent", Value: "    "},
				{Type: "AccountNameDelimiter", Value: "("},
				{Type: "AccountNameSegment", Value: "expenses"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "Groceries"},
				{Type: "AccountNameDelimiter", Value: ")"},
				{Type: "Whitespace", Value: "      "},
				{Type: "Amount", Value: "1,234.56 €"},
				{Type: "Newline", Value: "\n"},
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, tokens)
		})

		t.Run("lexes a posting with a balanced virtual account.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    [expenses:Groceries]      1,234.56 €\n")

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Indent", Value: "    "},
				{Type: "AccountNameDelimiter", Value: "["},
				{Type: "AccountNameSegment", Value: "expenses"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "Groceries"},
				{Type: "AccountNameDelimiter", Value: "]"},
				{Type: "Whitespace", Value: "      "},
				{Type: "Amount", Value: "1,234.56 €"},
				{Type: "Newline", Value: "\n"},
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, tokens)
		})

		t.Run("lexes a posting with a status indicator and virtual account.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    ! (expenses:Groceries)      1,234.56 €\n")

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
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
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, tokens)
		})

		t.Run("lexes a posting with a currency conversion that includes a double space.", func(t *testing.T) {
			lexer, tokens, err := runLexer("    expenses:Groceries   -1,234.56 BTC @  1,234.50 €\n")

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Indent", Value: "    "},
				{Type: "AccountNameSegment", Value: "expenses"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "Groceries"},
				{Type: "Whitespace", Value: "   "},
				{Type: "Amount", Value: "-1,234.56 BTC @  1,234.50 €"},
				{Type: "Newline", Value: "\n"},
			})

			assert.NoError(t, err)
			assert.Equal(t, expectedTokens, tokens)
		})

		t.Run("fails on invalid inputs.", func(t *testing.T) {
			invalidInputs := []string{
				"    !expenses:Groceries\n",
				"    *expenses:Groceries\n",
				"    )expenses:Groceries\n",
				"    ]expenses:Groceries\n",
			}

			for _, invalidInput := range invalidInputs {
				_, _, err := runLexer(invalidInput)
				assert.Error(t, err)
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
				_, _, err := runLexer(validInput)
				assert.NoError(t, err)
			}
		})
	})

	t.Run("Mixed", func(t *testing.T) {
		t.Run("Lexes a journal file containing many different directives, postings and comments", func(t *testing.T) {
			lexer, tokens, err := runLexer(`; This is a cool journal file
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
`)
			assert.NoError(t, err)

			expectedTokens := makeTokens(lexer, []lextesting.MiniToken{
				{Type: "Garbage", Value: "; This is a cool journal file"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "# It includes many things"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "; This next line is empty, but contains whitespace:"},
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
				{Type: "Garbage", Value: "; indented comment"},
				{Type: "Newline", Value: "\n"},
				{Type: "AccountDirective", Value: "account"},
				{Type: "Whitespace", Value: " "},
				{Type: "AccountNameSegment", Value: "expenses"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "Gro ce"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "ries"},
				{Type: "InlineCommentIndicator", Value: "  ;"},
				{Type: "Garbage", Value: " hehe"},
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "payee Some Cool Person"},
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "commodity EUR"},
				{Type: "Newline", Value: "\n"},
				{Type: "Indent", Value: "    "},
				{Type: "Garbage", Value: "format 1,000.00 €"},
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "2024-11-25 ! (code) Payee | transaction reason  ; inline comment"},
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
				{Type: "Garbage", Value: " inline comment"},
				{Type: "Newline", Value: "\n"},
			})

			assert.Equal(t, expectedTokens, tokens)
		})
	})
}
