package server

import (
	"context"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

type server struct {
	protocol.Server
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

func (h server) Shutdown(ctx context.Context) error {
	return nil
}

func NewServer(ctx context.Context, protocolServer protocol.Server, logger *zap.Logger) (server, context.Context, error) {
	// Do initialization logic here, including
	// stuff like setting state variables
	// by returning a new context with
	// context.WithValue(context, ...)
	// instead of just context
	return server{Server: protocolServer, logger: logger}, ctx, nil
}
