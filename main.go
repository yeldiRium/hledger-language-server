package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/yeldiRium/hledger-language-server/server"
)

func main() {
	dir, err := os.MkdirTemp("", "hledger-language-server")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	logFile := filepath.Join(dir, "log.txt")

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.OutputPaths = []string{logFile}
	zapConfig.ErrorOutputPaths = []string{logFile}
	logger, _ := zapConfig.Build()
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
