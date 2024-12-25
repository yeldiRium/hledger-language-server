package ledger_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeldiRium/hledger-language-server/ledger"
)

func TestAccountNames(t *testing.T) {
	t.Run("returns an empty list for an empty journal.", func(t *testing.T) {
		journal := &ledger.Journal{
			Entries: []ledger.Entry{},
		}

		accountNames := ledger.AccountNames(journal)

		assert.Equal(t, []ledger.AccountName{}, accountNames)
	})

	t.Run("returns a list of all account names and their prefixes in the journal.", func(t *testing.T) {
		journal := &ledger.Journal{
			Entries: []ledger.Entry{
				&ledger.AccountDirective{
					AccountName: &ledger.AccountName{
						Segments: []string{"assets", "Cash", "Checking"},
					},
				},
				&ledger.AccountDirective{
					AccountName: &ledger.AccountName{
						Segments: []string{"revenue", "Salary"},
					},
				},
			},
		}

		accountNames := ledger.AccountNames(journal)

		assert.ElementsMatch(
			t,
			[]ledger.AccountName{
				{Segments: []string{"assets"}},
				{Segments: []string{"assets", "Cash"}},
				{Segments: []string{"assets", "Cash", "Checking"}},
				{Segments: []string{"revenue"}},
				{Segments: []string{"revenue", "Salary"}},
			},
			accountNames,
		)
	})
}

func TestFilterAccountNamesByPrefix(t *testing.T) {
	t.Run("returns an empty list if no account names are given.", func(t *testing.T) {
		accountNames := []ledger.AccountName{}

		matchingAccountNames := ledger.FilterAccountNamesByPrefix(accountNames, nil)

		assert.Equal(t, []ledger.AccountName{}, matchingAccountNames)
	})

	t.Run("returns all 1-length account names if no query is given", func(t *testing.T) {
		accountNames := []ledger.AccountName{
			{Segments: []string{"assets"}},
			{Segments: []string{"assets", "Cash"}},
			{Segments: []string{"assets", "Cash", "Checking"}},
			{Segments: []string{"assets", "Capital"}},
			{Segments: []string{"assets", "Capital", "ETFs"}},
			{Segments: []string{"revenue"}},
			{Segments: []string{"revenue", "Salary"}},
		}

		matchingAccountNames := ledger.FilterAccountNamesByPrefix(accountNames, nil)

		assert.Equal(
			t,
			[]ledger.AccountName{
				{Segments: []string{"assets"}},
				{Segments: []string{"revenue"}},
			},
			matchingAccountNames,
		)
	})

	t.Run("returns all account names matching the query that are at most one segment longer than the query.", func(t *testing.T) {
		accountNames := []ledger.AccountName{
			{Segments: []string{"assets"}},
			{Segments: []string{"assets", "Cash"}},
			{Segments: []string{"assets", "Cash", "Checking"}},
			{Segments: []string{"assets", "Capital"}},
			{Segments: []string{"assets", "Capital", "ETFs"}},
			{Segments: []string{"revenue"}},
			{Segments: []string{"revenue", "Salary"}},
		}

		type testCase struct {
			query                *ledger.AccountName
			expectedAccountNames []ledger.AccountName
		}

		testCases := []testCase{
			{
				query: &ledger.AccountName{
					Segments: []string{"ass"},
				},
				expectedAccountNames: []ledger.AccountName{
					{Segments: []string{"assets"}},
					{Segments: []string{"assets", "Cash"}},
					{Segments: []string{"assets", "Capital"}},
				},
			},
		}
		for i, testCase := range testCases {
			t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
				matchingAccountNames := ledger.FilterAccountNamesByPrefix(accountNames, testCase.query)

				assert.Equal(t, testCase.expectedAccountNames, matchingAccountNames)
			})
		}
	})
}
