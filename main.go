package main

import (
	"context"
	"io"
	"os"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopmentConfig().Build()
	connection := jsonrpc2.NewConn(jsonrpc2.NewStream(&rwCloser{os.Stdin, os.Stdout}))

	handler, ctx, err := NewHandler(context.Background(), protocol.ServerDispatcher(connection, logger), logger)
	if err != nil {
		logger.Sugar().Fatalf("while initializing handler: %w", err)
	}

	connection.Go(ctx, protocol.ServerHandler(handler, jsonrpc2.MethodNotFoundHandler))
	<-connection.Done()
}

type Handler struct {
	protocol.Server
	logger *zap.Logger
}

func (h Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
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

func (h Handler) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	h.logger.Sugar().Infof("hovering over %s", params.TextDocument.URI)
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: "Hello, world!",
		},
	}, nil
}

func NewHandler(ctx context.Context, server protocol.Server, logger *zap.Logger) (Handler, context.Context, error) {
	// Do initialization logic here, including
	// stuff like setting state variables
	// by returning a new context with
	// context.WithValue(context, ...)
	// instead of just context
	return Handler{Server: server, logger: logger}, ctx, nil
}

type rwCloser struct {
	io.ReadCloser
	io.WriteCloser
}

// SetWriteDeadline implements rpc.Conn.
func (rw rwCloser) SetWriteDeadline(time.Time) error {
	return nil
}

func (rw rwCloser) Close() error {
	err := rw.ReadCloser.Close()
	if err != nil {
		return err
	}
	err = rw.WriteCloser.Close()
	if err != nil {
		return err
	}
	return nil
}
