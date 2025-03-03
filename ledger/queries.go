package ledger

import "strings"

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

func AccountNames(journal *Journal) []AccountName {
	accountNameSet := make(map[string]AccountName)

	for _, entry := range journal.Entries {
		var accountName *AccountName

		switch entry := entry.(type) {
		case *AccountDirective:
			if entry.AccountName == nil {
				continue
			}
			accountName = entry.AccountName
		case *RealPosting:
			if entry.AccountName == nil {
				continue
			}
			accountName = entry.AccountName
		case *VirtualPosting:
			if entry.AccountName == nil {
				continue
			}
			accountName = entry.AccountName
		case *VirtualBalancedPosting:
			if entry.AccountName == nil {
				continue
			}
			accountName = entry.AccountName
		default:
			continue
		}

		prefixes := accountName.Prefixes()
		for _, prefix := range prefixes {
			accountNameSet[prefix.String()] = prefix
		}
	}

	accountNames := make([]AccountName, 0, len(accountNameSet))
	for _, accountName := range accountNameSet {
		accountNames = append(accountNames, accountName)
	}

	return accountNames
}

// TODO: fuzzy match segments
func FilterAccountNamesByPrefix(accountNames []AccountName, query *AccountName) []AccountName {
	matchingAccountNames := make([]AccountName, 0)
	for _, accountName := range accountNames {
		if query == nil {
			// If no query is given, we do not want to go into the matching logic at all.
			// But we do want all account names with a length of 1.
			if len(accountName.Segments) == 1 {
				matchingAccountNames = append(matchingAccountNames, accountName)
			}
			continue
		}

		queryIndex := 0
		accountNameIndex := 0
		for queryIndex < len(query.Segments) && accountNameIndex < len(accountName.Segments) {
			accountNameSegment := strings.ToLower(accountName.Segments[accountNameIndex])
			querySegment := strings.ToLower(query.Segments[queryIndex])

			if _, ok := strings.CutPrefix(accountNameSegment, querySegment); ok {
				queryIndex += 1
			}
			accountNameIndex += 1
		}

		// We only want to include account names that match all query segments.
		if queryIndex != len(query.Segments) {
			continue
		}
		// Of those matching all query segments, we only want to include the account
		// names that are at most one segment longer than the query match.
		if len(accountName.Segments) > accountNameIndex + 1 {
			continue
		}

		matchingAccountNames = append(matchingAccountNames, accountName)
	}

	return matchingAccountNames
}
