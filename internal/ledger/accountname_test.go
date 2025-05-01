package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountName(t *testing.T) {
	t.Run("Prefixes", func(t *testing.T) {
		t.Run("returns an empty list if the account name has no segments.", func(t *testing.T) {
			accountName := &AccountName{
				Segments: []string{},
			}

			prefixes := accountName.Prefixes()

			assert.Equal(t, []AccountName{}, prefixes)
		})

		t.Run("returns all segment prefixes of the account name.", func(t *testing.T) {
			accountName := &AccountName{
				Segments: []string{"assets", "Cash", "Checking", "Whatever", "another layer"},
			}

			prefixes := accountName.Prefixes()

			assert.Equal(
				t,
				[]AccountName{
					{Segments: []string{"assets"}},
					{Segments: []string{"assets", "Cash"}},
					{Segments: []string{"assets", "Cash", "Checking"}},
					{Segments: []string{"assets", "Cash", "Checking", "Whatever"}},
					{Segments: []string{"assets", "Cash", "Checking", "Whatever", "another layer"}},
				},
				prefixes,
			)
		})
	})
}
