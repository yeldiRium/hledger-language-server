package ledger

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	collectLexerTokens := func(lexer2 lexer.Lexer) ([]lexer.Token, error) {
		tokens := make([]lexer.Token, 0)
		for token, err := lexer2.Next(); token.Type != lexer.EOF; token, err = lexer2.Next() {
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		}

		return tokens, nil
	}

	t.Run("LexerDefinition", func(t *testing.T) {
		t.Run("contains a default list of symbols", func(t *testing.T) {
			definition := MakeLexerDefinition(nil, []string{})
			assert.Equal(t, map[string]lexer.TokenType{
				"Error": 0,
				"EOF":   1,
			}, definition.symbols)
		})

		t.Run("can be extended with custom symbols that are automatically enumerated", func(t *testing.T) {
			definition := MakeLexerDefinition(nil, []string{"foo", "bar"})
			assert.Equal(t, map[string]lexer.TokenType{
				"Error": 0,
				"EOF":   1,
				"foo":   3,
				"bar":   4,
			}, definition.symbols)
		})

		t.Run("LexString", func(t *testing.T) {
			t.Run("LexString runs the lexer until the end of the input is reached", func(t *testing.T) {
				var rootState StateFn
				rootState = func(l *Lexer) StateFn {
					rune, _ := l.NextRune()
					if rune == eof {
						return nil
					}
					l.Emit(1337)
					return rootState
				}

				lexerDefinition := &LexerDefinition{
					initialState: rootState,
					symbols: map[string]lexer.TokenType{
						"Char": 1337,
					},
				}
				lexer2, err := lexerDefinition.LexString("testFile", "foo")
				assert.NoError(t, err)

				tokens, err := collectLexerTokens(lexer2)
				assert.NoError(t, err)
				assert.Equal(t, []lexer.Token{
					{Type: 1337, Value: "f", Pos: lexer.Position{Filename: "testFile", Line: 1, Column: 1, Offset: 0}},
					{Type: 1337, Value: "o", Pos: lexer.Position{Filename: "testFile", Line: 1, Column: 2, Offset: 1}},
					{Type: 1337, Value: "o", Pos: lexer.Position{Filename: "testFile", Line: 1, Column: 3, Offset: 2}},
				}, tokens)

				token, _ := lexer2.Next()
				assert.Equal(t, lexer.EOF, token.Type)
			})

			t.Run("LexString runs the lexer until an error is encountered", func(t *testing.T) {
				var rootState StateFn
				rootState = func(l *Lexer) StateFn {
					l.Errorf("something went wrong")
					return nil
				}

				lexerDefinition := &LexerDefinition{
					initialState: rootState,
					symbols: map[string]lexer.TokenType{
						"Char": 1337,
					},
				}
				lexer2, err := lexerDefinition.LexString("testFile", "foo")
				assert.NoError(t, err)

				_, err = collectLexerTokens(lexer2)
				assert.Error(t, err, "something went wrong")

				token, _ := lexer2.Next()
				assert.Equal(t, lexer.EOF, token.Type)
			})
		})
	})

	t.Run("Lexer", func(t *testing.T) {
		t.Run("accept", func(t *testing.T) {
			t.Run("accepts a single character, returns true and advances the position", func(t *testing.T) {
				filename := "testFile"
				l := &Lexer{
					name:   filename,
					input:  "test input",
					start:  lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					pos:    lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					tokens: make(chan lexer.Token),
				}

				ok, _ := l.Accept("t")
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.pos)
			})

			t.Run("returns false and stays at position if the next character does not match", func(t *testing.T) {
				filename := "testFile"
				l := &Lexer{
					name:   filename,
					input:  "test input",
					start:  lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					pos:    lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					tokens: make(chan lexer.Token),
				}

				ok, _ := l.Accept("x")
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.pos)
			})

			t.Run("returns a backup function that rewinds the position to before the accept call", func(t *testing.T) {
				filename := "testFile"
				l := &Lexer{
					name:   filename,
					input:  "test input",
					start:  lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					pos:    lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					tokens: make(chan lexer.Token),
				}

				_, backup := l.Accept("t")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.pos)
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.pos)
			})
		})
	})
}
