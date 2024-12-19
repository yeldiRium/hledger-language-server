package ledger

// FindAccountNameUnderCursor's parameters line and column are 1-based.
func FindAccountNameUnderCursor(journal *Journal, fileName string, line, column int) *AccountName {
	for _, entry := range journal.Entries {
		switch entry := entry.(type) {
		case *AccountDirective:
			if entry.AccountName.Pos.Filename == fileName && entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		case *RealPosting:
			if entry.AccountName.Pos.Filename == fileName && entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		case *VirtualPosting:
			if entry.AccountName.Pos.Filename == fileName && entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		case *VirtualBalancedPosting:
			if entry.AccountName.Pos.Filename == fileName && entry.AccountName.Pos.Line == line && entry.AccountName.Pos.Column <= column && entry.AccountName.EndPos.Column >= column {
				return entry.AccountName
			}
		}
	}

	return nil
}

func AccountNames(journal *Journal) []string {
	accountNameSet := make(map[string]interface{})

	for _, entry := range journal.Entries {
		switch entry := entry.(type) {
		case *AccountDirective:
			if entry.AccountName == nil {
				continue
			}
			accountNameSet[entry.AccountName.String()] = struct{}{}
		case *RealPosting:
			if entry.AccountName == nil {
				continue
			}
			accountNameSet[entry.AccountName.String()] = struct{}{}
		case *VirtualPosting:
			if entry.AccountName == nil {
				continue
			}
			accountNameSet[entry.AccountName.String()] = struct{}{}
		case *VirtualBalancedPosting:
			if entry.AccountName == nil {
				continue
			}
			accountNameSet[entry.AccountName.String()] = struct{}{}
		default:
			continue
		}
	}

	accountNames := make([]string, 0, len(accountNameSet))
	for accountName := range accountNameSet {
		accountNames = append(accountNames, accountName)
	}

	return accountNames
}
