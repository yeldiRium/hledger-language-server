package ledger

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"
)

const EOF = -1
const (
	symbolError lexer.TokenType = iota
	symbolEOF

	symbolThisShouldAlwaysBeLastAndIsUsedForAddingMoreSymbols
)

var ErrEof = fmt.Errorf("unexpected end of file")
var ErrBof = fmt.Errorf("unexpected beginning of file")

func extendSymbols(symbolNames []string) map[string]lexer.TokenType {
	symbols := map[string]lexer.TokenType{
		"Error":       symbolError,
		"EOF":         symbolEOF,
		"placeholder": symbolThisShouldAlwaysBeLastAndIsUsedForAddingMoreSymbols,
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

type BackupFn func()
type Lexer struct {
	name       string // used only for error reports
	definition *LexerDefinition
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

func (l *Lexer) NextRune() (rune, BackupFn) {
	backup := l.MakeBackup()

	if l.pos.Offset >= len(l.input) {
		return EOF, backup
	}

	// TODO: handle potential encoding error
	rune, _ := utf8.DecodeRuneInString(l.input[l.pos.Offset:])
	l.pos.Advance(string(rune))

	return rune, backup
}

func (l *Lexer) Peek() rune {
	rune, backup := l.NextRune()
	backup()
	return rune
}

func (l *Lexer) MakeBackup() BackupFn {
	backupPos := l.pos
	return func() {
		l.pos = backupPos
	}
}

func (l *Lexer) Emit(t lexer.TokenType) {
	l.tokens <- lexer.Token{
		Type:  t,
		Value: l.input[l.start.Offset:l.pos.Offset],
		Pos:   l.start,
	}
	l.start = l.pos
}

func (l *Lexer) Ignore() {
	l.start = l.pos
}

func (l *Lexer) Error(err error) StateFn {
	l.tokens <- lexer.Token{
		Type:  symbolError,
		Value: err.Error(),
		Pos:   l.start,
	}
	return nil
}

func (l *Lexer) Errorf(format string, args ...interface{}) StateFn {
	l.tokens <- lexer.Token{
		Type:  symbolError,
		Value: fmt.Sprintf(format, args...),
		Pos:   l.start,
	}
	return nil
}

func (l *Lexer) AcceptEof() (bool, BackupFn) {
	rune, backup := l.NextRune()
	if rune == EOF {
		return true, backup
	}
	backup()
	return false, backup
}

func (l *Lexer) Accept(valid string) (bool, BackupFn, error) {
	rune, backup := l.NextRune()
	if rune == EOF {
		return false, nil, ErrEof
	}
	if strings.IndexRune(valid, rune) != -1 {
		return true, backup, nil
	}
	backup()
	return false, backup, nil
}

func (l *Lexer) AcceptString(valid string) (bool, BackupFn, error) {
	backup := l.MakeBackup()
	for _, r := range valid {
		if ok, _, err := l.Accept(string(r)); err != nil {
			return false, nil, err
		} else if !ok {
			backup()
			return false, backup, nil
		}
	}
	return true, backup, nil
}

func (l *Lexer) AcceptFn(valid func(rune) bool) (bool, BackupFn, error) {
	rune, backup := l.NextRune()
	if rune == EOF {
		return false, nil, ErrEof
	}
	if valid(rune) {
		return true, backup, nil
	}
	backup()
	return false, backup, nil
}

// AcceptRun does not return an error when encountering EOF.
// Instead it just ends the run so that an EOF may be accepted afterwards.
func (l *Lexer) AcceptRun(valid string) (didConsumeRunes bool, backup BackupFn, err error) {
	backup = l.MakeBackup()

	didConsumeRunes = false
	for {
		rune, backupOnce := l.NextRune()
		if strings.IndexRune(valid, rune) == -1 {
			backupOnce()
			break
		}
		didConsumeRunes = true
	}
	return didConsumeRunes, backup, nil
}

// AcceptFnRun does not return an error when encountering EOF.
// Instead it just ends the run so that an EOF may be accepted afterwards.
func (l *Lexer) AcceptRunFn(predicate func(rune) bool) (didConsumeRunes bool, backup BackupFn, err error) {
	backup = l.MakeBackup()

	didConsumeRunes = false
	for {
		rune, backupOnce := l.NextRune()
		if !predicate(rune) {
			backupOnce()
			break
		}
		didConsumeRunes = true
	}
	return didConsumeRunes, backup, nil
}

func (l *Lexer) AcceptUntil(invalid string) (didConsumeRunes bool, backup BackupFn, err error) {
	backup = l.MakeBackup()

	didConsumeRunes = false
	for {
		rune, backupOnce := l.NextRune()
		if strings.IndexRune(invalid, rune) != -1 || rune == EOF {
			backupOnce()
			break
		}
		didConsumeRunes = true
	}

	return didConsumeRunes, backup, nil
}

// PeekBackwards assumes that the current position in the input
// was reached by successfully parsing zero or more runes and
// thus there is no invalid utf-8 in the input up until the
// given offset.
// Passing unchecked input may break this and result in a panic.
func (l *Lexer) PeekBackwards(fromOffset int) (rune, int, error) {
	offset := fromOffset - 1
	rune := rune(-1)
	iterations := 0
	for {
		if offset < 0 {
			return -1, -1, ErrBof
		}
		iterations += 1
		if iterations > 4 {
			panic("tried to parse unicode rune for more than four bytes, this should never happen")
		}

		rune, _ = utf8.DecodeLastRuneInString(l.input[offset:])
		if rune != utf8.RuneError {
			break
		}
	}

	return rune, offset, nil
}

func (l *Lexer) AssertAfter(valid string) bool {
	if l.pos.Offset <= 0 {
		return false
	}
	rune, _, err := l.PeekBackwards(l.pos.Offset)
	if err != nil {
		panic(err)
	}

	if strings.IndexRune(valid, rune) != -1 {
		return true
	}
	return false
}
