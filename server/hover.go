package server

import (
	"context"
	"fmt"

	"go.lsp.dev/protocol"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func registerHoverCapabilities(serverCapabilities *protocol.ServerCapabilities) {
	serverCapabilities.HoverProvider = true
}

func (server server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	span := trace.SpanFromContext(ctx)

	lineNumber := int(params.Position.Line + 1)
	columnNumber := int(params.Position.Character + 1)

	span.SetAttributes(
		attribute.String("lsp.documentURI", string(params.TextDocument.URI)),
		attribute.Int("lsp.cursorLineNumber", lineNumber),
		attribute.Int("lsp.cursorColumnNumber", columnNumber),
	)

	filePath := getFilePathFromURI(params.TextDocument.URI)

	journal, err := server.parserCache.Parse(ctx, filePath)
	if err != nil {
		err = fmt.Errorf("failed to open/parse journal: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	resolvedJournal, err := server.parserCache.ResolveIncludes(ctx, journal, filePath)
	if err != nil {
		err = fmt.Errorf("failed to resolve includes: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	accountNameUnderCursor := ledger.FindAccountNameUnderCursor(resolvedJournal, filePath, lineNumber, columnNumber)

	if accountNameUnderCursor == nil {
		span.SetAttributes(
			attribute.Bool("lsp.hover.targetFound", false),
		)
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: "Not hovering over an account name",
			},
		}, nil
	}

	span.SetAttributes(
		attribute.String("lsp.cursorElementType", "accountName"),
		attribute.String("lsp.cursorElementValue", accountNameUnderCursor.String()),
		attribute.Bool("lsp.hover.targetFound", true),
	)

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: fmt.Sprintf("You're hovering over \"%s\"", accountNameUnderCursor),
		},
	}, nil
}
