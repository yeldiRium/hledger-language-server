package ledger_test

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/yeldiRium/hledger-language-server/ledger"
)

func TestJournalLexer(t *testing.T) {
	filename := "testFile"
	runLexerWithFilename := func(input string) (*ledger.Lexer, []lexer.Token, error) {
		lexerDefinition := ledger.MakeJournalLexer()
		l, err := lexerDefinition.LexString(filename, input)
		if err != nil {
			return nil, nil, err
		}
		tokens := make([]lexer.Token, 0)
		for token, err := l.Next(); token.Type != lexer.EOF; token, err = l.Next() {
			if err != nil {
				return nil, nil, err
			}
			tokens = append(tokens, token)
		}

		fmt.Printf("tokens: %#v\n", tokens)

		return l, tokens, nil
	}
	runLexer := func(input string) (*ledger.Lexer, []lexer.Token, error) {
		return runLexerWithFilename(input)
	}

	type MiniToken struct {
		Type  string
		Value string
	}
	makeTokens := func(l *ledger.Lexer, miniTokens []MiniToken) []lexer.Token {
		tokens := make([]lexer.Token, len(miniTokens))
		pos := lexer.Position{Filename: filename, Offset: 0, Line: 1, Column: 1}
		for i, token := range miniTokens {
			tokens[i] = lexer.Token{
				Type:  l.Symbol(token.Type),
				Value: token.Value,
				Pos:   pos,
			}
			pos.Advance(token.Value)
		}
		return tokens
	}

	t.Run("Miscellaneous", func(t *testing.T) {
		t.Run("Succeeds on an empty input.", func(t *testing.T) {
			_, tokens, err := runLexer("")
			assert.NoError(t, err)
			assert.Equal(t, []lexer.Token{}, tokens)
		})

		t.Run("Lexes newlines", func(t *testing.T) {
			l, tokens, err := runLexer("\n\n")
			assert.NoError(t, err)
			assert.Equal(t, makeTokens(l, []MiniToken{
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
			}), tokens)
		})

		t.Run("Lexes garbage and newlines.", func(t *testing.T) {
			l, tokens, err := runLexer("this is not a valid journal file\n\n heckmeck\n")
			assert.NoError(t, err)
			assert.Equal(t, makeTokens(l, []MiniToken{
				{Type: "Garbage", Value: "this is not a valid journal file"},
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: " heckmeck"},
				{Type: "Newline", Value: "\n"},
			}), tokens)
		})
	})

	t.Run("Helpers", func(t *testing.T) {
		t.Run("AcceptInlineCommentIndicator", func(t *testing.T) {
			t.Run("accepts an inline comment indicator using a semicolon.", func(t *testing.T) {
				filename := filename
				lexerDefinition := ledger.MakeLexerDefinition(
					func(l *ledger.Lexer) ledger.StateFn { return nil },
					[]string{
						"Char",
					},
				)
				l := ledger.MakeLexer(
					filename,
					lexerDefinition,
					"  ; foo",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := ledger.AcceptInlineCommentIndicator(l)
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, l.Pos())
			})

			t.Run("accepts an inline comment indicator using a hash.", func(t *testing.T) {
				filename := filename
				lexerDefinition := ledger.MakeLexerDefinition(
					func(l *ledger.Lexer) ledger.StateFn { return nil },
					[]string{
						"Char",
					},
				)
				l := ledger.MakeLexer(
					filename,
					lexerDefinition,
					"  # foo",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := ledger.AcceptInlineCommentIndicator(l)
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, l.Pos())
			})

			t.Run("does not accept things not starting with spaces.", func(t *testing.T) {
				filename := filename
				lexerDefinition := ledger.MakeLexerDefinition(
					func(l *ledger.Lexer) ledger.StateFn { return nil },
					[]string{
						"Char",
					},
				)
				l := ledger.MakeLexer(
					filename,
					lexerDefinition,
					"foo",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := ledger.AcceptInlineCommentIndicator(l)
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})

			t.Run("does not accept inline comments starting with anything else.", func(t *testing.T) {
				filename := filename
				lexerDefinition := ledger.MakeLexerDefinition(
					func(l *ledger.Lexer) ledger.StateFn { return nil },
					[]string{
						"Char",
					},
				)
				l := ledger.MakeLexer(
					filename,
					lexerDefinition,
					"  foo",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := ledger.AcceptInlineCommentIndicator(l)
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Ignores missing accout names, since that is the parser's responsibility.", func(t *testing.T) {
			_, _, err := runLexer("account \n")
			assert.NoError(t, err)
		})

		t.Run("Lexes a file containing an account directive with multiple segments", func(t *testing.T) {
			l, tokens, err := runLexer("account assets:Cash:Checking\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				makeTokens(l, []MiniToken{
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
			l, tokens, err := runLexer("account assets:Cash:Che cking:Spe-ci_al\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				makeTokens(l, []MiniToken{
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
			l, tokens, err := runLexer("account assets:Cash:Checking  ; inline comment\n")
			assert.NoError(t, err)
			assert.Greater(t, len(tokens), 7)
			assert.Equal(
				t,
				[]lexer.Token{
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: filename, Offset: 0, Line: 1, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: filename, Offset: 7, Line: 1, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "assets", Pos: lexer.Position{Filename: filename, Offset: 8, Line: 1, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: filename, Offset: 14, Line: 1, Column: 15}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Cash", Pos: lexer.Position{Filename: filename, Offset: 15, Line: 1, Column: 16}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: filename, Offset: 19, Line: 1, Column: 20}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Checking", Pos: lexer.Position{Filename: filename, Offset: 20, Line: 1, Column: 21}},
				},
				tokens[0:7],
			)
		})

		t.Run("Lexes a file containing multiple account directives", func(t *testing.T) {
			l, tokens, err := runLexer(`account assets:Cash:Checking
account expenses:Food:Groceries
`)
			assert.NoError(t, err)
			assert.Equal(
				t,
				[]lexer.Token{
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: filename, Offset: 0, Line: 1, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: filename, Offset: 7, Line: 1, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "assets", Pos: lexer.Position{Filename: filename, Offset: 8, Line: 1, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: filename, Offset: 14, Line: 1, Column: 15}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Cash", Pos: lexer.Position{Filename: filename, Offset: 15, Line: 1, Column: 16}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: filename, Offset: 19, Line: 1, Column: 20}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Checking", Pos: lexer.Position{Filename: filename, Offset: 20, Line: 1, Column: 21}},
					{Type: l.Symbol("Newline"), Value: "\n", Pos: lexer.Position{Filename: filename, Offset: 28, Line: 1, Column: 29}},
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: filename, Offset: 29, Line: 2, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: filename, Offset: 36, Line: 2, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "expenses", Pos: lexer.Position{Filename: filename, Offset: 37, Line: 2, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: filename, Offset: 45, Line: 2, Column: 17}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Food", Pos: lexer.Position{Filename: filename, Offset: 46, Line: 2, Column: 18}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: filename, Offset: 50, Line: 2, Column: 22}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Groceries", Pos: lexer.Position{Filename: filename, Offset: 51, Line: 2, Column: 23}},
					{Type: l.Symbol("Newline"), Value: "\n", Pos: lexer.Position{Filename: filename, Offset: 60, Line: 2, Column: 32}},
				},
				tokens,
			)
		})

		//t.Run("Parses a file containing an account directive followed by an inline comment", func(t *testing.T) {
		//	testParserWithFileContent(t, "account assets:Cash:Checking  ; hehe\n", &Journal{
		//		Entries: []Entry{
		//			&AccountDirective{
		//				AccountName: &AccountName{
		//					Segments: []string{"assets", "Cash", "Checking"},
		//				},
		//				Comment: &InlineComment{
		//					String: "hehe",
		//				},
		//			},
		//		},
		//	})
		//})

		t.Run("Stops lexing the account name after a double space.", func(t *testing.T) {
			_, _, err := runLexer("account assets:Cash:Che  cking\n")
			assert.ErrorContains(t, err, "expected account name segment, but found nothing")
		})
	})

	t.Run("Mixed", func(t *testing.T) {
		t.Run("Lexes a journal file containing many different directives, postings and comments", func(t *testing.T) {
			l, tokens, err := runLexer(`; This is a cool journal file
# It includes many things
account assets:Cash:Checking
account expenses:Gro ce:ries  ; hehe

payee Some Cool Person

2024-11-25 ! (code) Payee | transaction reason  ; inline comment
    expenses:Groceries      1,234.56 €
    assets:Cash:Checking   -1,234.56 €  ; inline comment
`)

			assert.NoError(t, err)
			expectedTokens := makeTokens(l, []MiniToken{
				{Type: "Garbage", Value: "; This is a cool journal file"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "# It includes many things"},
				{Type: "Newline", Value: "\n"},
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
				{Type: "AccountNameSegment", Value: "Gro ce"},
				{Type: "AccountNameSeparator", Value: ":"},
				{Type: "AccountNameSegment", Value: "ries"},
				{Type: "Garbage", Value: "  ; hehe"},
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "payee Some Cool Person"},
				{Type: "Newline", Value: "\n"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "2024-11-25 ! (code) Payee | transaction reason  ; inline comment"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "    expenses:Groceries      1,234.56 €"},
				{Type: "Newline", Value: "\n"},
				{Type: "Garbage", Value: "    assets:Cash:Checking   -1,234.56 €  ; inline comment"},
				{Type: "Newline", Value: "\n"},
			})
			assert.Equal(t, expectedTokens, tokens)
		})
	})
}
