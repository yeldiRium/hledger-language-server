package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func (h server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	h.logger.Info(
		"textDocument/hover",
		zap.String("DocumentURI", string(params.TextDocument.URI)),
		zap.Uint32("line", params.Position.Line),
		zap.Uint32("character", params.Position.Character),
	)

	lineNumber := int(params.Position.Line + 1)
	columnNumber := int(params.Position.Character + 1)

	fileName := params.TextDocument.URI.Filename()
	fileName = strings.TrimPrefix(fileName, "file://")

	fileContent, ok := h.cache.GetFile(params.TextDocument.URI)
	var fileReader io.Reader
	if !ok {
		h.logger.Warn(
			"textDocument/hover target not found in cache, opening from file system",
			zap.String("DocumentURI", string(params.TextDocument.URI)),
		)
		fileReader, err := os.Open(fileName)
		defer fileReader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
	} else {
		h.logger.Info(
			"textDocument/hover target found in cache",
			zap.String("DocumentURI", string(params.TextDocument.URI)),
		)
		fileReader = bytes.NewBuffer([]byte(fileContent))
	}

	parser := ledger.NewJournalParser()
	journal, err := parser.Parse(fileName, fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse journal: %w", err)
	}
	resolvedJournal, err := ledger.ResolveIncludes(journal, parser, os.DirFS(path.Dir(fileName)))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve includes: %w", err)
	}

	accountNameUnderCursor := ledger.FindAccountNameUnderCursor(resolvedJournal, fileName, lineNumber, columnNumber)

	if accountNameUnderCursor == nil {
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: "Not hovering over an account name",
			},
		}, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: fmt.Sprintf("You're hovering over \"%s\"", accountNameUnderCursor),
		},
	}, nil
}
