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
	minLength := 1
	maxLength := 1
	if query != nil {
		minLength = len(query.Segments)
		maxLength = minLength + 1
	}

	matchingAccountNames := make([]AccountName, 0)
accountNameLoop:
	for _, accountName := range accountNames {
		if len(accountName.Segments) > maxLength || len(accountName.Segments) < minLength {
			continue
		}

		if query == nil {
			matchingAccountNames = append(matchingAccountNames, accountName)
			continue
		}

		for i, segment := range query.Segments {
			if _, ok := strings.CutPrefix(accountName.Segments[i], segment); !ok {
				continue accountNameLoop
			}
		}
		matchingAccountNames = append(matchingAccountNames, accountName)
	}

	return matchingAccountNames
}
