package ledger

// FindAccountNameUnderCursor's parameters line and column are 1-based.
func FindAccountNameUnderCursor(journal *Journal, line, column int) *AccountName {
	for _, entry := range journal.Entries {
		switch entry := entry.(type) {
		case *AccountDirective:
			if entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		case *RealPosting:
			if entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		case *VirtualPosting:
			if entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		case *VirtualBalancedPosting:
			if entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		}
	}

	return nil
}
