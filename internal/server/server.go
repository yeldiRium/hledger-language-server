package server

import (
	"context"
	"encoding/json"
	"os"

	"go.lsp.dev/protocol"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/yeldirium/hledger-language-server/internal/documentcache"
	"github.com/yeldirium/hledger-language-server/internal/parsercache"
)

type server struct {
	protocol.Server
	client        protocol.Client
	logger        *zap.Logger
	documentCache *documentcache.DocumentCache
	parserCache   *parsercache.ParserCache
	clientInformation clientInformation
}

type clientInformation struct {
	clientName string
	clientVersion string
	clientLocale string
	clientParentProcessID int
}

func (c clientInformation) AddToSpan(span trace.Span) {
	span.SetAttributes(
		attribute.String("client.name", c.clientName),
		attribute.String("client.version", c.clientVersion),
		attribute.String("client.locale", c.clientLocale),
		attribute.Int("client.parentProcessID", c.clientParentProcessID),
	)
}

func collectServerCapabilities() protocol.ServerCapabilities {
	capabilities := protocol.ServerCapabilities{}
	registerCompletionCapabilities(&capabilities)
	registerDocumentSyncCapabilities(&capabilities)
	registerHoverCapabilities(&capabilities)
	return capabilities
}

func (server server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	span := trace.SpanFromContext(ctx)
	
	server.clientInformation = clientInformation{
		clientName: params.ClientInfo.Name,
		clientVersion: params.ClientInfo.Version,
		clientLocale: params.Locale,
		clientParentProcessID: int(params.ProcessID),
	}
	server.clientInformation.AddToSpan(span)

	clientCapabilitiesJson, err := json.Marshal(params.Capabilities)
	if err != nil {
		span.SetAttributes(
			attribute.String("lsp.clientCapabilities", string(clientCapabilitiesJson)),
		)
	}

	capabilities := collectServerCapabilities()
	serverCapabilitiesJson, err := json.Marshal(capabilities)
	if err != nil {
		span.SetAttributes(
			attribute.String("lsp.serverCapabilities", string(serverCapabilitiesJson)),
		)
	}

	return &protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.ServerInfo{
			Name:    "hledger-language-server",
			Version: "0.0.1",
		},
	}, nil
}

func (server server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	span := trace.SpanFromContext(ctx)
	server.clientInformation.AddToSpan(span)

	return nil
}

func (server server) Shutdown(ctx context.Context) error {
	span := trace.SpanFromContext(ctx)
	server.clientInformation.AddToSpan(span)

	return nil
}

// Request catches all requests that are not handled otherwise. The main purpose
// for this is to catche $/cancelRequest requests, which we do not handle yet.
// TODO: handle cancelRequests so that each handler can opt-in to cancellation
func (server server) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	span := trace.SpanFromContext(ctx)
	server.clientInformation.AddToSpan(span)

	return struct{}{}, nil
}

func NewServer(ctx context.Context, protocolServer protocol.Server, protocolClient protocol.Client, logger *zap.Logger) (server, context.Context, error) {
	// Do initialization logic here, including
	// stuff like setting state variables
	// by returning a new context with
	// context.WithValue(context, ...)
	// instead of just context
	documentCache := documentcache.NewCache(os.DirFS("/"))
	return server{
		Server: protocolServer,
		client: protocolClient,
		logger: logger,
		// TODO: set cache workspace based on project workspace reported from client
		documentCache: documentCache,
		parserCache:   parsercache.NewCache(documentCache),
		clientInformation: clientInformation{},
	}, ctx, nil
}
