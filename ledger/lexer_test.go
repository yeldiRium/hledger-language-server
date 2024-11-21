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

	t.Run("journalLexer", func(t *testing.T) {
		t.Run("accept", func(t *testing.T) {
			t.Run("accepts a single character, returns true and advances the position", func(t *testing.T) {
				filename := "testFile"
				l := &journalLexer{
					name:   filename,
					input:  "test input",
					start:  lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					pos:    lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					tokens: make(chan lexer.Token),
				}

				ok, _ := l.accept("t")
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.pos)
			})

			t.Run("returns false and stays at position if the next character does not match", func(t *testing.T) {
				filename := "testFile"
				l := &journalLexer{
					name:   filename,
					input:  "test input",
					start:  lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					pos:    lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					tokens: make(chan lexer.Token),
				}

				ok, _ := l.accept("x")
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.pos)
			})

			t.Run("returns a backup function that rewinds the position to before the accept call", func(t *testing.T) {
				filename := "testFile"
				l := &journalLexer{
					name:   filename,
					input:  "test input",
					start:  lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					pos:    lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					tokens: make(chan lexer.Token),
				}

				_, backup := l.accept("t")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.pos)
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.pos)
			})
		})
	})

	t.Run("LexString", func(t *testing.T) {
		t.Run("LexString runs the lexer until the end of the input is reached", func(t *testing.T) {
			var rootState stateFn
			rootState = func(l *journalLexer) stateFn {
				rune, _ := l.next()
				if rune == eof {
					return nil
				}
				l.emit(1337)
				return rootState
			}

			lexerDefinition := &journalLexerDefinition{
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
			var rootState stateFn
			rootState = func(l *journalLexer) stateFn {
				l.errorf("something went wrong")
				return nil
			}

			lexerDefinition := &journalLexerDefinition{
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
}
