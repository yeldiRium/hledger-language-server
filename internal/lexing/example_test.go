package lexing_test

import (
	"errors"

	"github.com/yeldiRium/hledger-language-server/internal/lexing"
)

// ExampleLexerDefinition demonstrates a very simple lexer that recognizes
// newlines and anything in between. It handles EOF errors correctly.
func ExampleLexerDefinition() {
	var rootState lexing.StateFn
	rootState = func(lexer *lexing.Lexer) lexing.StateFn {
		if ok, _, _ := lexer.AcceptUntil("\n"); ok {
			lexer.Emit(lexer.Symbol("String"))
		}

		if ok, _, err := lexer.Accept("\n"); err != nil {
			if !errors.Is(err, lexing.ErrEof) {
				panic(err)
			}
			return nil
		} else if ok {
			lexer.Emit(lexer.Symbol("Newline"))
		}

		return rootState
	}
	lexing.NewLexerDefinition(rootState, []string{"String", "Newline"})
}
