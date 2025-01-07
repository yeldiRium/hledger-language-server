package server

import (
	"strings"

	"go.lsp.dev/uri"
)

func getFileNameFromURI(documentURI uri.URI) string {
	fileName := documentURI.Filename()
	trimmedFileName := strings.TrimPrefix(fileName, "file://")

	return trimmedFileName
}
