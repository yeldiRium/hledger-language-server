package parsercache_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/alecthomas/assert/v2"
	participleLexer "github.com/alecthomas/participle/v2/lexer"

	"github.com/yeldiRium/hledger-language-server/documentcache"
	"github.com/yeldiRium/hledger-language-server/ledger"
	"github.com/yeldiRium/hledger-language-server/parsercache"
)

func pruneMetadataFromAst(ast *ledger.Journal) {
	for _, entry := range ast.Entries {
		switch entry := entry.(type) {
		case *ledger.AccountDirective:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *ledger.RealPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *ledger.VirtualPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *ledger.VirtualBalancedPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		}
	}
}

func TestParserCache(t *testing.T) {
	t.Run("NewCache", func(t *testing.T) {
		t.Run("creates an empty cache.", func(t *testing.T) {
			documentCache := documentcache.NewCache(fstest.MapFS{})
			cache := parsercache.NewCache(documentCache)
			cacheSize := cache.Size()

			assert.Equal(t, 0, cacheSize)
		})
	})

	t.Run("Parse", func(t *testing.T) {
		t.Run("parses a journal, if it wasn't cached before.", func(t *testing.T) {
			documentCache := documentcache.NewCache(fstest.MapFS{
				"tmp/foo/bar.journal": &fstest.MapFile{
					Data: []byte("account assets:Cash:Checking\n"),
				},
			})
			cache := parsercache.NewCache(documentCache)

			ast, err := cache.Parse(context.Background(), "tmp/foo/bar.journal")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: &ledger.AccountName{
							Segments: []string{"assets", "Cash", "Checking"},
						},
					},
				},
			}, ast)
		})

		t.Run("returns an error if parsing a journal fails.", func(t *testing.T) {
			documentCache := documentcache.NewCache(fstest.MapFS{
				"tmp/foo/bar.journal": &fstest.MapFile{
					Data: []byte("account\n"),
				},
			})
			cache := parsercache.NewCache(documentCache)

			_, err := cache.Parse(context.Background(), "tmp/foo/bar.journal")

			assert.Error(t, err)
		})

		t.Run("takes the AST from the cache, if there is an entry for the file path. In this case, the file is not retrieved from the document cache.", func(t *testing.T) {
			documentCache := documentcache.NewCache(fstest.MapFS{
				"tmp/foo/bar.journal": &fstest.MapFile{
					Data: []byte("account assets:Cash:Checking\n"),
				},
			})
			cache := parsercache.NewCache(documentCache)

			_, err := cache.Parse(context.Background(), "tmp/foo/bar.journal")
			assert.NoError(t, err)

			documentCache.SetFile("tmp/foo/bar.journal", "account\n")

			_, err = cache.Parse(context.Background(), "tmp/foo/bar.journal")
			// If this Parse call had tried to parse the file from the document cache
			// again, it would have failed, since that document now contains an
			// invalid ledger format.
			assert.NoError(t, err)
		})

		t.Run("also caches errors.", func(t *testing.T) {
			documentCache := documentcache.NewCache(fstest.MapFS{
				"tmp/foo/bar.journal": &fstest.MapFile{
					Data: []byte("account\n"),
				},
			})
			cache := parsercache.NewCache(documentCache)

			_, err := cache.Parse(context.Background(), "tmp/foo/bar.journal")
			assert.Error(t, err)

			documentCache.SetFile("tmp/foo/bar.journal", "account assets:Cash:Checking\n")

			_, err = cache.Parse(context.Background(), "tmp/foo/bar.journal")
			assert.Error(t, err)
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("removes an entry from the cache and makes the next Parse call parse the journal again.", func(t *testing.T) {
			documentCache := documentcache.NewCache(fstest.MapFS{
				"tmp/foo/bar.journal": &fstest.MapFile{
					Data: []byte("account assets:Cash:Checking\n"),
				},
			})
			cache := parsercache.NewCache(documentCache)

			ast, err := cache.Parse(context.Background(), "tmp/foo/bar.journal")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: &ledger.AccountName{
							Segments: []string{"assets", "Cash", "Checking"},
						},
					},
				},
			}, ast)

			cache.Remove("tmp/foo/bar.journal")
			documentCache.SetFile("tmp/foo/bar.journal", "account assets:Cash:Something Else\n")

			ast, err = cache.Parse(context.Background(), "tmp/foo/bar.journal")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: &ledger.AccountName{
							Segments: []string{"assets", "Cash", "Something Else"},
						},
					},
				},
			}, ast)
		})
	})

	t.Run("ResolveIncludes", func(t *testing.T) {
		t.Run("it resolves include directives and replaces them in the journal with their parsed content.", func(t *testing.T) {
			journalFilePath := "some/path/root.journal"
			includedFilePath := "some/path/to/an/include.journal"

			fs := fstest.MapFS{
				includedFilePath: &fstest.MapFile{
					Data: []byte("account assets:Checking\n"),
				},
			}
			documentCache := documentcache.NewCache(fs)
			cache := parsercache.NewCache(documentCache)
			inputJournal := &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.IncludeDirective{
						IncludePath: "to/an/include.journal",
					},
				},
			}

			resolvedJournal, err := cache.ResolveIncludes(context.Background(), inputJournal, journalFilePath)

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

		t.Run("it never uses absolute paths, instead it converts paths to be relative by removing the preceding /.", func(t *testing.T) {
			journalFilePath := "some/path/root.journal"
			includedFilePath := "some/path/to/an/include.journal"

			fs := fstest.MapFS{
				includedFilePath: &fstest.MapFile{
					Data: []byte("account assets:Checking\n"),
				},
			}
			documentCache := documentcache.NewCache(fs)
			cache := parsercache.NewCache(documentCache)
			inputJournal := &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.IncludeDirective{
						IncludePath: "/some/path/to/an/include.journal",
					},
				},
			}

			resolvedJournal, err := cache.ResolveIncludes(context.Background(), inputJournal, journalFilePath)

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
	})
}
