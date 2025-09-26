package server

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"go.lsp.dev/protocol"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/yeldirium/hledger-language-server/internal/ledger"
)

func registerCompletionCapabilities(capabilities *protocol.ServerCapabilities) {
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		ResolveProvider:   false,
		TriggerCharacters: strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 :", ""),
	}
}

func (server server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	span := trace.SpanFromContext(ctx)

	lineNumber := int(params.Position.Line + 1)
	columnNumber := int(params.Position.Character + 1)

	span.SetAttributes(
		attribute.String("lsp.documentURI", string(params.TextDocument.URI)),
		attribute.Int("lsp.cursorLineNumber", lineNumber),
		attribute.Int("lsp.cursorColumnNumber", columnNumber),
	)

	filePath := getFilePathFromURI(params.TextDocument.URI)
	span.SetAttributes(
		attribute.String("lsp.documentFilePath", filePath),
	)

	journal, err := server.parserCache.Parse(ctx, filePath)
	if err != nil {
		err = fmt.Errorf("failed to open/parse journal: %w", err)
		span.RecordError(err)
		return nil, err
	}

	resolvedJournal, err := server.parserCache.ResolveIncludes(ctx, journal, filePath)
	if err != nil {
		err = fmt.Errorf("failed to resolve includes: %w", err)
		span.RecordError(err)
		return nil, err
	}

	accountNames := ledger.AccountNames(resolvedJournal)

	accountNameUnderCursor := ledger.FindAccountNameUnderCursor(resolvedJournal, filePath, lineNumber, columnNumber)
	if accountNameUnderCursor != nil {
		span.SetAttributes(
			attribute.String("lsp.cursorElementType", "accountName"),
			attribute.String("lsp.cursorElementValue", accountNameUnderCursor.String()),
		)

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

	span.SetAttributes(
		attribute.Int("lsp.completion.completionListSize", len(result.Items)),
	)

	return &result, nil
}
