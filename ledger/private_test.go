package ledger

import "github.com/alecthomas/participle/v2/lexer"

func MakeLexer(
	name string,
	definition *LexerDefinition,
	input string,
	start lexer.Position,
	pos lexer.Position,
	tokens chan lexer.Token,
) *Lexer {
	return &Lexer{
		name,
		definition,
		input,
		start,
		pos,
		tokens,
	}
}

func (lexer *Lexer) Start() lexer.Position {
	return lexer.start
}

func (lexer *Lexer) Pos() lexer.Position {
	return lexer.pos
}
