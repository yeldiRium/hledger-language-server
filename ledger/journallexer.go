package ledger

func lexRoot(l *Lexer) StateFn {
	if ok, _ := l.Accept("\n"); ok {
		l.Emit(l.Symbol("NewLine"))
		return lexRoot
	}
	if ok, _ := l.AcceptString("account"); ok {
		l.Emit(l.Symbol("AccountDirective"))
		l.AcceptRun(" ")
		l.Emit(l.Symbol("Whitespace"))
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
	l.Emit(l.Symbol("AccountName"))
	l.Accept("\n")
	l.Emit(l.Symbol("NewLine"))
	return lexRoot
}

func MakeJournalLexer() *LexerDefinition {
	return MakeLexerDefinition(lexRoot, []string {
		"Newline",
		"Whitespace",
		"AccountDirective",
		"AccountName",
	})
}
