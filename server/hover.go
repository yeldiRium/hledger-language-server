package server

import (
	"context"
	"fmt"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func registerHoverCapabilities(serverCapabilities *protocol.ServerCapabilities) {
	serverCapabilities.HoverProvider = true
}

func (server server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	server.logger.Info(
		"textDocument/hover",
		zap.String("DocumentURI", string(params.TextDocument.URI)),
		zap.Uint32("line", params.Position.Line),
		zap.Uint32("character", params.Position.Character),
	)

	lineNumber := int(params.Position.Line + 1)
	columnNumber := int(params.Position.Character + 1)

	filePath := getFilePathFromURI(params.TextDocument.URI)

	journal, err := server.parserCache.Parse(filePath)
	if err != nil {
		server.logger.Warn("textDocument/hover failed to open/parse a journal",
			zap.String("filePath", filePath),
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse journal: %w", err)
	}

	resolvedJournal, err := server.parserCache.ResolveIncludes(journal, filePath)
	if err != nil {
		server.logger.Error("textDocument/hover failed to resolve includes",
			zap.Error(err))
		return nil, fmt.Errorf("failed to resolve includes: %w", err)
	}

	accountNameUnderCursor := ledger.FindAccountNameUnderCursor(resolvedJournal, filePath, lineNumber, columnNumber)

	if accountNameUnderCursor == nil {
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: "Not hovering over an account name",
			},
		}, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: fmt.Sprintf("You're hovering over \"%s\"", accountNameUnderCursor),
		},
	}, nil
}
