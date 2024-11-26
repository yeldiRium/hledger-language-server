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
		t.Run("contains a default list of symbols.", func(t *testing.T) {
			definition := ledger.MakeLexerDefinition(nil, []string{})

			assert.Equal(t, lexer.TokenType(0), definition.Symbol("Error"))
			assert.Equal(t, lexer.TokenType(1), definition.Symbol("EOF"))
		})

		t.Run("can be extended with custom symbols that are automatically enumerated.", func(t *testing.T) {
			definition := ledger.MakeLexerDefinition(nil, []string{"foo", "bar"})

			assert.Equal(t, lexer.TokenType(0), definition.Symbol("Error"))
			assert.Equal(t, lexer.TokenType(1), definition.Symbol("EOF"))
			assert.Equal(t, lexer.TokenType(3), definition.Symbol("foo"))
			assert.Equal(t, lexer.TokenType(4), definition.Symbol("bar"))
		})

		t.Run("Symbol", func(t *testing.T) {
			t.Run("returns a TokenType for the given token name.", func(t *testing.T) {
				definition := ledger.MakeLexerDefinition(nil, []string{"foo", "bar"})

				symbol := definition.Symbol("foo")
				assert.Equal(t, lexer.TokenType(3), symbol)
			})

			t.Run("panics if the token name is unknown.", func(t *testing.T) {
				definition := ledger.MakeLexerDefinition(nil, []string{"foo", "bar"})

				assert.Panics(t, func() {
					definition.Symbol("unknown")
				})
			})
		})

		t.Run("LexString", func(t *testing.T) {
			t.Run("runs the lexer until the end of the input is reached.", func(t *testing.T) {
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

			t.Run("runs the lexer until an error is encountered.", func(t *testing.T) {
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
		prepareLexer := func(input string, tokenNames []string, rootState ledger.StateFn) (l *ledger.Lexer, filename string) {
			filename = "testFile"
			definition := ledger.MakeLexerDefinition(rootState, tokenNames)
			l = ledger.MakeLexer(
				filename,
				definition,
				input,
				lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
				lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
				make(chan lexer.Token),
			)

			return l, filename
		}

		t.Run("Symbol", func(t *testing.T) {
			t.Run("returns a TokenType for the given token name.", func(t *testing.T) {
				l, _ := prepareLexer("test input", []string{"foo", "bar"}, nil)

				symbol := l.Symbol("foo")
				assert.Equal(t, lexer.TokenType(3), symbol)
			})

			t.Run("panics if the token name is unknown.", func(t *testing.T) {
				l, _ := prepareLexer("test input", []string{"foo", "bar"}, nil)

				assert.Panics(t, func() {
					l.Symbol("unknown")
				})
			})
		})

		t.Run("NextRune", func(t *testing.T) {
			t.Run("returns the next rune in the input and moves the position forward by one.", func(t *testing.T) {
				l, filename := prepareLexer("foo", []string{}, nil)

				r, _ := l.NextRune()
				assert.Equal(t, 'f', r)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
			})

			t.Run("returns a backup function that moves the position back to before the rune.", func(t *testing.T) {
				l, filename := prepareLexer("foo", []string{}, nil)

				_, backup := l.NextRune()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("Peek", func(t *testing.T) {
			t.Run("returns the next rune in the input without moving the position.", func(t *testing.T) {
				l, filename := prepareLexer("foo", []string{}, nil)

				r := l.Peek()
				assert.Equal(t, 'f', r)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("Emit", func(t *testing.T) {
			t.Run("emits the lexed string between start and pos with the given token type, then starts a new token by setting start to pos.", func(t *testing.T) {
				l, filename := prepareLexer("foo", []string{"String"}, nil)

				ok, _ := l.AcceptString("foo")
				assert.True(t, ok)
				go func() {
					l.Emit(l.Symbol("String"))
				}()

				token, _ := l.Next()
				assert.Equal(t, lexer.Token{Type: l.Symbol("String"), Value: "foo", Pos: lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}}, token)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, l.Pos())
			})
		})

		t.Run("Ignore", func(t *testing.T) {
			t.Run("ignores the lexed string between start and pos and starts a new token by setting start to pos.", func(t *testing.T) {
				l, filename := prepareLexer("foo", []string{"String"}, nil)

				ok, _ := l.AcceptString("foo")
				assert.True(t, ok)
				l.Ignore()

				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, l.Start())
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, l.Pos())
			})
		})

		t.Run("Errorf", func(t *testing.T) {
			t.Run("emits an error token with the given message.", func(t *testing.T) {
				l, _ := prepareLexer("foo", []string{"String"}, nil)

				go func() {
					l.Errorf("something went wrong")
				}()

				_, err := l.Next()
				assert.ErrorContains(t, err, "something went wrong")
			})
		})

		t.Run("Accept", func(t *testing.T) {
			t.Run("accepts a single character, returns true and advances the position.", func(t *testing.T) {
				l, filename := prepareLexer("test input", []string{}, nil)

				ok, _ := l.Accept("t")
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
			})

			t.Run("returns false and stays at position if the next character does not match.", func(t *testing.T) {
				l, filename := prepareLexer("test input", []string{}, nil)

				ok, _ := l.Accept("x")
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the accept call.", func(t *testing.T) {
				l, filename := prepareLexer("test input", []string{}, nil)

				_, backup := l.Accept("t")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("AcceptFn", func(t *testing.T) {
			t.Run("accepts a single character for which the callback returns true.", func(t *testing.T) {
				l, filename := prepareLexer("0", []string{}, nil)

				ok, _ := l.AcceptFn(func(r rune) bool {
					return strings.IndexRune("01234", r) != -1
				})
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, l.Pos())
			})

			t.Run("does not accept a character for which the callback returns false.", func(t *testing.T) {
				l, filename := prepareLexer("7", []string{}, nil)

				ok, _ := l.AcceptFn(func(r rune) bool {
					return strings.IndexRune("01234", r) != -1
				})
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("AcceptRun", func(t *testing.T) {
			t.Run("accepts any number of characters from the valid set and advances the position.", func(t *testing.T) {
				l, filename := prepareLexer("0000001111", []string{}, nil)

				_ = l.AcceptRun("0")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 7, Offset: 6}, l.Pos())
			})

			t.Run("does nothing if the next rune is not in the set of valid characters.", func(t *testing.T) {
				l, filename := prepareLexer("0000001111", []string{}, nil)

				_ = l.AcceptRun("5")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptRun call.", func(t *testing.T) {
				l, filename := prepareLexer("0000001111", []string{}, nil)

				backup := l.AcceptRun("0")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 7, Offset: 6}, l.Pos())
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("AcceptFnRun", func(t *testing.T) {
			t.Run("accepts a run of characters for which the callback returns true.", func(t *testing.T) {
				l, filename := prepareLexer("0123456789", []string{}, nil)

				_ = l.AcceptRunFn(func(r rune) bool {
					return strings.IndexRune("01234", r) != -1
				})
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 6, Offset: 5}, l.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptRunFn call.", func(t *testing.T) {
				l, filename := prepareLexer("0123456789", []string{}, nil)

				backup := l.AcceptRunFn(func(r rune) bool {
					return strings.IndexRune("01234", r) != -1
				})
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 6, Offset: 5}, l.Pos())
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("AcceptString", func(t *testing.T) {
			t.Run("accepts the full given string and advances the position.", func(t *testing.T) {
				l, filename := prepareLexer("test input", []string{}, nil)

				ok, _ := l.AcceptString("test")
				assert.True(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 5, Offset: 4}, l.Pos())
			})

			t.Run("returns false and stays at position if the next characters do not match.", func(t *testing.T) {
				l, filename := prepareLexer("test input", []string{}, nil)

				ok, _ := l.AcceptString("foo")
				assert.False(t, ok)
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptString call.", func(t *testing.T) {
				l, filename := prepareLexer("test input", []string{}, nil)

				_, backup := l.AcceptString("test")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 5, Offset: 4}, l.Pos())
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})

		t.Run("AcceptUntil", func(t *testing.T) {
			t.Run("accepts any number of characters until an invalid character is encountered and advances the position.", func(t *testing.T) {
				l, filename := prepareLexer("ngqxflguiasordsoxl0xeflq", []string{}, nil)

				_ = l.AcceptUntil("0")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 19, Offset: 18}, l.Pos())
			})

			t.Run("accepts any number of characters until EOF is encountered and advances the position.", func(t *testing.T) {
				l, filename := prepareLexer("nlgeqxgenui", []string{}, nil)

				_ = l.AcceptUntil("0")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 12, Offset: 11}, l.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptUntil call.", func(t *testing.T) {
				l, filename := prepareLexer("ngqxflguiasordsoxl0xeflq", []string{}, nil)

				backup := l.AcceptUntil("0")
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 19, Offset: 18}, l.Pos())
				backup()
				assert.Equal(t, lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, l.Pos())
			})
		})
	})
}
