package lexing

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
)

const EOF = -1
const (
	symbolError participleLexer.TokenType = iota
	symbolEOF

	symbolThisShouldAlwaysBeLastAndIsUsedForAddingMoreSymbols
)

var ErrEof = fmt.Errorf("unexpected end of file")
var ErrBof = fmt.Errorf("unexpected beginning of file")

func extendSymbols(symbolNames []string) map[string]participleLexer.TokenType {
	symbols := map[string]participleLexer.TokenType{
		"Error":       symbolError,
		"EOF":         symbolEOF,
		"placeholder": symbolThisShouldAlwaysBeLastAndIsUsedForAddingMoreSymbols,
	}

	for k, v := range symbolNames {
		symbols[v] = participleLexer.TokenType(k + int(symbolThisShouldAlwaysBeLastAndIsUsedForAddingMoreSymbols) + 1)
	}

	return symbols
}

// NewLexerDefinition creates a new lexer definition and sets a starting state
// as well as creating a list of token types that can be emitted by derived
// lexers.
// The list of token types is created programatically so that tests and checks
// for certain types do not rely on magic numbers but explicitly defined symbol
// names.
func NewLexerDefinition(initialState StateFn, symbolNames []string) *LexerDefinition {
	definition := &LexerDefinition{
		initialState: initialState,
		symbols:      extendSymbols(symbolNames),
	}

	return definition
}

type StateFn func(*Lexer) StateFn

type LexerDefinition struct {
	initialState StateFn
	symbols      map[string]participleLexer.TokenType
}

// LexString instantiates a new lexer, which can then be used to iterate over
// the input and emit tokens.
func (lexerDefinition *LexerDefinition) LexString(filename string, input string) (*Lexer, error) {
	l := &Lexer{
		name:       filename,
		definition: lexerDefinition,
		input:      input,
		start:      participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
		pos:        participleLexer.Position{Filename: filename, Line: 1, Column: 1, Offset: 0},
		tokens:     make(chan participleLexer.Token),
	}

	go l.run(lexerDefinition.initialState)

	return l, nil
}

// Lex reads from the given reader and instantiates a new lexer, which can then
// be used to iterate over the input and emit tokens.
// Note that the input from reader is read all at once, so this won't work with
// infinite inputs or very large files.
// Part of the `Definition` Interface from participle: https://pkg.go.dev/github.com/alecthomas/participle/v2/lexer#Definition
func (lexerDefinition *LexerDefinition) Lex(filename string, r io.Reader) (participleLexer.Lexer, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return lexerDefinition.LexString(filename, string(input))
}

// Symbols gives access to all defined symbol types.
// Part of the `Definition` Interface from participle: https://pkg.go.dev/github.com/alecthomas/participle/v2/lexer#Definition
func (lexerDefinition *LexerDefinition) Symbols() map[string]participleLexer.TokenType {
	return lexerDefinition.symbols
}

func (lexerDefinition *LexerDefinition) Symbol(name string) participleLexer.TokenType {
	if t, ok := lexerDefinition.symbols[name]; ok {
		return t
	}
	panic(fmt.Sprintf("unknown lexer token type: %q", name))
}

type BackupFn func()
type Lexer struct {
	name       string // used only for error reports
	definition *LexerDefinition
	input      string                     // the string being lexed
	start      participleLexer.Position   // start of the current token
	pos        participleLexer.Position   // current position in the input
	tokens     chan participleLexer.Token // channel of lexed tokens
}

