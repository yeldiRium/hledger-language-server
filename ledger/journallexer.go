package ledger

import (
	"strings"

	"github.com/yeldiRium/hledger-language-server/lexing"
)

func lexRoot(lexer *lexing.Lexer) lexing.StateFn {
	if ok, _ := lexer.AcceptEof(); ok {
		return nil
	}
	if ok, _, _ := lexer.Accept("\n"); ok {
		lexer.Emit(lexer.Symbol("Newline"))
		return lexRoot
	}
	if ok, _, err := lexer.AcceptString("account"); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("AccountDirective"))
		lexer.AcceptRun(" ")
		lexer.Emit(lexer.Symbol("Whitespace"))
		return lexAccountDirective
	}

	if lexer.AssertAfter("\n") || lexer.AssertAtStart() {
		if ok, _, err := lexer.AcceptRun(" "); err != nil {
			lexer.Error(err)
			return nil
		} else if ok {
			lexer.Emit(lexer.Symbol("Indent"))

			if ok, _, err := AcceptCommentIndicator(lexer); err != nil {
				lexer.Error(err)
				return nil
			} else if ok {
				return lexRoot // TODO: Handle comment
			}

			if ok, _, err := lexer.AcceptString("format"); err != nil {
				lexer.Error(err)
				return nil
			} else if ok {
				return lexRoot // TODO: Handle commodity directive format subdirective
			}

			return lexPosting
		}
	}

	if didConsumeRunes, _, err := lexer.AcceptUntil("\n"); err != nil {
		lexer.Error(err)
		return nil
	} else if didConsumeRunes {
		lexer.Emit(lexer.Symbol("Garbage"))
	}
	return lexRoot
}

func lexAccountDirective(lexer *lexing.Lexer) lexing.StateFn {
	if ok, _, err := AcceptAccountName(lexer); err != nil {
		lexer.Error(err)
		return nil
	} else if !ok {
		lexer.Errorf("expected account name")
		return nil
	}

	if ok, _, err := AcceptInlineCommentIndicator(lexer); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		return lexRoot // TODO: Handle inline comment
	}

	return lexRoot
}

func AcceptAccountName(lexer *lexing.Lexer) (didConsumeAccountNameSegments bool, backup lexing.BackupFn, err error) {
	backup = lexer.NewBackup()
	didConsumeAccountNameSegments = false

	for {
		if ok, _, err := lexer.Accept(":"); err != nil {
			lexer.Error(err)
			return false, nil, err
		} else if ok {
			lexer.Emit(lexer.Symbol("AccountNameSeparator"))
			continue
		}

		if didConsumeRunes, _, err := lexer.AcceptRunFn(func(r rune) bool {
			if r == ' ' {
				nextRune := lexer.Peek()
				if nextRune == ' ' {
					return false
				}
				return true
			}
			return strings.IndexRune("()[]:\n", r) == -1
		}); err != nil {
			return false, nil, err
		} else if didConsumeRunes {
			lexer.Emit(lexer.Symbol("AccountNameSegment"))
			didConsumeAccountNameSegments = true
			continue
		}

		break
	}

	return didConsumeAccountNameSegments, backup, nil
}

func lexPosting(lexer *lexing.Lexer) lexing.StateFn {
	if ok, _, err := lexer.Accept("!*"); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("PostingStatusIndicator"))
		if ok, _, _ := lexer.Accept(" "); !ok {
			lexer.Errorf("expected whitespace after posting status indicator")
		}
		lexer.Emit(lexer.Symbol("Whitespace"))
	}

	if ok, _, err := lexer.Accept("("); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("AccountNameDelimiter"))
	} else if ok, _, err := lexer.Accept("["); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("AccountNameDelimiter"))
	}

	// TODO: maybe handle whitespace before account name?
	if ok, _, err := AcceptAccountName(lexer); err != nil {
		lexer.Error(err)
		return nil
	} else if !ok {
		lexer.Errorf("expected account name")
		return nil
	}
	// TODO: maybe handle whitespace after account name?

	if ok, _, err := lexer.Accept(")"); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("AccountNameDelimiter"))
	} else if ok, _, err := lexer.Accept("]"); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("AccountNameDelimiter"))
	}

	if ok, _, err := lexer.AcceptRun(" "); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("Whitespace"))
	}

	if ok, _, err := lexer.AcceptRunFn(func(r rune) bool {
		if r == ' ' {
			nextRune := lexer.Peek()
			if nextRune == ' ' {
				return false
			}
			return true
		}
		return strings.IndexRune("()[]\n", r) == -1
	}); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		lexer.Emit(lexer.Symbol("Amount"))
	}

	if ok, _, err := AcceptInlineCommentIndicator(lexer); err != nil {
		lexer.Error(err)
		return nil
	} else if ok {
		return lexRoot // TODO: Handle inline comment
	}

	return lexRoot
}

func AcceptCommentIndicator(lexer *lexing.Lexer) (bool, lexing.BackupFn, error) {
	return lexer.Accept("#;")
}

func AcceptInlineCommentIndicator(lexer *lexing.Lexer) (bool, lexing.BackupFn, error) {
	backup := lexer.NewBackup()

	if ok, _, err := lexer.AcceptString("  ;"); err != nil {
		return false, nil, err
	} else if ok {
		lexer.Emit(lexer.Symbol("InlineCommentIndicator"))
		return true, backup, nil
	}

	if ok, _, err := lexer.AcceptString("  #"); err != nil {
		return false, nil, err
	} else if ok {
		lexer.Emit(lexer.Symbol("InlineCommentIndicator"))
		return true, backup, nil
	}

	return false, backup, nil
}

func NewJournalLexer() *lexing.LexerDefinition {
	return lexing.NewLexerDefinition(lexRoot, []string{
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
