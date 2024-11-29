package ledger

import (
	participleLexer "github.com/alecthomas/participle/v2/lexer"
)

func MakeLexer(
	name string,
	definition *LexerDefinition,
	input string,
	start participleLexer.Position,
	pos participleLexer.Position,
	tokens chan participleLexer.Token,
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

func (lexer *Lexer) Start() participleLexer.Position {
	return lexer.start
}

func (lexer *Lexer) Pos() participleLexer.Position {
	return lexer.pos
}
