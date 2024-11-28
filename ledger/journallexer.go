package ledger

import (
	"strings"
)

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

	if l.AssertAfter("\n") || l.AssertAtStart() {
		if ok, _, err := l.AcceptRun(" "); err != nil {
			l.Error(err)
			return nil
		} else if ok {
			l.Emit(l.Symbol("Indent"))

			if ok, _, err := AcceptCommentIndicator(l); err != nil {
				l.Error(err)
				return nil
			} else if ok {
				return lexRoot // TODO: Handle comment
			}

			if ok, _, err := l.AcceptString("format"); err != nil {
				l.Error(err)
				return nil
			} else if ok {
				return lexRoot // TODO: Handle commodity directive format subdirective
			}

			return lexPosting
		}
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
	if ok, _, err := AcceptAccountName(l); err != nil {
		l.Error(err)
		return nil
	} else if !ok {
		l.Errorf("expected account name")
		return nil
	}

	if ok, _, err := AcceptInlineCommentIndicator(l); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		return lexRoot // TODO: Handle inline comment
	}

	return lexRoot
}

func AcceptAccountName(l *Lexer) (didConsumeAccountNameSegments bool, backup BackupFn, err error) {
	backup = l.MakeBackup()
	didConsumeAccountNameSegments = false

	for {
		if ok, _, err := l.Accept(":"); err != nil {
			l.Error(err)
			return false, nil, err
		} else if ok {
			l.Emit(l.Symbol("AccountNameSeparator"))
			continue
		}

		if didConsumeRunes, _, err := l.AcceptRunFn(func(r rune) bool {
			if r == ' ' {
				nextRune := l.Peek()
				if nextRune == ' ' {
					return false
				}
				return true
			}
			return strings.IndexRune("()[]:\n", r) == -1
		}); err != nil {
			return false, nil, err
		} else if didConsumeRunes {
			l.Emit(l.Symbol("AccountNameSegment"))
			didConsumeAccountNameSegments = true
			continue
		}

		break
	}

	return didConsumeAccountNameSegments, backup, nil
}

func lexPosting(l *Lexer) StateFn {
	if ok, _, err := l.Accept("!*"); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("PostingStatusIndicator"))
		if ok, _, _ := l.Accept(" "); !ok {
			l.Errorf("expected whitespace after posting status indicator")
		}
		l.Emit(l.Symbol("Whitespace"))
	}

	if ok, _, err := l.Accept("("); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("AccountNameDelimiter"))
	} else if ok, _, err := l.Accept("["); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("AccountNameDelimiter"))
	}

	// TODO: maybe handle whitespace before account name?
	if ok, _, err := AcceptAccountName(l); err != nil {
		l.Error(err)
		return nil
	} else if !ok {
		l.Errorf("expected account name")
		return nil
	}
	// TODO: maybe handle whitespace after account name?

	if ok, _, err := l.Accept(")"); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("AccountNameDelimiter"))
	} else if ok, _, err := l.Accept("]"); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("AccountNameDelimiter"))
	}

	if ok, _, err := l.AcceptRun(" "); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("Whitespace"))
	}

	if ok, _, err := l.AcceptRunFn(func(r rune) bool {
		if r == ' ' {
			nextRune := l.Peek()
			if nextRune == ' ' {
				return false
			}
			return true
		}
		return strings.IndexRune("()[]\n", r) == -1
	}); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		l.Emit(l.Symbol("Amount"))
	}

	if ok, _, err := AcceptInlineCommentIndicator(l); err != nil {
		l.Error(err)
		return nil
	} else if ok {
		return lexRoot // TODO: Handle inline comment
	}

	return lexRoot
}

func AcceptCommentIndicator(l *Lexer) (bool, BackupFn, error) {
	return l.Accept("#;")
}

func AcceptInlineCommentIndicator(l *Lexer) (bool, BackupFn, error) {
	backup := l.MakeBackup()

	if ok, _, err := l.AcceptString("  ;"); err != nil {
		return false, nil, err
	} else if ok {
		l.Emit(l.Symbol("InlineCommentIndicator"))
		return true, backup, nil
	}

	if ok, _, err := l.AcceptString("  #"); err != nil {
		return false, nil, err
	} else if ok {
		l.Emit(l.Symbol("InlineCommentIndicator"))
		return true, backup, nil
	}

	return false, backup, nil
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
		"PostingStatusIndicator",
		"AccountNameDelimiter",
		"InlineCommentIndicator",
	})
}
