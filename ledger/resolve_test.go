package ledger_test

import (
	"testing"
	"testing/fstest"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func TestResolveIncludes(t *testing.T) {
	t.Run("It resolves include directives and replaces them in the journal with their parsed content.", func(t *testing.T) {
		includePath := "some/path/to/a.journal"
		fs := fstest.MapFS{
			includePath: &fstest.MapFile{
				Data: []byte("account assets:Checking\n"),
			},
		}
		inputJournal := &ledger.Journal{
			Entries: []ledger.Entry{
				&ledger.IncludeDirective{
					IncludePath: includePath,
				},
			},
		}

		parser := ledger.NewJournalParser()

		resolvedJournal, err := ledger.ResolveIncludes(inputJournal, parser, fs)

		assert.NoError(t, err)
		assert.Equal(t, &ledger.Journal{
			Entries: []ledger.Entry{
				&ledger.AccountDirective{
					AccountName: &ledger.AccountName{
						Pos: participleLexer.Position{
							Filename: includePath,
							Offset: 8,
							Line: 1,
							Column: 9,
						},
						EndPos: participleLexer.Position{
							Filename: includePath,
							Offset: 23,
							Line: 1,
							Column: 24,
						},
						Segments: []string{"assets", "Checking"},
					},
				},
			},
		}, resolvedJournal)
	})
}
