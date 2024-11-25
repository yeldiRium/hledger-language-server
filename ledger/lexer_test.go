package ledger_test

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/yeldiRium/hledger-language-server/ledger"
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
			definition := ledger.MakeLexerDefinition(nil, []string{})

			assert.Equal(t, lexer.TokenType(0), definition.Symbol("Error"))
			assert.Equal(t, lexer.TokenType(1), definition.Symbol("EOF"))
		})

		t.Run("can be extended with custom symbols that are automatically enumerated", func(t *testing.T) {
			definition := ledger.MakeLexerDefinition(nil, []string{"foo", "bar"})

			assert.Equal(t, lexer.TokenType(0), definition.Symbol("Error"))
			assert.Equal(t, lexer.TokenType(1), definition.Symbol("EOF"))
			assert.Equal(t, lexer.TokenType(3), definition.Symbol("foo"))
			assert.Equal(t, lexer.TokenType(4), definition.Symbol("bar"))
		})

		t.Run("Symbol", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("LexString", func(t *testing.T) {
			t.Run("runs the lexer until the end of the input is reached", func(t *testing.T) {
				var rootState ledger.StateFn
				rootState = func(l *ledger.Lexer) ledger.StateFn {
					ok, _ := l.AcceptEof()
					if ok {
						return nil
					}
					l.NextRune()
					l.Emit(1337)
					return rootState
				}

				lexerDefinition := ledger.MakeLexerDefinition(
					rootState,
					[]string{
						"Char",
					},
				)
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

			t.Run("runs the lexer until an error is encountered", func(t *testing.T) {
				var rootState ledger.StateFn
				rootState = func(l *ledger.Lexer) ledger.StateFn {
					l.Errorf("something went wrong")
					return nil
				}

				lexerDefinition := ledger.MakeLexerDefinition(
					rootState,
					[]string{
						"Char",
					},
				)
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
		t.Run("Symbol", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("NextRune", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("Peek", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("Ignore", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("Emit", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("Errorf", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("Accept", func(t *testing.T) {
			t.Run("accepts a single character, returns true and advances the position", func(t *testing.T) {
				filename := "testFile"
				l := ledger.MakeLexer(
					filename,
					ledger.MakeLexerDefinition(nil, []string{}),
					"test input",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := l.Accept("t")
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
			})

			t.Run("returns false and stays at position if the next character does not match", func(t *testing.T) {
				filename := "testFile"
				l := ledger.MakeLexer(
					filename,
					ledger.MakeLexerDefinition(nil, []string{}),
					"test input",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := l.Accept("x")
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the accept call", func(t *testing.T) {
				filename := "testFile"
				l := ledger.MakeLexer(
					filename,
					ledger.MakeLexerDefinition(nil, []string{}),
					"test input",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				_, backup := l.Accept("t")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("AcceptFn", func(t *testing.T) {
			t.Run("Accepts a single character for which the callback returns true.", func(t *testing.T) {
				filename := "testFile"
				l := ledger.MakeLexer(
					filename,
					ledger.MakeLexerDefinition(nil, []string{}),
					"0",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := l.AcceptFn(func(r rune) bool {
					return strings.IndexRune("01234", r) != -1
				})
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
			})

			t.Run("Does not accept a character for which the callback returns false.", func(t *testing.T) {
				filename := "testFile"
				l := ledger.MakeLexer(
					filename,
					ledger.MakeLexerDefinition(nil, []string{}),
					"7",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				ok, _ := l.AcceptFn(func(r rune) bool {
					return strings.IndexRune("01234", r) != -1
				})
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("AcceptRun", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("AcceptFnRun", func(t *testing.T) {
			t.Run("Accepts a run of characters for which the callback returns true.", func(t *testing.T) {
				filename := "testFile"
				l := ledger.MakeLexer(
					filename,
					ledger.MakeLexerDefinition(nil, []string{}),
					"0123456789",
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
					make(chan lexer.Token),
				)

				_ = l.AcceptRunFn(func(r rune) bool {
					return strings.IndexRune("01234", r) != -1
				})
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 6, Offset: 5}, l.Pos())
			})
		})

		t.Run("AcceptString", func(t *testing.T) {
			// TODO: add tests
		})

		t.Run("AcceptUntil", func(t *testing.T) {
			// TODO: add tests
		})
	})
}
