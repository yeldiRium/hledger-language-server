package ledger

import (
	"strings"
)

func (accountName *AccountName) String() string {
	return strings.Join(accountName.Segments, ":")
}

func (accountName *AccountName) Prefixes() []AccountName {
	prefixes := make([]AccountName, len(accountName.Segments))

	for i := range accountName.Segments {
		prefixes[i] = AccountName{Segments: make([]string, i+1)}
		for j := 0; j <= i; j++ {
			prefixes[i].Segments[j] = accountName.Segments[j]
		}
	}

	return prefixes
}

func (accountName AccountName) Equals(otherAccountName AccountName) bool {
	if len(accountName.Segments) != len(otherAccountName.Segments) {
		return false
	}
	for i, segment := range accountName.Segments {
		if otherAccountName.Segments[i] != segment {
			return false
		}
	}
	return true
}
