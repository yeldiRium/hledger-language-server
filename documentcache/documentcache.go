package documentcache

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"sync"
	"time"
)

var (
	ErrFileNotFound = fmt.Errorf("file not found")
)

type DocumentCache struct {
	sync.RWMutex
	files     map[string]string
	workspace fs.FS
}

func NewCache(workspace fs.FS) *DocumentCache {
	return &DocumentCache{
		files:     make(map[string]string),
		workspace: workspace,
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

func (fs *DocumentCache) Open(name string) (fs.File, error) {
	fileContent, ok := fs.GetFile(name)
	if !ok {
		file, err := fs.workspace.Open(name)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFileNotFound, err)
		}

		rawFileContent, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		fileContent = string(rawFileContent)

		fs.SetFile(name, fileContent)
	}

	return &documentCacheFile{
		Buffer:      bytes.NewBuffer([]byte(fileContent)),
		fileContent: fileContent,
		fileName:    name,
		// TODO: track last modified time
		lastModified: time.Unix(0, 0),
	}, nil
}

type documentCacheFile struct {
	*bytes.Buffer

	fileContent  string
	fileName     string
	lastModified time.Time
}

func (file *documentCacheFile) Stat() (fs.FileInfo, error) {
	return file, nil
}
func (file *documentCacheFile) Read(buffer []byte) (int, error) {
	return file.Buffer.Read(buffer)
}
func (file *documentCacheFile) Close() error {
	return nil
}

func (file *documentCacheFile) Name() string {
	return file.fileName
}
func (file *documentCacheFile) Size() int64 {
	return int64(len(file.fileContent))
}
func (file *documentCacheFile) Mode() fs.FileMode {
	return fs.ModePerm
}
func (file *documentCacheFile) ModTime() time.Time {
	return file.lastModified
}
func (file *documentCacheFile) IsDir() bool {
	return false
}
func (file *documentCacheFile) Sys() any {
	return nil
}
