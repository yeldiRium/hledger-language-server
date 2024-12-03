package server

import (
	"context"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

type server struct {
	protocol.Server
	client protocol.Client
	logger *zap.Logger
}

func (h server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			HoverProvider: true,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "hledger-language-server",
			Version: "0.0.1",
		},
	}, nil
}

func (server server) Initialized(ctx context.Context, parames *protocol.InitializedParams) error {
	server.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type: protocol.MessageTypeInfo,
		Message: "Hello there!",
	})
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
	return server{Server: protocolServer, client: protocolClient, logger: logger}, ctx, nil
}

