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

	"github.com/yeldiRium/hledger-language-server/internal/server"
	"github.com/yeldiRium/hledger-language-server/internal/telemetry"
)

func main() {
	dir, err := os.MkdirTemp("", "hledger-language-server")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	logFile := filepath.Join(dir, "log.txt")

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.OutputPaths = []string{logFile}
	zapConfig.ErrorOutputPaths = []string{logFile}
	logger, _ := zapConfig.Build()
	connection := jsonrpc2.NewConn(jsonrpc2.NewStream(&rwCloser{os.Stdin, os.Stdout}))

	clientDispatcher := protocol.ClientDispatcher(connection, logger)
	serverDispatcher := protocol.ServerDispatcher(connection, logger)
	handler, ctx, err := server.NewServer(context.Background(), serverDispatcher, clientDispatcher, logger)
	if err != nil {
		logger.Sugar().Fatalf("while initializing handler: %w", err)
	}

	jsonRpcHandler := protocol.ServerHandler(handler, jsonrpc2.MethodNotFoundHandler)

	t, shutdown, err := telemetry.SetupTelemetry(ctx, logger)
	if err != nil {
		logger.Sugar().Errorf("failed to setup telemetry: %w", err)
	} else {
		defer func() {
			_ = shutdown(ctx)
		}()

		logger.Sugar().Infof("successfully connected to telemetry backend")
		jsonRpcHandler = telemetry.WrapInTelemetry(t, jsonRpcHandler)
	}

	if err != nil {
		logger.Sugar().Fatalf("failed to start telemetry instrumentation")
	}
	defer func() {
		_ = shutdown(ctx)
	}()

	connection.Go(ctx, jsonRpcHandler)
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
