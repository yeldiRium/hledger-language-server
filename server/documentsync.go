package server

import (
	"context"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

func registerDocumentSyncCapabilities(serverCapabilities *protocol.ServerCapabilities) {
	serverCapabilities.TextDocumentSync = protocol.TextDocumentSyncKindFull
}

func (server server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	server.logger.Info("textDocument/didOpen", zap.String("DocumentURI", string(params.TextDocument.URI)))

	server.cache.SetFile(params.TextDocument.URI, params.TextDocument.Text)

	return nil
}

func (server server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	server.logger.Info("textDocument/didChange", zap.String("DocumentURI", string(params.TextDocument.URI)))

	if len(params.ContentChanges) == 1 {
	server.cache.SetFile(params.TextDocument.URI, params.ContentChanges[0].Text)
	} else {
		server.logger.Warn("textDocument/didChange got unexpected amount of content changes", zap.Int("count", len(params.ContentChanges)))
	}

	return nil
}

func (server server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	server.logger.Info("textDocument/didClose", zap.String("DocumentURI", string(params.TextDocument.URI)))

	server.cache.DeleteFile(params.TextDocument.URI)

	return nil
}

