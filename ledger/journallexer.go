package ledger

import "github.com/alecthomas/participle/v2/lexer"

const (
	itemNewline lexer.TokenType = iota + 2
	itemWhitespace
	itemAccountDirective
	itemAccountName
)

func lexRoot(l *Lexer) StateFn {
	if ok, _ := l.Accept("\n"); ok {
		l.Emit(itemNewline)
		return lexRoot
	}
	if ok, _ := l.AcceptString("account"); ok {
		l.Emit(itemAccountDirective)
		l.AcceptRun(" ")
		l.Emit(itemWhitespace)
		return lexAccountDirective
	}

	return nil
}

func lexAccountDirective(l *Lexer) StateFn {
	l.AcceptUntil("\n")
	if l.pos.Offset == l.start.Offset {
		l.Errorf("expected account name in account directive, but found nothing")
	}
	// TODO: parse account name segments
	l.Emit(itemAccountName)
	l.Accept("\n")
	l.Emit(itemNewline)
	return lexRoot
}

func MakeJournalLexer() *LexerDefinition {
	return &LexerDefinition{
		initialState: lexRoot,
		symbols: map[string]lexer.TokenType{
			"Newline":          itemNewline,
			"Whitespace":       itemWhitespace,
			"AccountDirective": itemAccountDirective,
			"AccountName":      itemAccountName,
		},
	}
}
