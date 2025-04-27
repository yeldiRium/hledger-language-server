package lexing

import (
	"fmt"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
)

type MiniToken struct {
	Type  string
	Value string
}

func MakeTokens(lexer *Lexer, fileName string, miniTokens []MiniToken) []participleLexer.Token {
	tokens := make([]participleLexer.Token, len(miniTokens))
	pos := participleLexer.Position{Filename: fileName, Offset: 0, Line: 1, Column: 1}
	for i, token := range miniTokens {
		tokens[i] = participleLexer.Token{
			Type:  lexer.Symbol(token.Type),
			Value: token.Value,
			Pos:   pos,
		}
		pos.Advance(token.Value)
	}
	return tokens
}

func RunLexerWithFileName(lexerDefinition *LexerDefinition, input string, fileName string) (*Lexer, []participleLexer.Token, error) {
	lexer, err := lexerDefinition.LexString(fileName, input)
	if err != nil {
		return nil, nil, err
	}
	tokens := make([]participleLexer.Token, 0)
	for token, err := lexer.Next(); token.Type != participleLexer.EOF; token, err = lexer.Next() {
		if err != nil {
			return nil, nil, err
		}
		tokens = append(tokens, token)
	}

	fmt.Printf("tokens: %#v\n", tokens)

	return lexer, tokens, nil
}
