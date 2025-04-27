package ledger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountNames(t *testing.T) {
	t.Run("returns an empty list for an empty journal.", func(t *testing.T) {
		journal := &Journal{
			Entries: []Entry{},
		}

		accountNames := AccountNames(journal)

		assert.Equal(t, []AccountName{}, accountNames)
	})

	t.Run("returns a list of all account names and their prefixes in the journal.", func(t *testing.T) {
		journal := &Journal{
			Entries: []Entry{
				&AccountDirective{
					AccountName: &AccountName{
						Segments: []string{"assets", "Cash", "Checking"},
					},
				},
				&AccountDirective{
					AccountName: &AccountName{
						Segments: []string{"revenue", "Salary"},
					},
				},
			},
		}

		accountNames := AccountNames(journal)

		assert.ElementsMatch(
			t,
			[]AccountName{
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
		accountNames := []AccountName{}

		matchingAccountNames := FilterAccountNamesByPrefix(accountNames, nil)

		assert.Equal(t, []AccountName{}, matchingAccountNames)
	})

	t.Run("returns all 1-length account names if no query is given", func(t *testing.T) {
		accountNames := []AccountName{
			{Segments: []string{"assets"}},
			{Segments: []string{"assets", "Cash"}},
			{Segments: []string{"assets", "Cash", "Checking"}},
			{Segments: []string{"assets", "Capital"}},
			{Segments: []string{"assets", "Capital", "ETFs"}},
			{Segments: []string{"revenue"}},
			{Segments: []string{"revenue", "Salary"}},
		}

		matchingAccountNames := FilterAccountNamesByPrefix(accountNames, nil)

		assert.Equal(
			t,
			[]AccountName{
				{Segments: []string{"assets"}},
				{Segments: []string{"revenue"}},
			},
			matchingAccountNames,
		)
	})

	t.Run("returns all account names matching the query that are at most one segment longer than the query.", func(t *testing.T) {
		accountNames := []AccountName{
			{Segments: []string{"assets"}},
			{Segments: []string{"assets", "Cash"}},
			{Segments: []string{"assets", "Cash", "Checking"}},
			{Segments: []string{"assets", "Capital"}},
			{Segments: []string{"assets", "Capital", "ETFs"}},
			{Segments: []string{"revenue"}},
			{Segments: []string{"revenue", "Salary"}},
		}

		type testCase struct {
			query                *AccountName
			expectedAccountNames []AccountName
		}

		testCases := []testCase{
			{
				query: &AccountName{
					Segments: []string{"ass"},
				},
				expectedAccountNames: []AccountName{
					{Segments: []string{"assets"}},
					{Segments: []string{"assets", "Cash"}},
					{Segments: []string{"assets", "Capital"}},
				},
			},
			{
				query: &AccountName{
					Segments: []string{"a", "C", "C"},
				},
				expectedAccountNames: []AccountName{
					{Segments: []string{"assets", "Cash", "Checking"}},
				},
			},
			{
				query: &AccountName{
					Segments: []string{"a", "c", "c"},
				},
				expectedAccountNames: []AccountName{
					{Segments: []string{"assets", "Cash", "Checking"}},
				},
			},
		}
		for i, testCase := range testCases {
			t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
				matchingAccountNames := FilterAccountNamesByPrefix(accountNames, testCase.query)

				assert.Equal(t, testCase.expectedAccountNames, matchingAccountNames)
			})
		}
	})

	t.Run("returns all account names that contain segments starting with the given segments in order, but may include other segments in between.", func(t *testing.T) {
		accountNames := []AccountName{
			{Segments: []string{"assets"}},
			{Segments: []string{"assets", "Cash"}},
			{Segments: []string{"assets", "Cash", "Checking"}},
			{Segments: []string{"assets", "Cash", "Various"}},
			{Segments: []string{"assets", "Cash", "Various", "Next"}},
			{Segments: []string{"assets", "Cash", "Various", "Next", "Too Long"}},
			{Segments: []string{"assets", "Capital"}},
			{Segments: []string{"assets", "Capital", "ETFs"}},
			{Segments: []string{"assets", "Capital", "Various"}},
			{Segments: []string{"revenue"}},
			{Segments: []string{"revenue", "Salary"}},
		}

		type testCase struct {
			query                *AccountName
			expectedAccountNames []AccountName
		}

		testCases := []testCase{
			{
				query: &AccountName{
					Segments: []string{"ass", "Var"},
				},
				expectedAccountNames: []AccountName{
					{Segments: []string{"assets", "Cash", "Various"}},
					{Segments: []string{"assets", "Cash", "Various", "Next"}},
					{Segments: []string{"assets", "Capital", "Various"}},
				},
			},
		}
		for i, testCase := range testCases {
			t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
				matchingAccountNames := FilterAccountNamesByPrefix(accountNames, testCase.query)

				assert.Equal(t, testCase.expectedAccountNames, matchingAccountNames)
			})
		}
	})
}
