package documentcache

import (
	"sync"

	"go.lsp.dev/uri"
)

type DocumentCache struct {
	sync.RWMutex
	files map[uri.URI]string
}

func NewCache() *DocumentCache {
	return &DocumentCache{
		files: make(map[uri.URI]string),
	}
}

func (c *DocumentCache) GetFile(documentURI uri.URI) (string, bool) {
	c.RLock()
	defer c.RUnlock()

	fileContent, ok := c.files[documentURI]
	return fileContent, ok
}

func (c *DocumentCache) SetFile(documentURI uri.URI, content string) {
	c.Lock()
	defer c.Unlock()

	c.files[documentURI] = content
}

func (c *DocumentCache) DeleteFile(documentURI uri.URI) {
	c.Lock()
	defer c.Unlock()

	delete(c.files, documentURI)
}
