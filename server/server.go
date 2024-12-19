package server

import (
	"context"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/yeldiRium/hledger-language-server/cache"
)

type server struct {
	protocol.Server
	client protocol.Client
	logger *zap.Logger
	cache *cache.Cache
}

func (s server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			HoverProvider: true,
			TextDocumentSync: protocol.TextDocumentSyncKindFull,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "hledger-language-server",
			Version: "0.0.1",
		},
	}, nil
}

func (server server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	server.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type: protocol.MessageTypeInfo,
		Message: "Hello there!",
	})
	return nil
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

func (h server) Shutdown(ctx context.Context) error {
	return nil
}

func NewServer(ctx context.Context, protocolServer protocol.Server, protocolClient protocol.Client, logger *zap.Logger) (server, context.Context, error) {
	// Do initialization logic here, including
	// stuff like setting state variables
	// by returning a new context with
	// context.WithValue(context, ...)
	// instead of just context
	return server{
		Server: protocolServer,
		client: protocolClient,
		logger: logger,
		cache: cache.NewCache(),
	}, ctx, nil
}

