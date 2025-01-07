package server

import (
	"strings"

	"go.lsp.dev/uri"
)

func getFilePathFromURI(documentURI uri.URI) string {
	filePath := documentURI.Filename()
	trimmedFilePath := strings.TrimPrefix(filePath, "file://")

	return trimmedFilePath
}
