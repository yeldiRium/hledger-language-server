package ledger

import (
	"errors"
	"fmt"
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
	itemWhitespace
	itemAccountDirective
	itemAccountName
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

type backupFn func()
type journalLexer struct {
	name   string           // used only for error reports
	input  string           // the string being lexed
	start  lexer.Position   // start of the current token
	pos    lexer.Position   // current position in the input
	tokens chan lexer.Token // channel of lexed tokens
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

	if token.Type == itemError {
		return token, errors.New(token.Value)
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
	backupPos := l.pos
	return func() {
		l.pos = backupPos
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

func (l *journalLexer) acceptRun(valid string) backupFn {
	backup := l.makeBackup()
	for {
		rune, backupOnce := l.next()
		if strings.IndexRune(valid, rune) == -1 {
			backupOnce()
			break
		}
	}
	return backup
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

func (l *journalLexer) acceptUntil(invalid string) backupFn {
	backup := l.makeBackup()
	for {
		rune, backupOnce := l.next()
		if strings.IndexRune(invalid, rune) != -1 {
			backupOnce()
			break
		}
	}
	return backup
}

func (l *journalLexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- lexer.Token{
		Type:  itemError,
		Value: fmt.Sprintf(format, args...),
		Pos:   l.start,
	}
	return nil
}

type stateFn func(*journalLexer) stateFn

func lexRoot(l *journalLexer) stateFn {
	if ok, _ := l.accept("\n"); ok {
		l.emit(itemNewline)
		return lexRoot
	}
	if ok, _ := l.acceptString("account"); ok {
		l.emit(itemAccountDirective)
		l.acceptRun(" ")
		l.emit(itemWhitespace)
		return lexAccountDirective
	}

	return nil
}

func lexAccountDirective(l *journalLexer) stateFn {
	l.acceptUntil("\n")
	if l.pos.Offset == l.start.Offset {
		l.errorf("expected account name in account directive, but found nothing")
	}
	// TODO: parse account name segments
	l.emit(itemAccountName)
	l.accept("\n")
	l.emit(itemNewline)
	return lexRoot
}

func MakeLexer() *journalLexerDefinition {
	return &journalLexerDefinition{}
}
