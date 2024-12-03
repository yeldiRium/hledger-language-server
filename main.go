package main

import (
	"context"
	"io"
	"os"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/yeldiRium/hledger-language-server/server"
)

func main() {
	logger, _ := zap.NewDevelopmentConfig().Build()
	connection := jsonrpc2.NewConn(jsonrpc2.NewStream(&rwCloser{os.Stdin, os.Stdout}))

	client := protocol.ClientDispatcher(connection, logger)
	handler, ctx, err := server.NewServer(context.Background(), protocol.ServerDispatcher(connection, logger), client, logger)
	if err != nil {
		logger.Sugar().Fatalf("while initializing handler: %w", err)
	}

	connection.Go(ctx, protocol.ServerHandler(handler, jsonrpc2.MethodNotFoundHandler))
	<-connection.Done()
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
