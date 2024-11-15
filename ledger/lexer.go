package journal

import (
	"io"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

const (
	itemError lexer.TokenType = iota
	itemEOF
	itemNewline
)

type journalLexerDefinition struct{}

func (j *journalLexerDefinition) LexString(filename string, input string) (lexer.Lexer, error) {
	l := &journalLexer{
		name:   filename,
		input:  input,
		start:  lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
		pos:    lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
		width:  0,
		tokens: make(chan lexer.Token),
	}

	go l.run()

	return l, nil
}

func (j *journalLexerDefinition) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return j.LexString(filename, string(input))
}

func (j *journalLexerDefinition) Symbols() map[string]lexer.TokenType {
	return map[string]lexer.TokenType{
		"Error":	itemError,
		"EOF":	itemEOF,
		"EOL": itemNewline,
	}
}

type journalLexer struct {
	name   string           // used only for error reports
	input  string           // the string being lexed
	start  lexer.Position   // start of the current token
	pos    lexer.Position   // current position in the input
	width  int              // width of the last read rune
	tokens chan lexer.Token // channel of lexed tokens
}

func (j *journalLexer) run() {
	for state := lexRoot; state != nil; {
		state = state(j)
	}
	close(j.tokens)
}

func (j *journalLexer) Next() (lexer.Token, error) {
	token, ok := <-j.tokens
	if !ok {
		return lexer.Token{
			Type: lexer.EOF,
		}, nil
	}
	return token, nil
}

func (j *journalLexer) Emit(t lexer.TokenType) {
	j.tokens <- lexer.Token{
		Type:  t,
		Value: j.input[j.start.Offset:j.pos.Offset],
		Pos:   j.start,
	}
	j.start = j.pos
}

type stateFn func(*journalLexer) stateFn

func lexRoot(l *journalLexer) stateFn {
	if strings.HasPrefix(l.input[l.pos.Offset:], "\n") {
		l.pos.Advance("\n")
		l.Emit(itemNewline)
		return lexRoot
	}
	return nil
}

func MakeLexer() lexer.StringDefinition {
	return &journalLexerDefinition{}
}
