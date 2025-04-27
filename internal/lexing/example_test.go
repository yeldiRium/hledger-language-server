package lexing

import (
	"errors"
)

// ExampleLexerDefinition demonstrates a very simple lexer that recognizes
// newlines and anything in between. It handles EOF errors correctly.
func ExampleLexerDefinition() {
	var rootState StateFn
	rootState = func(lexer *Lexer) StateFn {
		if ok, _, _ := lexer.AcceptUntil("\n"); ok {
			lexer.Emit(lexer.Symbol("String"))
		}

		if ok, _, err := lexer.Accept("\n"); err != nil {
			if !errors.Is(err, ErrEof) {
				panic(err)
			}
			return nil
		} else if ok {
			lexer.Emit(lexer.Symbol("Newline"))
		}

		return rootState
	}
	NewLexerDefinition(rootState, []string{"String", "Newline"})
}
