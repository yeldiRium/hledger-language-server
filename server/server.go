package server

import (
	"context"
	"os"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/yeldiRium/hledger-language-server/documentcache"
)

type server struct {
	protocol.Server
	client protocol.Client
	logger *zap.Logger
	cache  *documentcache.DocumentCache
}

func collectServerCapabilities() protocol.ServerCapabilities {
	capabilities := protocol.ServerCapabilities{}
	registerCompletionCapabilities(&capabilities)
	registerDocumentSyncCapabilities(&capabilities)
	registerHoverCapabilities(&capabilities)
	return capabilities
}

func (s server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	capabilities := collectServerCapabilities()
	s.logger.Info("initialize", zap.Any("serverCapabilities", capabilities))

	return &protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.ServerInfo{
			Name:    "hledger-language-server",
			Version: "0.0.1",
		},
	}, nil
}

func (server server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	server.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type:    protocol.MessageTypeInfo,
		Message: "Hello there!",
	})
	return nil
}

func (h server) Shutdown(ctx context.Context) error {
	return nil
}

// Request catches all requests that are not handled otherwise. The main purpose
// for this is to catche $/cancelRequest requests, which we do not handle yet.
// TODO: handle cancelRequests so that each handler can opt-in to cancellation
func (server server) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	server.logger.Debug(
		"request",
		zap.String("method", method),
		zap.Any("params", params),
	)

	return struct{}{}, nil
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
		// TODO: set cache workspace based on project workspace reported from client
		cache:  documentcache.NewCache(os.DirFS("/")),
	}, ctx, nil
}
