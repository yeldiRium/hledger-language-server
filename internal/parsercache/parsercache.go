package parsercache

import (
	"context"
	"fmt"
	"path"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	"github.com/yeldiRium/hledger-language-server/internal/documentcache"
	"github.com/yeldiRium/hledger-language-server/internal/ledger"
	"github.com/yeldiRium/hledger-language-server/internal/telemetry"
)

type ParsingResult struct {
	journal *ledger.Journal
	err     error
}

type ParserCache struct {
	sync.RWMutex
	documentCache *documentcache.DocumentCache
	asts          map[string]ParsingResult
	parser        *ledger.JournalParser
}

func NewCache(documentCache *documentcache.DocumentCache) *ParserCache {
	return &ParserCache{
		asts:          make(map[string]ParsingResult),
		parser:        ledger.NewJournalParser(),
		documentCache: documentCache,
	}
}

func (cache *ParserCache) Size() int {
	return len(cache.asts)
}

func (cache *ParserCache) Parse(ctx context.Context, filePath string) (*ledger.Journal, error) {
	tracer := telemetry.TracerFromContext(ctx)
	ctx, span := tracer.Start(ctx, "parsercache/parse")
	defer span.End()

	span.SetAttributes(
		attribute.String("parsercache.filePath", filePath),
	)

	cache.RLock()
	ast, ok := cache.asts[filePath]
	cache.RUnlock()

	span.SetAttributes(
		attribute.Bool("parsercache.hit", ok),
	)
	if ok {
		return ast.journal, ast.err
	}

	file, err := cache.documentCache.Open(ctx, filePath)
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

func (cache *ParserCache) ResolveIncludes(ctx context.Context, journal *ledger.Journal, journalFilePath string) (*ledger.Journal, error) {
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

			includeJournal, err := cache.Parse(ctx, includePath)
			if err != nil {
				return nil, err
			}
			resolvedIncludeJournal, err := cache.ResolveIncludes(ctx, includeJournal, journalFilePath)
			if err != nil {
				return nil, err
			}

			newJournal.Entries = append(newJournal.Entries, resolvedIncludeJournal.Entries...)
		default:
			newJournal.Entries = append(newJournal.Entries, entry)
		}
	}

	return &newJournal, nil
}
