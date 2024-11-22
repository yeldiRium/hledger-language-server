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
	symbolError lexer.TokenType = iota
	symbolEOF

	symbolThisShouldAlwaysBeLastAndIsUsedForAddingMoreSymbols
)

func extendSymbols(symbolNames []string) map[string]lexer.TokenType {
	symbols := map[string]lexer.TokenType{
		"Error": symbolError,
		"EOF":   symbolEOF,
	}

	for k, v := range symbolNames {
		symbols[v] = lexer.TokenType(k + int(symbolThisShouldAlwaysBeLastAndIsUsedForAddingMoreSymbols) + 1)
	}

	return symbols
}

func MakeLexerDefinition(initialState StateFn, symbolNames []string) *LexerDefinition {
	definition := &LexerDefinition{
		initialState: initialState,
		symbols:      extendSymbols(symbolNames),
	}

	return definition
}

type StateFn func(*Lexer) StateFn

type LexerDefinition struct {
	initialState StateFn
	symbols      map[string]lexer.TokenType
}

func (lexerDefinition *LexerDefinition) LexString(filename string, input string) (*Lexer, error) {
	l := &Lexer{
		name:       filename,
		definition: lexerDefinition,
		input:      input,
		start:      lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
		pos:        lexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
		tokens:     make(chan lexer.Token),
	}

	go l.run(lexerDefinition.initialState)

	return l, nil
}

func (lexerDefinition *LexerDefinition) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return lexerDefinition.LexString(filename, string(input))
}

func (lexerDefinition *LexerDefinition) Symbols() map[string]lexer.TokenType {
	return lexerDefinition.symbols
}

func (lexerDefinition *LexerDefinition) Symbol(name string) lexer.TokenType {
	if t, ok := lexerDefinition.symbols[name]; ok {
		return t
	}
	panic(fmt.Sprintf("unknown lexer token type: %q", name))
}

type backupFn func()
type Lexer struct {
	name       string // used only for error reports
	definition *LexerDefinition
	symbols    map[string]lexer.TokenType
	input      string           // the string being lexed
	start      lexer.Position   // start of the current token
	pos        lexer.Position   // current position in the input
	tokens     chan lexer.Token // channel of lexed tokens
}

func (l *Lexer) run(initialState StateFn) {
	for state := initialState; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func (l *Lexer) Next() (lexer.Token, error) {
	token, ok := <-l.tokens
	if !ok {
		return lexer.Token{
			Type: lexer.EOF,
		}, nil
	}

	if token.Type == symbolError {
		return token, errors.New(token.Value)
	}

	return token, nil
}

func (l *Lexer) Symbol(name string) lexer.TokenType {
	return l.definition.Symbol(name)
}

func (l *Lexer) Emit(t lexer.TokenType) {
	l.tokens <- lexer.Token{
		Type:  t,
		Value: l.input[l.start.Offset:l.pos.Offset],
		Pos:   l.start,
	}
	l.start = l.pos
}

func (l *Lexer) NextRune() (rune, backupFn) {
	backup := l.makeBackup()

	if l.pos.Offset >= len(l.input) {
		return eof, backup
	}

	// TODO: handle potential encoding error
	rune, _ := utf8.DecodeRuneInString(l.input[l.pos.Offset:])
	l.pos.Advance(string(rune))

	return rune, backup
}

func (l *Lexer) makeBackup() backupFn {
	backupPos := l.pos
	return func() {
		l.pos = backupPos
	}
}

func (l *Lexer) Ignore() {
	l.start = l.pos
}

func (l *Lexer) Peek() rune {
	rune, backup := l.NextRune()
	backup()
	return rune
}

func (l *Lexer) Accept(valid string) (bool, backupFn) {
	rune, backup := l.NextRune()
	if strings.IndexRune(valid, rune) != -1 {
		return true, backup
	}
	backup()
	return false, backup
}

func (l *Lexer) AcceptRun(valid string) backupFn {
	backup := l.makeBackup()
	for {
		rune, backupOnce := l.NextRune()
		if strings.IndexRune(valid, rune) == -1 {
			backupOnce()
			break
		}
	}
	return backup
}

func (l *Lexer) AcceptString(valid string) (bool, backupFn) {
	backup := l.makeBackup()
	for _, r := range valid {
		if ok, _ := l.Accept(string(r)); !ok {
			backup()
			return false, backup
		}
	}
	return true, backup
}

func (l *Lexer) AcceptUntil(invalid string) backupFn {
	backup := l.makeBackup()
	for {
		rune, backupOnce := l.NextRune()
		if strings.IndexRune(invalid, rune) != -1 {
			backupOnce()
			break
		}
	}
	return backup
}

func (l *Lexer) Errorf(format string, args ...interface{}) StateFn {
	l.tokens <- lexer.Token{
		Type:  symbolError,
		Value: fmt.Sprintf(format, args...),
		Pos:   l.start,
	}
	return nil
}
