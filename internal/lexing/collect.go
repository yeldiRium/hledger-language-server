package lexing

import (
	participleLexer "github.com/alecthomas/participle/v2/lexer"
)

// CollectLexerTokensWithIncludeExclude collects the tokens produced by a lexer
// while only collecting included symbols and discarding excluded symbols.
// If no included symbols are given, all symbols (expect for the excluded
// symbols) are collected.
func CollectLexerTokensWithIncludeExclude(
	lexer participleLexer.Lexer,
	includedSymbols map[participleLexer.TokenType]struct{},
	excludedSymbols map[participleLexer.TokenType]struct{},
) ([]participleLexer.Token, error) {
	tokens := make([]participleLexer.Token, 0)
	for token, err := lexer.Next(); token.Type != participleLexer.EOF; token, err = lexer.Next() {
		if err != nil {
			return nil, err
		}

		if len(includedSymbols) > 0 {
			if _, ok := includedSymbols[token.Type]; !ok {
				continue
			}
		}

		if _, ok := excludedSymbols[token.Type]; ok {
			continue
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func CollectAllLexerTokens(lexer participleLexer.Lexer) ([]participleLexer.Token, error) {
	return CollectLexerTokensWithIncludeExclude(
		lexer,
		map[participleLexer.TokenType]struct{}{},
		map[participleLexer.TokenType]struct{}{},
	)
}
