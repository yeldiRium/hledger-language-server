package server

import (
	"context"
	"fmt"
	"os"
	"path"

	"go.lsp.dev/protocol"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func (h server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	h.logger.Sugar().Infof("hovering over %s in line %d at position %d", params.TextDocument.URI, params.Position.Line, params.Position.Character)

	lineNumber := int(params.Position.Line + 1)
	columnNumber := int(params.Position.Character + 1)

	filename := params.TextDocument.URI.Filename()
	parser := ledger.NewJournalParser()
	fileHandle, err := os.Open(filename)
	defer fileHandle.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	journal, err := parser.Parse(filename, fileHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to parse journal: %w", err)
	}
	resolvedJournal, err := ledger.ResolveIncludes(journal, parser, os.DirFS(path.Dir(filename)))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve includes: %w", err)
	}

	accountNameUnderCursor := ledger.FindAccountNameUnderCursor(resolvedJournal, lineNumber, columnNumber)

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
