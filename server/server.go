package server

import (
	"context"
	"sync"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

type server struct {
	protocol.Server
	client protocol.Client
	logger *zap.Logger
	cache *cache
}

type cache struct {
	sync.RWMutex
	files map[uri.URI]string
}

func newCache() *cache {
	return &cache{
		files: make(map[uri.URI]string),
	}
}

func (c *cache) GetFile(documentURI uri.URI) (string, bool) {
	c.RLock()
	defer c.RUnlock()

	fileContent, ok := c.files[documentURI]
	return fileContent, ok
}

func (c *cache) AddFile(documentURI uri.URI, content string) {
	c.Lock()
	defer c.Unlock()

	c.files[documentURI] = content
}

// TODO: Remove document from cache when it is closed

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

	server.cache.AddFile(params.TextDocument.URI, params.TextDocument.Text)

	return nil
}

func (server server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	server.logger.Info("textDocument/didChange", zap.String("DocumentURI", string(params.TextDocument.URI)))

	if len(params.ContentChanges) == 1 {
	server.cache.AddFile(params.TextDocument.URI, params.ContentChanges[0].Text)
	} else {
		server.logger.Warn("textDocument/didChange got unexpected amount of content changes", zap.Int("count", len(params.ContentChanges)))
	}

	return nil
}

func (server server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	server.logger.Info("textDocument/didClose", zap.String("DocumentURI", string(params.TextDocument.URI)))
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
		cache: newCache(),
	}, ctx, nil
}

