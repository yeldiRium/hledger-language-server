package lexing_test

import (
	"strings"
	"testing"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"

	"github.com/yeldiRium/hledger-language-server/lexing"
)

func TestLexer(t *testing.T) {
	collectLexerTokens := func(lexer participleLexer.Lexer) ([]participleLexer.Token, error) {
		tokens := make([]participleLexer.Token, 0)
		for token, err := lexer.Next(); token.Type != participleLexer.EOF; token, err = lexer.Next() {
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		}

		return tokens, nil
	}

	t.Run("LexerDefinition", func(t *testing.T) {
		t.Run("contains a default list of symbols.", func(t *testing.T) {
			definition := lexing.NewLexerDefinition(nil, []string{})

			assert.Equal(t, participleLexer.TokenType(0), definition.Symbol("Error"))
			assert.Equal(t, participleLexer.TokenType(1), definition.Symbol("EOF"))
		})

		t.Run("can be extended with custom symbols that are automatically enumerated.", func(t *testing.T) {
			definition := lexing.NewLexerDefinition(nil, []string{"foo", "bar"})

			assert.Equal(t, participleLexer.TokenType(0), definition.Symbol("Error"))
			assert.Equal(t, participleLexer.TokenType(1), definition.Symbol("EOF"))
			assert.Equal(t, participleLexer.TokenType(3), definition.Symbol("foo"))
			assert.Equal(t, participleLexer.TokenType(4), definition.Symbol("bar"))
		})

		t.Run("Symbol", func(t *testing.T) {
			t.Run("returns a TokenType for the given token name.", func(t *testing.T) {
				definition := lexing.NewLexerDefinition(nil, []string{"foo", "bar"})

				symbol := definition.Symbol("foo")
				assert.Equal(t, participleLexer.TokenType(3), symbol)
			})

			t.Run("panics if the token name is unknown.", func(t *testing.T) {
				definition := lexing.NewLexerDefinition(nil, []string{"foo", "bar"})

				assert.Panics(t, func() {
					definition.Symbol("unknown")
				})
			})
		})

		t.Run("LexString", func(t *testing.T) {
			t.Run("runs the lexer until the end of the input is reached.", func(t *testing.T) {
				var rootState lexing.StateFn
				rootState = func(lexer *lexing.Lexer) lexing.StateFn {
					ok, _ := lexer.AcceptEof()
					if ok {
						return nil
					}
					lexer.NextRune()
					lexer.Emit(1337)
					return rootState
				}

				lexerDefinition := lexing.NewLexerDefinition(
					rootState,
					[]string{
						"Char",
					},
				)
				lexer, err := lexerDefinition.LexString("testFile", "foo")
				assert.NoError(t, err)

				tokens, err := collectLexerTokens(lexer)
				assert.NoError(t, err)
				assert.Equal(t, []participleLexer.Token{
					{Type: 1337, Value: "f", Pos: participleLexer.Position{Filename: "testFile", Line: 1, Column: 1, Offset: 0}},
					{Type: 1337, Value: "o", Pos: participleLexer.Position{Filename: "testFile", Line: 1, Column: 2, Offset: 1}},
					{Type: 1337, Value: "o", Pos: participleLexer.Position{Filename: "testFile", Line: 1, Column: 3, Offset: 2}},
				}, tokens)

				token, _ := lexer.Next()
				assert.Equal(t, participleLexer.EOF, token.Type)
			})

			t.Run("runs the lexer until an error is encountered.", func(t *testing.T) {
				rootState := func(lexer *lexing.Lexer) lexing.StateFn {
					lexer.Errorf("something went wrong")
					return nil
				}

				lexerDefinition := lexing.NewLexerDefinition(
					rootState,
					[]string{
						"Char",
					},
				)
				lexer, err := lexerDefinition.LexString("testFile", "foo")
				assert.NoError(t, err)

				_, err = collectLexerTokens(lexer)
				assert.Error(t, err, "something went wrong")

				token, _ := lexer.Next()
				assert.Equal(t, participleLexer.EOF, token.Type)
			})
		})
	})

	t.Run("Lexer", func(t *testing.T) {
		prepareLexer := func(input string, tokenNames []string, rootState lexing.StateFn) (lexer *lexing.Lexer, filename string) {
			filename = "testFile"
			definition := lexing.NewLexerDefinition(rootState, tokenNames)
			lexer = lexing.NewLexer(
				filename,
				definition,
				input,
				participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
				participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
				make(chan participleLexer.Token),
			)

			return lexer, filename
		}

		t.Run("Symbol", func(t *testing.T) {
			t.Run("returns a TokenType for the given token name.", func(t *testing.T) {
				lexer, _ := prepareLexer("test input", []string{"foo", "bar"}, nil)

				symbol := lexer.Symbol("foo")
				assert.Equal(t, participleLexer.TokenType(3), symbol)
			})

			t.Run("panics if the token name is unknown.", func(t *testing.T) {
				lexer, _ := prepareLexer("test input", []string{"foo", "bar"}, nil)

				assert.Panics(t, func() {
					lexer.Symbol("unknown")
				})
			})
		})

		t.Run("NextRune", func(t *testing.T) {
			t.Run("returns the next rune in the input and moves the position forward by one.", func(t *testing.T) {
				lexer, filename := prepareLexer("foo", []string{}, nil)

				r, _ := lexer.NextRune()
				assert.Equal(t, 'f', r)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, lexer.Pos())
			})

			t.Run("returns a backup function that moves the position back to before the rune.", func(t *testing.T) {
				lexer, filename := prepareLexer("foo", []string{}, nil)

				_, backup := lexer.NextRune()
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, lexer.Pos())
				backup()
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})
		})

		t.Run("Peek", func(t *testing.T) {
			t.Run("returns the next rune in the input without moving the position.", func(t *testing.T) {
				lexer, filename := prepareLexer("foo", []string{}, nil)

				r := lexer.Peek()
				assert.Equal(t, 'f', r)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})
		})

		t.Run("Emit", func(t *testing.T) {
			t.Run("emits the lexed string between start and pos with the given token type, then starts a new token by setting start to pos.", func(t *testing.T) {
				lexer, filename := prepareLexer("foo", []string{"String"}, nil)

				ok, _, err := lexer.AcceptString("foo")
				assert.NoError(t, err)
				assert.True(t, ok)
				go func() {
					lexer.Emit(lexer.Symbol("String"))
				}()

				token, _ := lexer.Next()
				assert.Equal(t, participleLexer.Token{Type: lexer.Symbol("String"), Value: "foo", Pos: participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}}, token)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, lexer.Pos())
			})
		})

		t.Run("Ignore", func(t *testing.T) {
			t.Run("ignores the lexed string between start and pos and starts a new token by setting start to pos.", func(t *testing.T) {
				lexer, filename := prepareLexer("foo", []string{"String"}, nil)

				ok, _, err := lexer.AcceptString("foo")
				assert.NoError(t, err)
				assert.True(t, ok)
				lexer.Ignore()

				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, lexer.Start())
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 4, Offset: 3}, lexer.Pos())
			})
		})

		t.Run("Errorf", func(t *testing.T) {
			t.Run("emits an error token with the given message.", func(t *testing.T) {
				lexer, _ := prepareLexer("foo", []string{"String"}, nil)

				go func() {
					lexer.Errorf("something went wrong")
				}()

				_, err := lexer.Next()
				assert.ErrorContains(t, err, "something went wrong")
			})
		})

		t.Run("Accept", func(t *testing.T) {
			t.Run("accepts a single character, returns true and advances the position.", func(t *testing.T) {
				lexer, filename := prepareLexer("test input", []string{}, nil)

				ok, _, err := lexer.Accept("t")
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, lexer.Pos())
			})

			t.Run("returns false and stays at position if the next character does not match.", func(t *testing.T) {
				lexer, filename := prepareLexer("test input", []string{}, nil)

				ok, _, err := lexer.Accept("x")
				assert.NoError(t, err)
				assert.False(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the accept call.", func(t *testing.T) {
				lexer, filename := prepareLexer("test input", []string{}, nil)

				_, backup, err := lexer.Accept("t")
				assert.NoError(t, err)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, lexer.Pos())
				backup()
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns an error when encountering EOF.", func(t *testing.T) {
				lexer, _ := prepareLexer("", []string{}, nil)

				_, _, err := lexer.Accept("t")
				assert.ErrorIs(t, err, lexing.ErrEof)
			})
		})

		t.Run("AcceptFn", func(t *testing.T) {
			t.Run("accepts a single character for which the callback returns true.", func(t *testing.T) {
				lexer, filename := prepareLexer("0", []string{}, nil)

				ok, _, err := lexer.AcceptFn(func(r rune) bool {
					return strings.ContainsRune("01234", r)
				})
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 2, Offset: 1}, lexer.Pos())
			})

			t.Run("does not accept a character for which the callback returns false.", func(t *testing.T) {
				lexer, filename := prepareLexer("7", []string{}, nil)

				ok, _, err := lexer.AcceptFn(func(r rune) bool {
					return strings.ContainsRune("01234", r)
				})
				assert.NoError(t, err)
				assert.False(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})
		})

		t.Run("AcceptRun", func(t *testing.T) {
			t.Run("accepts any number of characters from the valid set and advances the position.", func(t *testing.T) {
				lexer, filename := prepareLexer("0000001111", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptRun("0")
				assert.NoError(t, err)
				assert.True(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 7, Offset: 6}, lexer.Pos())
			})

			t.Run("does nothing if the next rune is not in the set of valid characters.", func(t *testing.T) {
				lexer, filename := prepareLexer("0000001111", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptRun("5")
				assert.NoError(t, err)
				assert.False(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptRun call.", func(t *testing.T) {
				lexer, filename := prepareLexer("0000001111", []string{}, nil)

				didConsumeRunes, backup, err := lexer.AcceptRun("0")
				assert.NoError(t, err)
				assert.True(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 7, Offset: 6}, lexer.Pos())
				backup()
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("ends the run and does not return an error when encountering EOF.", func(t *testing.T) {
				lexer, _ := prepareLexer("", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptRun("0")
				assert.NoError(t, err)
				assert.False(t, didConsumeRunes)
			})
		})

		t.Run("AcceptFnRun", func(t *testing.T) {
			t.Run("accepts a run of characters for which the callback returns true.", func(t *testing.T) {
				lexer, filename := prepareLexer("0123456789", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptRunFn(func(r rune) bool {
					return strings.ContainsRune("01234", r)
				})
				assert.NoError(t, err)
				assert.True(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 6, Offset: 5}, lexer.Pos())
			})

			t.Run("does nothing if the next rune does not fulfil the predicate.", func(t *testing.T) {
				lexer, filename := prepareLexer("555", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptRunFn(func(r rune) bool {
					return strings.ContainsRune("01234", r)
				})
				assert.NoError(t, err)
				assert.False(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptRunFn call.", func(t *testing.T) {
				lexer, filename := prepareLexer("0123456789", []string{}, nil)

				didConsumeRunes, backup, err := lexer.AcceptRunFn(func(r rune) bool {
					return strings.ContainsRune("01234", r)
				})
				assert.NoError(t, err)
				assert.True(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 6, Offset: 5}, lexer.Pos())
				backup()
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("ends the run and does not return an error when encountering EOF.", func(t *testing.T) {
				lexer, _ := prepareLexer("", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptRunFn(func(r rune) bool {
					return strings.ContainsRune("01234", r)
				})
				assert.NoError(t, err)
				assert.False(t, didConsumeRunes)
			})
		})

		t.Run("AcceptString", func(t *testing.T) {
			t.Run("accepts the full given string and advances the position.", func(t *testing.T) {
				lexer, filename := prepareLexer("test input", []string{}, nil)

				ok, _, err := lexer.AcceptString("test")
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 5, Offset: 4}, lexer.Pos())
			})

			t.Run("returns false and stays at position if the next characters do not match.", func(t *testing.T) {
				lexer, filename := prepareLexer("test input", []string{}, nil)

				ok, _, err := lexer.AcceptString("foo")
				assert.NoError(t, err)
				assert.False(t, ok)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptString call.", func(t *testing.T) {
				lexer, filename := prepareLexer("test input", []string{}, nil)

				_, backup, err := lexer.AcceptString("test")
				assert.NoError(t, err)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 5, Offset: 4}, lexer.Pos())
				backup()
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns an error when encountering EOF.", func(t *testing.T) {
				lexer, _ := prepareLexer("te", []string{}, nil)

				_, _, err := lexer.AcceptString("test")
				assert.ErrorIs(t, err, lexing.ErrEof)
			})
		})

		t.Run("AcceptUntil", func(t *testing.T) {
			t.Run("accepts any number of characters until an invalid character is encountered and advances the position.", func(t *testing.T) {
				lexer, filename := prepareLexer("ngqxflguiasordsoxl0xeflq", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptUntil("0")
				assert.NoError(t, err)
				assert.True(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 19, Offset: 18}, lexer.Pos())
			})

			t.Run("does nothing if the next rune is an invalid character.", func(t *testing.T) {
				lexer, filename := prepareLexer("5000001111", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptUntil("5")
				assert.NoError(t, err)
				assert.False(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("returns a backup function that rewinds the position to before the acceptUntil call.", func(t *testing.T) {
				lexer, filename := prepareLexer("ngqxflguiasordsoxl0xeflq", []string{}, nil)

				didConsumeRunes, backup, err := lexer.AcceptUntil("0")
				assert.NoError(t, err)
				assert.True(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 19, Offset: 18}, lexer.Pos())
				backup()
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0}, lexer.Pos())
			})

			t.Run("ends the run and does not return an error when encountering EOF.", func(t *testing.T) {
				lexer, filename := prepareLexer("nlgeqxgenui", []string{}, nil)

				didConsumeRunes, _, err := lexer.AcceptUntil("0")
				assert.NoError(t, err)
				assert.True(t, didConsumeRunes)
				assert.Equal(t, participleLexer.Position{Filename: filename, Line: 1, Column: 12, Offset: 11}, lexer.Pos())
			})
		})

		t.Run("AssertAfter", func(t *testing.T) {
			t.Run("returns false if the current position is at the beginning of the input.", func(t *testing.T) {
				lexer, _ := prepareLexer("test input", []string{}, nil)

				ok := lexer.AssertAfter("abc")
				assert.False(t, ok)
			})

			t.Run("returns true if the previous rune in the input belongs to the valid set.", func(t *testing.T) {
				lexer, _ := prepareLexer("test input", []string{}, nil)

				_, _, _ = lexer.Accept("t")

				ok := lexer.AssertAfter("t")
				assert.True(t, ok)
			})

			t.Run("returns false if the previous rune in the input does not belong to the valid set.", func(t *testing.T) {
				lexer, _ := prepareLexer("test input", []string{}, nil)

				_, _, _ = lexer.Accept("t")

				ok := lexer.AssertAfter("abc")
				assert.False(t, ok)
			})

			t.Run("does not fail when the previous rune is multiple bytes long.", func(t *testing.T) {
				lexer, _ := prepareLexer("€ ", []string{}, nil)

				ok, _, err := lexer.Accept("€")
				assert.NoError(t, err)
				assert.True(t, ok)

				ok = lexer.AssertAfter("\n")
				assert.False(t, ok)
			})
		})

		t.Run("AssertAtStart", func(t *testing.T) {
			t.Run("returns true if the current position is at the beginning of the input.", func(t *testing.T) {
				lexer, _ := prepareLexer("test input", []string{}, nil)

				ok := lexer.AssertAtStart()
				assert.True(t, ok)
			})

			t.Run("returns false if the current position is not at the beginning of the input.", func(t *testing.T) {
				lexer, _ := prepareLexer("test input", []string{}, nil)

				_, _, _ = lexer.Accept("t")

				ok := lexer.AssertAtStart()
				assert.False(t, ok)
			})
		})
	})
}
