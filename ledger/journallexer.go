package ledger

func lexRoot(l *Lexer) StateFn {
	if ok, _ := l.AcceptEof(); ok {
		return nil
	}
	if ok, _, _ := l.Accept("\n"); ok {
		l.Emit(l.Symbol("Newline"))
		return lexRoot
	}
	if ok, _, err := l.AcceptString("account"); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("AccountDirective"))
		l.AcceptRun(" ")
		l.Emit(l.Symbol("Whitespace"))
		return lexAccountDirective
	}

	if didConsumeRunes, _, err := l.AcceptUntil("\n"); err != nil {
		l.Error(err)
		return nil
	} else if didConsumeRunes {
		l.Emit(l.Symbol("Garbage"))
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

		if ok, _, err := AcceptInlineCommentIndicator(l); err != nil {
			l.Error(err)
			return nil
		} else if ok {
			return lexRoot // TODO: Handle inline comment
		}

		if ok, _, err := l.Accept(":"); err != nil {
			l.Error(err)
			return nil
		} else if ok {
			l.Emit(l.Symbol("AccountNameSeparator"))
			continue
		}

		if ok, _, err := l.Accept("\n"); err != nil {
			l.Error(err)
			return nil
		} else if ok {
			l.Emit(l.Symbol("Newline"))
			return lexRoot
		}

		if didConsumeRunes, _, err := l.AcceptRunFn(func(r rune) bool {
			if r == ' ' {
				nextRune := l.Peek()
				if nextRune == ' ' {
					return false
				}
				return true
			}
			return r != '\n' && r != ':' && r != EOF
		}); err != nil {
			l.Error(err)
			return nil
		} else if !didConsumeRunes {
			l.Errorf("expected account name segment, but found nothing")
		}
		l.Emit(l.Symbol("AccountNameSegment"))
	}
}

func AcceptInlineCommentIndicator(l *Lexer) (bool, BackupFn, error) {
	// TODO: Emit lineCommentIndicator token.
	if ok, backup, err := l.AcceptString("  ;"); err != nil {
		return false, nil, err
	} else if ok {
		return true, backup, nil
	}
	return l.AcceptString("  #")
}

func MakeJournalLexer() *LexerDefinition {
	return MakeLexerDefinition(lexRoot, []string{
		"Garbage",
		"Newline",
		"Whitespace",
		"AccountDirective",
		"AccountNameSegment",
		"AccountNameSeparator",
		"Indent",
		"Amount",
	})
}
