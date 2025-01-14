package server

import (
	"context"
	"fmt"
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

	filePath := getFilePathFromURI(params.TextDocument.URI)

	journal, err := server.parserCache.Parse(filePath)
	if err != nil {
		server.logger.Warn("textDocument/completion failed to open/parse a journal",
			zap.String("filePath", filePath),
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse journal: %w", err)
	}

	resolvedJournal, err := server.parserCache.ResolveIncludes(journal, filePath)
	if err != nil {
		server.logger.Error("textDocument/completion failed to resolve includes",
			zap.Error(err))
		return nil, fmt.Errorf("failed to resolve includes: %w", err)
	}

	lineNumber := int(params.Position.Line + 1)
	columnNumber := int(params.Position.Character + 1)
	accountNameUnderCursor := ledger.FindAccountNameUnderCursor(resolvedJournal, filePath, lineNumber, columnNumber)

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
						Line:      replaceTextLine,
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
