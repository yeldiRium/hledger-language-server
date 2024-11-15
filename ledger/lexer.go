package ledger

import (
	"io"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"
)

const eof = -1
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
		"Error": itemError,
		"EOF":   itemEOF,
		"EOL":   itemNewline,
	}
}

type backupFn func() lexer.Position
type journalLexer struct {
	name     string           // used only for error reports
	input    string           // the string being lexed
	start    lexer.Position   // start of the current token
	pos      lexer.Position   // current position in the input
	tokens   chan lexer.Token // channel of lexed tokens
}

func (l *journalLexer) run() {
	for state := lexRoot; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func (l *journalLexer) Next() (lexer.Token, error) {
	token, ok := <-l.tokens
	if !ok {
		return lexer.Token{
			Type: lexer.EOF,
		}, nil
	}
	return token, nil
}

func (l *journalLexer) emit(t lexer.TokenType) {
	l.tokens <- lexer.Token{
		Type:  t,
		Value: l.input[l.start.Offset:l.pos.Offset],
		Pos:   l.start,
	}
	l.start = l.pos
}

func (l *journalLexer) next() (rune, backupFn) {
    backup := l.makeBackup()

	if l.pos.Offset >= len(l.input) {
		return eof, backup
	}

	// TODO: handle potential encoding error
	rune, _ := utf8.DecodeRuneInString(l.input[l.pos.Offset:])
	l.pos.Advance(string(rune))

	return rune, backup
}

func (l *journalLexer) makeBackup() backupFn {
	return func() lexer.Position {
		backupPos := l.pos
		return backupPos
	}
}

func (l *journalLexer) ignore() {
	l.start = l.pos
}

func (l *journalLexer) peek() rune {
	rune, backup := l.next()
	backup()
	return rune
}

func (l *journalLexer) accept(valid string) (bool, backupFn) {
    rune, backup := l.next()
	if strings.IndexRune(valid, rune) != -1 {
		return true, backup
	}
    backup()
	return false, backup
}

func (l *journalLexer) acceptString(valid string) (bool, backupFn) {
    backup := l.makeBackup()
	for _, r := range valid {
        if ok, _ := l.accept(string(r)); !ok {
            backup()
			return false, backup
		}
	}
	return true, backup
}

type stateFn func(*journalLexer) stateFn

func lexRoot(l *journalLexer) stateFn {
    if ok, _ := l.accept("\n"); ok {
		l.emit(itemNewline)
		return lexRoot
	}
	return nil
}

func lexAccountDirective(l *journalLexer) stateFn {
	return nil
}

func MakeLexer() *journalLexerDefinition {
	return &journalLexerDefinition{}
}
