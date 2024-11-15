package journal

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	runLexerWithFilename := func(t *testing.T, input, filename string) ([]lexer.Token, error) {
		lexerDefinition := MakeLexer()
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
}
