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

type journalLexer struct {
	name     string                // used only for error reports
	input    string                // the string being lexed
	start    lexer.Position        // start of the current token
	pos      lexer.Position        // current position in the input
	tokens   chan lexer.Token      // channel of lexed tokens
	backupFn func() lexer.Position // rewind position one step, if it is set
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

func (l *journalLexer) next() rune {
	if l.pos.Offset >= len(l.input) {
		return eof
	}

    l.backupFn = func() lexer.Position {
        backupPos := l.pos
        return backupPos
    }
            
	// TODO: handle potential encoding error
	rune, _ := utf8.DecodeRuneInString(l.input[l.pos.Offset:])
	l.pos.Advance(string(rune))

	return rune
}

// This can only be used once, since the backup function is dependent
// on the previously executed action. Calling backup twice is idempotent.
// Many functions change backup's behavior. E.g. it is not possible to
// call `peek()` and then `backup()` afterwards, since peek writes and
// uses backup.
func (l *journalLexer) backup() {
	if l.backupFn != nil {
		l.pos = l.backupFn()
	}
}

func (l *journalLexer) ignore() {
    l.start = l.pos
}

func (l *journalLexer) peek() rune {
    rune := l.next()
    l.backup()
    return rune
}

func (l *journalLexer) accept(valid string) bool {
    if strings.IndexRune(valid, l.next()) != -1 {
        return true
    }
    l.backup()
    return false
}


type stateFn func(*journalLexer) stateFn

func lexRoot(l *journalLexer) stateFn {
    if l.accept("\n") {
        l.emit(itemNewline)
        return lexRoot
    }
	return nil
}

func MakeLexer() lexer.StringDefinition {
	return &journalLexerDefinition{}
}
