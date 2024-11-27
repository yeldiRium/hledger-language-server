package ledger

func lexRoot(l *Lexer) StateFn {
	if ok, _ := l.AcceptEof(); ok {
		return nil
	}
	if ok, _ := l.Accept("\n"); ok {
		l.Emit(l.Symbol("Newline"))
		return lexRoot
	}
	if ok, _ := l.AcceptString("account"); ok {
		l.Emit(l.Symbol("AccountDirective"))
		l.AcceptRun(" ")
		l.Emit(l.Symbol("Whitespace"))
		return lexAccountDirective
	}

	l.AcceptUntil("\n")
	if l.pos.Offset != l.start.Offset {
		l.Ignore()
	}
	return lexRoot
}

func lexAccountDirective(l *Lexer) StateFn {
	for {
		ok, _ := l.AcceptEof()
		if ok {
			l.Errorf("unexpected EOF")
			return nil
		}

		ok, _ = AcceptInlineCommentIndicator(l)
		if ok {
			return lexRoot // TODO: Handle inline comment
		}

		ok, _ = l.Accept(":")
		if ok {
			l.Emit(l.Symbol("AccountNameSeparator"))
			continue
		}

		ok, _ = l.Accept("\n")
		if ok {
			l.Emit(l.Symbol("Newline"))
			return lexRoot
		}

		l.AcceptRunFn(func(r rune) bool {
			if r == ' ' {
				nextRune := l.Peek()
				if nextRune == ' ' {
					return false
				}
				return true
			}
			return r != '\n' && r != ':' && r != EOF
		})
		if l.pos.Offset == l.start.Offset {
			l.Errorf("expected account name segment, but found nothing")
		}
		l.Emit(l.Symbol("AccountNameSegment"))
	}
}

func AcceptInlineCommentIndicator(l *Lexer) (bool, BackupFn) {
	ok, backup := l.AcceptString("  ;")
	if ok {
		return true, backup
	}
	ok, backup = l.AcceptString("  #")
	return ok, backup
}

func MakeJournalLexer() *LexerDefinition {
	return MakeLexerDefinition(lexRoot, []string{
		"Newline",
		"Whitespace",
		"AccountDirective",
		"AccountNameSegment",
		"AccountNameSeparator",
	})
}
