package documentcache

import (
	"sync"
)

type DocumentCache struct {
	sync.RWMutex
	files map[string]string
}

func NewCache() *DocumentCache {
	return &DocumentCache{
		files: make(map[string]string),
	}
}

func (c *DocumentCache) GetFile(fileName string) (string, bool) {
	c.RLock()
	defer c.RUnlock()

	fileContent, ok := c.files[fileName]
	return fileContent, ok
}

func (c *DocumentCache) SetFile(fileName string, content string) {
	c.Lock()
	defer c.Unlock()

	c.files[fileName] = content
}

func (c *DocumentCache) DeleteFile(fileName string) {
	c.Lock()
	defer c.Unlock()

	delete(c.files, fileName)
}
