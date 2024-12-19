package cache

import (
	"sync"

	"go.lsp.dev/uri"
)

type Cache struct {
	sync.RWMutex
	files map[uri.URI]string
}

func NewCache() *Cache {
	return &Cache{
		files: make(map[uri.URI]string),
	}
}

func (c *Cache) GetFile(documentURI uri.URI) (string, bool) {
	c.RLock()
	defer c.RUnlock()

	fileContent, ok := c.files[documentURI]
	return fileContent, ok
}

func (c *Cache) SetFile(documentURI uri.URI, content string) {
	c.Lock()
	defer c.Unlock()

	c.files[documentURI] = content
}

func (c *Cache) DeleteFile(documentURI uri.URI) {
	delete(c.files, documentURI)
}
