package ledger_test

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/yeldiRium/hledger-language-server/ledger"
)

func TestJournalLexer(t *testing.T) {
	runLexerWithFilename := func(input, filename string) (*ledger.Lexer, []lexer.Token, error) {
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
		return runLexerWithFilename(input, "testFile")
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
			assert.Equal(t, []lexer.Token{
				{
					Type:  l.Symbol("Newline"),
					Value: "\n",
					Pos:   lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1},
				},
				{
					Type:  l.Symbol("Newline"),
					Value: "\n",
					Pos:   lexer.Position{Filename: "testFile", Offset: 1, Line: 2, Column: 1},
				},
			}, tokens)
		})

		t.Run("Lexes garbage and newlines.", func(t *testing.T) {
			l, tokens, err := runLexer("this is not a valid journal file\n\n heckmeck\n")
			assert.NoError(t, err)
			assert.Equal(t, []lexer.Token{
				{
					Type:  l.Symbol("Garbage"),
					Value: "this is not a valid journal file",
					Pos:   lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1},
				},
				{
					Type:  l.Symbol("Newline"),
					Value: "\n",
					Pos:   lexer.Position{Filename: "testFile", Offset: 32, Line: 1, Column: 33},
				},
				{
					Type:  l.Symbol("Newline"),
					Value: "\n",
					Pos:   lexer.Position{Filename: "testFile", Offset: 33, Line: 2, Column: 1},
				},
				{
					Type:  l.Symbol("Garbage"),
					Value: " heckmeck",
					Pos:   lexer.Position{Filename: "testFile", Offset: 34, Line: 3, Column: 1},
				},
				{
					Type:  l.Symbol("Newline"),
					Value: "\n",
					Pos:   lexer.Position{Filename: "testFile", Offset: 43, Line: 3, Column: 10},
				},
			}, tokens)
		})
	})

	t.Run("Helpers", func(t *testing.T) {
		t.Run("AcceptInlineCommentIndicator", func(t *testing.T) {
			t.Run("accepts an inline comment indicator using a semicolon.", func(t *testing.T) {
				filename := "testFile"
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
				filename := "testFile"
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
				filename := "testFile"
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
				filename := "testFile"
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
				[]lexer.Token{
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: "testFile", Offset: 7, Line: 1, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "assets", Pos: lexer.Position{Filename: "testFile", Offset: 8, Line: 1, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 14, Line: 1, Column: 15}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Cash", Pos: lexer.Position{Filename: "testFile", Offset: 15, Line: 1, Column: 16}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 19, Line: 1, Column: 20}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Checking", Pos: lexer.Position{Filename: "testFile", Offset: 20, Line: 1, Column: 21}},
					{Type: l.Symbol("Newline"), Value: "\n", Pos: lexer.Position{Filename: "testFile", Offset: 28, Line: 1, Column: 29}},
				},
				tokens,
			)
		})

		t.Run("Lexes a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			l, tokens, err := runLexer("account assets:Cash:Che cking:Spe-ci_al\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				[]lexer.Token{
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: "testFile", Offset: 7, Line: 1, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "assets", Pos: lexer.Position{Filename: "testFile", Offset: 8, Line: 1, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 14, Line: 1, Column: 15}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Cash", Pos: lexer.Position{Filename: "testFile", Offset: 15, Line: 1, Column: 16}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 19, Line: 1, Column: 20}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Che cking", Pos: lexer.Position{Filename: "testFile", Offset: 20, Line: 1, Column: 21}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 29, Line: 1, Column: 30}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Spe-ci_al", Pos: lexer.Position{Filename: "testFile", Offset: 30, Line: 1, Column: 31}},
					{Type: l.Symbol("Newline"), Value: "\n", Pos: lexer.Position{Filename: "testFile", Offset: 39, Line: 1, Column: 40}},
				},
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
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: "testFile", Offset: 7, Line: 1, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "assets", Pos: lexer.Position{Filename: "testFile", Offset: 8, Line: 1, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 14, Line: 1, Column: 15}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Cash", Pos: lexer.Position{Filename: "testFile", Offset: 15, Line: 1, Column: 16}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 19, Line: 1, Column: 20}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Checking", Pos: lexer.Position{Filename: "testFile", Offset: 20, Line: 1, Column: 21}},
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
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: "testFile", Offset: 7, Line: 1, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "assets", Pos: lexer.Position{Filename: "testFile", Offset: 8, Line: 1, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 14, Line: 1, Column: 15}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Cash", Pos: lexer.Position{Filename: "testFile", Offset: 15, Line: 1, Column: 16}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 19, Line: 1, Column: 20}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Checking", Pos: lexer.Position{Filename: "testFile", Offset: 20, Line: 1, Column: 21}},
					{Type: l.Symbol("Newline"), Value: "\n", Pos: lexer.Position{Filename: "testFile", Offset: 28, Line: 1, Column: 29}},
					{Type: l.Symbol("AccountDirective"), Value: "account", Pos: lexer.Position{Filename: "testFile", Offset: 29, Line: 2, Column: 1}},
					{Type: l.Symbol("Whitespace"), Value: " ", Pos: lexer.Position{Filename: "testFile", Offset: 36, Line: 2, Column: 8}},
					{Type: l.Symbol("AccountNameSegment"), Value: "expenses", Pos: lexer.Position{Filename: "testFile", Offset: 37, Line: 2, Column: 9}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 45, Line: 2, Column: 17}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Food", Pos: lexer.Position{Filename: "testFile", Offset: 46, Line: 2, Column: 18}},
					{Type: l.Symbol("AccountNameSeparator"), Value: ":", Pos: lexer.Position{Filename: "testFile", Offset: 50, Line: 2, Column: 22}},
					{Type: l.Symbol("AccountNameSegment"), Value: "Groceries", Pos: lexer.Position{Filename: "testFile", Offset: 51, Line: 2, Column: 23}},
					{Type: l.Symbol("Newline"), Value: "\n", Pos: lexer.Position{Filename: "testFile", Offset: 60, Line: 2, Column: 32}},
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
}
