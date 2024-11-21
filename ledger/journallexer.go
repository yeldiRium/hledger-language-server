package ledger

import "github.com/alecthomas/participle/v2/lexer"

const (
	itemNewline lexer.TokenType = iota + 2
	itemWhitespace
	itemAccountDirective
	itemAccountName
)

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

func MakeJournalLexer() *journalLexerDefinition {
	return &journalLexerDefinition{
		initialState: lexRoot,
		symbols: map[string]lexer.TokenType{
			"Newline":          itemNewline,
			"Whitespace":       itemWhitespace,
			"AccountDirective": itemAccountDirective,
			"AccountName":      itemAccountName,
		},
	}
}
