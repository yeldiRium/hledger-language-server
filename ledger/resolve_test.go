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
		journalFilePath := "some/path/root.journal"
		includedFilePath := "some/path/to/an/include.journal"

		fs := fstest.MapFS{
			includedFilePath: &fstest.MapFile{
				Data: []byte("account assets:Checking\n"),
			},
		}
		inputJournal := &ledger.Journal{
			Entries: []ledger.Entry{
				&ledger.IncludeDirective{
					IncludePath: "to/an/include.journal",
				},
			},
		}

		parser := ledger.NewJournalParser()

		resolvedJournal, err := ledger.ResolveIncludes(inputJournal, journalFilePath, parser, fs)

		assert.NoError(t, err)
		assert.Equal(t, &ledger.Journal{
			Entries: []ledger.Entry{
				&ledger.AccountDirective{
					AccountName: &ledger.AccountName{
						Pos: participleLexer.Position{
							Filename: includedFilePath,
							Offset:   8,
							Line:     1,
							Column:   9,
						},
						EndPos: participleLexer.Position{
							Filename: includedFilePath,
							Offset:   23,
							Line:     1,
							Column:   24,
						},
						Segments: []string{"assets", "Checking"},
					},
				},
			},
		}, resolvedJournal)
	})

	t.Run("It never uses absolute paths, instead it converts paths to be relative by removing the preceding /.", func(t *testing.T) {
		journalFilePath := "some/path/root.journal"
		includedFilePath := "some/path/to/an/include.journal"

		fs := fstest.MapFS{
			includedFilePath: &fstest.MapFile{
				Data: []byte("account assets:Checking\n"),
			},
		}
		inputJournal := &ledger.Journal{
			Entries: []ledger.Entry{
				&ledger.IncludeDirective{
					IncludePath: "/some/path/to/an/include.journal",
				},
			},
		}

		parser := ledger.NewJournalParser()

		resolvedJournal, err := ledger.ResolveIncludes(inputJournal, journalFilePath, parser, fs)

		assert.NoError(t, err)
		assert.Equal(t, &ledger.Journal{
			Entries: []ledger.Entry{
				&ledger.AccountDirective{
					AccountName: &ledger.AccountName{
						Pos: participleLexer.Position{
							Filename: includedFilePath,
							Offset:   8,
							Line:     1,
							Column:   9,
						},
						EndPos: participleLexer.Position{
							Filename: includedFilePath,
							Offset:   23,
							Line:     1,
							Column:   24,
						},
						Segments: []string{"assets", "Checking"},
					},
				},
			},
		}, resolvedJournal)
	})
}
