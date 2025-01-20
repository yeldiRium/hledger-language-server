package server

import (
	"context"

	"go.lsp.dev/protocol"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func registerDocumentSyncCapabilities(serverCapabilities *protocol.ServerCapabilities) {
	serverCapabilities.TextDocumentSync = protocol.TextDocumentSyncKindFull
}

func (server server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("lsp.documentURI", string(params.TextDocument.URI)),
	)

	filePath := getFilePathFromURI(params.TextDocument.URI)
	span.SetAttributes(
		attribute.String("lsp.documentFilePath", filePath),
	)

	server.documentCache.SetFile(filePath, params.TextDocument.Text)
	server.parserCache.Remove(filePath)

	return nil
}

func (server server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("lsp.documentURI", string(params.TextDocument.URI)),
	)

	filePath := getFilePathFromURI(params.TextDocument.URI)
	span.SetAttributes(
		attribute.String("lsp.documentFilePath", filePath),
	)

	if len(params.ContentChanges) == 1 {
		server.documentCache.SetFile(filePath, params.ContentChanges[0].Text)
		server.parserCache.Remove(filePath)
	} else {
		span.AddEvent("unexpected amount of content changes", trace.WithAttributes(
			attribute.Int("lsp.didChange.contentChangeSize", len(params.ContentChanges)),
		))
	}

	return nil
}

func (server server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("lsp.documentURI", string(params.TextDocument.URI)),
	)

	filePath := getFilePathFromURI(params.TextDocument.URI)
	span.SetAttributes(
		attribute.String("lsp.documentFilePath", filePath),
	)

	server.documentCache.DeleteFile(filePath)

	return nil
}