// NewLexer creates a new lexer instance by hand.
// You should never have to use this except when testing your own Accept... or
// Assert... functions.
func NewLexer(
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

func (lexer *Lexer) run(initialState StateFn) {
	for state := initialState; state != nil; {
		state = state(lexer)
	}
	close(lexer.tokens)
}

func (lexer *Lexer) Start() participleLexer.Position {
	return lexer.start
}

func (lexer *Lexer) Pos() participleLexer.Position {
	return lexer.pos
}

func (lexer *Lexer) Next() (participleLexer.Token, error) {
	token, ok := <-lexer.tokens
	if !ok {
		return participleLexer.Token{
			Type: participleLexer.EOF,
		}, nil
	}

	if token.Type == symbolError {
		return token, errors.New(token.Value)
	}

	return token, nil
}

func (lexer *Lexer) Symbol(name string) participleLexer.TokenType {
	return lexer.definition.Symbol(name)
}

func (lexer *Lexer) NextRune() (rune, BackupFn) {
	backup := lexer.NewBackup()

	if lexer.pos.Offset >= len(lexer.input) {
		return EOF, backup
	}

	// TODO: handle potential encoding error
	rune, _ := utf8.DecodeRuneInString(lexer.input[lexer.pos.Offset:])
	lexer.pos.Advance(string(rune))

	return rune, backup
}

func (lexer *Lexer) Peek() rune {
	rune, backup := lexer.NextRune()
	backup()
	return rune
}

func (lexer *Lexer) NewBackup() BackupFn {
	backupPos := lexer.pos
	return func() {
		lexer.pos = backupPos
	}
}

func (lexer *Lexer) Emit(t participleLexer.TokenType) {
	lexer.tokens <- participleLexer.Token{
		Type:  t,
		Value: lexer.input[lexer.start.Offset:lexer.pos.Offset],
		Pos:   lexer.start,
	}
	lexer.start = lexer.pos
}

func (lexer *Lexer) Ignore() {
	lexer.start = lexer.pos
}

func (lexer *Lexer) Error(err error) StateFn {
	lexer.tokens <- participleLexer.Token{
		Type:  symbolError,
		Value: err.Error(),
		Pos:   lexer.start,
	}
	return nil
}

func (lexer *Lexer) Errorf(format string, args ...interface{}) StateFn {
	lexer.tokens <- participleLexer.Token{
		Type:  symbolError,
		Value: fmt.Sprintf(format, args...),
		Pos:   lexer.start,
	}
	return nil
}

func (lexer *Lexer) AcceptEof() (bool, BackupFn) {
	rune, backup := lexer.NextRune()
	if rune == EOF {
		return true, backup
	}
	backup()
	return false, backup
}

func (lexer *Lexer) Accept(valid string) (bool, BackupFn, error) {
	rune, backup := lexer.NextRune()
	if rune == EOF {
		return false, nil, ErrEof
	}
	if strings.ContainsRune(valid, rune) {
		return true, backup, nil
	}
	backup()
	return false, backup, nil
}

func (lexer *Lexer) AcceptString(valid string) (bool, BackupFn, error) {
	backup := lexer.NewBackup()
	for _, r := range valid {
		if ok, _, err := lexer.Accept(string(r)); err != nil {
			return false, nil, err
		} else if !ok {
			backup()
			return false, backup, nil
		}
	}
	return true, backup, nil
}

func (lexer *Lexer) AcceptFn(valid func(rune) bool) (bool, BackupFn, error) {
	rune, backup := lexer.NextRune()
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
func (lexer *Lexer) AcceptRun(valid string) (didConsumeRunes bool, backup BackupFn, err error) {
	backup = lexer.NewBackup()

	didConsumeRunes = false
	for {
		rune, backupOnce := lexer.NextRune()
		if !strings.ContainsRune(valid, rune) {
			backupOnce()
			break
		}
		didConsumeRunes = true
	}
	return didConsumeRunes, backup, nil
}

// AcceptFnRun does not return an error when encountering EOF.
// Instead it just ends the run so that an EOF may be accepted afterwards.
func (lexer *Lexer) AcceptRunFn(predicate func(rune) bool) (didConsumeRunes bool, backup BackupFn, err error) {
	backup = lexer.NewBackup()

	didConsumeRunes = false
	for {
		rune, backupOnce := lexer.NextRune()
		if !predicate(rune) {
			backupOnce()
			break
		}
		didConsumeRunes = true
	}
	return didConsumeRunes, backup, nil
}

func (lexer *Lexer) AcceptUntil(invalid string) (didConsumeRunes bool, backup BackupFn, err error) {
	backup = lexer.NewBackup()

	didConsumeRunes = false
	for {
		rune, backupOnce := lexer.NextRune()
		if strings.ContainsRune(invalid, rune) || rune == EOF {
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
func (lexer *Lexer) PeekBackwards(fromOffset int) (rune, int, error) {
	offset := fromOffset
	rune := rune(-1)
	iterations := 0
	for {
		offset -= 1
		iterations += 1
		if offset < 0 {
			return -1, -1, ErrBof
		}
		if iterations > 4 {
			panic("tried to parse unicode rune for more than four bytes, this should never happen")
		}

		rune, _ = utf8.DecodeRuneInString(lexer.input[offset:])
		if rune != utf8.RuneError {
			break
		}
	}

	return rune, offset, nil
}

func (lexer *Lexer) AssertAfter(valid string) bool {
	if lexer.pos.Offset <= 0 {
		return false
	}
	rune, _, err := lexer.PeekBackwards(lexer.pos.Offset)
	if err != nil {
		panic(err)
	}

	if strings.ContainsRune(valid, rune) {
		return true
	}
	return false
}

func (lexer *Lexer) AssertAtStart() bool {
	return lexer.pos.Offset == 0
}
