package server

import (
	"strings"

	"go.lsp.dev/uri"
)

// getFilePathFromURI takes the file name from a URI and removes its prefix
// until it is a relative path.
func getFilePathFromURI(documentURI uri.URI) string {
	filePath := documentURI.Filename()
	trimmedFilePath := strings.TrimPrefix(filePath, "file:///")

	return trimmedFilePath
}
