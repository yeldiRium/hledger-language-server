package testing

import (
	"fmt"

	participleLexer "github.com/alecthomas/participle/v2/lexer"

	"github.com/yeldiRium/hledger-language-server/internal/lexing"
)

type MiniToken struct {
	Type  string
	Value string
}

func MakeTokens(lexer *lexing.Lexer, fileName string, miniTokens []MiniToken) []participleLexer.Token {
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

func RunLexerWithFileName(lexerDefinition *lexing.LexerDefinition, input string, fileName string) (*lexing.Lexer, []participleLexer.Token, error) {
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

//func TestLexer(lexerDefinition *LexerDefinition, input string, ...opts []TestLexerOption) {
//}
