package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/yeldiRium/hledger-language-server/ledger"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

func registerCompletionCapabilities(capabilities *protocol.ServerCapabilities) {
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		ResolveProvider:   false,
		TriggerCharacters: strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 :", ""),
	}
}

func (server server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	server.logger.Debug(
		"textDocument/completion",
		zap.String("documentURI", string(params.TextDocument.URI)),
	)

	fileName := params.TextDocument.URI.Filename()
	fileName = strings.TrimPrefix(fileName, "file://")

	fileContent, ok := server.cache.GetFile(params.TextDocument.URI)
	var fileReader io.Reader
	if !ok {
		// TODO: should never open from file system. always go via cache
		server.logger.Warn(
			"textDocument/hover target not found in cache, opening from file system",
			zap.String("DocumentURI", string(params.TextDocument.URI)),
		)
		fileReader, err := os.Open(fileName)
		defer fileReader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
	} else {
		server.logger.Info(
			"textDocument/hover target found in cache",
			zap.String("DocumentURI", string(params.TextDocument.URI)),
		)
		fileReader = bytes.NewBuffer([]byte(fileContent))
	}

	parser := ledger.NewJournalParser()
	journal, err := parser.Parse(fileName, fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse journal: %w", err)
	}
	// TODO: resolves need to respect the cache
	resolvedJournal, err := ledger.ResolveIncludes(journal, parser, os.DirFS(path.Dir(fileName)))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve includes: %w", err)
	}

	lineNumber := int(params.Position.Line + 1)
	columnNumber := int(params.Position.Character + 1)
	accountNameUnderCursor := ledger.FindAccountNameUnderCursor(resolvedJournal, fileName, lineNumber, columnNumber)

	accountNames := ledger.AccountNames(resolvedJournal)
	if accountNameUnderCursor != nil {
		accountNames = slices.DeleteFunc(accountNames, func(accountName ledger.AccountName) bool {
			return accountName.Equals(*accountNameUnderCursor)
		})
	}
	matchingAccountNames := ledger.FilterAccountNamesByPrefix(accountNames, accountNameUnderCursor)

	result := protocol.CompletionList{
		IsIncomplete: true,
		Items:        make([]protocol.CompletionItem, len(accountNames)),
	}

	replaceTextLine := params.Position.Line
	replaceTextCharacter := params.Position.Character
	if accountNameUnderCursor != nil {
		replaceTextLine = uint32(accountNameUnderCursor.Pos.Line - 1)
		replaceTextCharacter = uint32(accountNameUnderCursor.Pos.Column - 1)
	}

	for i, accountName := range matchingAccountNames {
		result.Items[i] = protocol.CompletionItem{
			Label: accountName.String(),
			TextEdit: &protocol.TextEdit{
				Range: protocol.Range{
					Start: protocol.Position{
						Line: replaceTextLine,
						Character: replaceTextCharacter,
					},
					End: params.Position,
				},
				NewText: accountName.String(),
			},
		}
	}

	return &result, nil
}
