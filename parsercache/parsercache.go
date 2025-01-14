package parsercache

import (
	"fmt"
	"path"
	"sync"

	"github.com/yeldiRium/hledger-language-server/documentcache"
	"github.com/yeldiRium/hledger-language-server/ledger"
)

type ParsingResult struct {
	journal *ledger.Journal
	err error
}

type ParserCache struct {
	sync.RWMutex
	documentCache *documentcache.DocumentCache
	asts map[string]ParsingResult
	parser *ledger.JournalParser
}

func NewCache(documentCache *documentcache.DocumentCache) *ParserCache {
	return &ParserCache{
		asts: make(map[string]ParsingResult),
		parser: ledger.NewJournalParser(),
		documentCache: documentCache,
	}
}

func (cache *ParserCache) Size() int {
	return len(cache.asts)
}

func (cache *ParserCache) Parse(filePath string) (*ledger.Journal, error) {
	cache.RLock()
	ast, ok := cache.asts[filePath]
	cache.RUnlock()

	if ok {
		return ast.journal, ast.err
	}

	file, err := cache.documentCache.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}

	journal, err := cache.parser.Parse(filePath, file)

	cache.Lock()
	defer cache.Unlock()

	cache.asts[filePath] = ParsingResult{
		journal,
		err,
	}

	return journal, err
}

func (cache *ParserCache) Remove(filePath string) {
	delete(cache.asts, filePath)
}

func (cache *ParserCache) ResolveIncludes(journal *ledger.Journal, journalFilePath string) (*ledger.Journal, error) {
	newJournal := ledger.Journal{
		Entries: make([]ledger.Entry, 0),
	}
	journalDir := path.Dir(journalFilePath)

	for _, entry := range journal.Entries {
		switch entry := entry.(type) {
		case *ledger.IncludeDirective:
			includePath := entry.IncludePath
			if !path.IsAbs(includePath) {
				includePath = path.Join(journalDir, includePath)
			}
			if path.IsAbs(includePath) {
				includePath = includePath[1:]
			}

			includeJournal, err := cache.Parse(includePath)
			if err != nil {
				return nil, err
			}
			resolvedIncludeJournal, err := cache.ResolveIncludes(includeJournal, journalFilePath)
			if err != nil {
				return nil, err
			}

			for _, entry := range resolvedIncludeJournal.Entries {
				newJournal.Entries = append(newJournal.Entries, entry)
			}
		default:
			newJournal.Entries = append(newJournal.Entries, entry)
		}
	}

	return &newJournal, nil
}
