package ledger

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

func TestJournalLexer(t *testing.T) {
	runLexerWithFilename := func(t *testing.T, input, filename string) ([]lexer.Token, error) {
		lexerDefinition := MakeJournalLexer()
		l, err := lexerDefinition.LexString(filename, input)
		if err != nil {
			return nil, err
		}
		tokens := make([]lexer.Token, 0)
		for token, err := l.Next(); token.Type != lexer.EOF; token, err = l.Next() {
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		}

		fmt.Printf("tokens: %#v\n", tokens)

		return tokens, nil
	}
	runLexer := func(t *testing.T, input string) ([]lexer.Token, error) {
		return runLexerWithFilename(t, input, "testFile")
	}

	t.Run("Miscellaneous", func(t *testing.T) {
		t.Run("Succeeds on an empty input.", func(t *testing.T) {
			tokens, err := runLexer(t, "")
			assert.NoError(t, err)
			assert.Equal(t, []lexer.Token{}, tokens)
		})
	})

	t.Run("Lexes newlines", func(t *testing.T) {
		tokens, err := runLexer(t, "\n\n")
		assert.NoError(t, err)
		assert.Equal(t, []lexer.Token{
			{
				Type:  itemNewline,
				Value: "\n",
				Pos:   lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1},
			},
			{
				Type:  itemNewline,
				Value: "\n",
				Pos:   lexer.Position{Filename: "testFile", Offset: 1, Line: 2, Column: 1},
			},
		}, tokens)
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Fails if the account directive does not contain an account name.", func(t *testing.T) {
			_, err := runLexer(t, "account \n")
			assert.Error(t, err, "foobar")
		})

		t.Run("Parses a file containing an account directive with multiple segments", func(t *testing.T) {
			tokens, err := runLexer(t, "account assets:Cash:Checking\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				[]lexer.Token{
					{Type: itemAccountDirective, Value: "account", Pos: lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1}},
					{Type: itemWhitespace, Value: " ", Pos: lexer.Position{Filename: "testFile", Offset: 7, Line: 1, Column: 8}},
					{Type: itemAccountName, Value: "assets:Cash:Checking", Pos: lexer.Position{Filename: "testFile", Offset: 8, Line: 1, Column: 9}},
					{Type: itemNewline, Value: "\n", Pos: lexer.Position{Filename: "testFile", Offset: 28, Line: 1, Column: 29}},
				},
				tokens,
			)
		})

		t.Run("Parses a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			tokens, err := runLexer(t, "account assets:Cash:Che cking:Spe-ci_al\n")
			assert.NoError(t, err)
			assert.Equal(
				t,
				[]lexer.Token{
					{Type: itemAccountDirective, Value: "account", Pos: lexer.Position{Filename: "testFile", Offset: 0, Line: 1, Column: 1}},
					{Type: itemWhitespace, Value: " ", Pos: lexer.Position{Filename: "testFile", Offset: 7, Line: 1, Column: 8}},
					{Type: itemAccountName, Value: "assets:Cash:Che cking:Spe-ci_al", Pos: lexer.Position{Filename: "testFile", Offset: 8, Line: 1, Column: 9}},
					{Type: itemNewline, Value: "\n", Pos: lexer.Position{Filename: "testFile", Offset: 39, Line: 1, Column: 40}},
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

		//t.Run("Fails on more than one consecutive space within an account name", func(t *testing.T) {
		//	testParserFails(
		//		t,
		//		"account assets:Cash:Che  cking\n",
		//		"testFile:1:24: unexpected token \" \" (expected <newline>)",
		//	)
		//})
	})
}
