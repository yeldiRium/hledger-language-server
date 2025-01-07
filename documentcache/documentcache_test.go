package documentcache_test

import (
	"io"
	"testing"
	"testing/fstest"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/yeldiRium/hledger-language-server/documentcache"
)

func TestCache(t *testing.T) {
	t.Run("can add and retrieve files.", func(t *testing.T) {
		cache := documentcache.NewCache(fstest.MapFS{})
		cache.SetFile("tmp/foo.txt", "file content")

		content, ok := cache.GetFile("tmp/foo.txt")

		assert.True(t, ok)
		assert.Equal(t, "file content", content)
	})

	t.Run("returns w/e, false for files that are not cached.", func(t *testing.T) {
		cache := documentcache.NewCache(fstest.MapFS{})

		_, ok := cache.GetFile("tmp/doesnt-exist")

		assert.False(t, ok)
	})

	t.Run("can overwrite a cached file.", func(t *testing.T) {
		cache := documentcache.NewCache(fstest.MapFS{})
		cache.SetFile("tmp/foo.txt", "file content")
		cache.SetFile("tmp/foo.txt", "file content overwritten")

		content, ok := cache.GetFile("tmp/foo.txt")

		assert.True(t, ok)
		assert.Equal(t, "file content overwritten", content)
	})

	t.Run("can delete a cached file.", func(t *testing.T) {
		cache := documentcache.NewCache(fstest.MapFS{})
		cache.SetFile("tmp/foo.txt", "file content")
		cache.DeleteFile("tmp/foo.txt")

		_, ok := cache.GetFile("tmp/foo.txt")

		assert.False(t, ok)
	})

	t.Run("Open", func(t *testing.T) {
		t.Run("fails if the file is neither in the cache nor can be found in the workspace", func(t *testing.T) {
			cache := documentcache.NewCache(fstest.MapFS{})

			_, err := cache.Open("tmp/foo.txt")

			assert.IsError(t, err, documentcache.ErrFileNotFound)
		})

		t.Run("returns a file from the cache.", func(t *testing.T) {
			cache := documentcache.NewCache(fstest.MapFS{})
			cache.SetFile("tmp/foo.txt", "file content")

			file, err := cache.Open("tmp/foo.txt")

			assert.NoError(t, err)

			fileContent, err := io.ReadAll(file)
			assert.NoError(t, err)
			assert.Equal(t, "file content", string(fileContent))

			fileInfo, err := file.Stat()
			assert.NoError(t, err)
			assert.Equal(t, "tmp/foo.txt", fileInfo.Name())
			assert.Equal(t, 12, fileInfo.Size())
		})

		t.Run("reads a file from the workspace FS if it is not found in the cache, then adds it to the cache.", func(t *testing.T) {
			cache := documentcache.NewCache(fstest.MapFS{
				"tmp/foo.txt": &fstest.MapFile{
					Data: []byte("file content"),
				},
			})

			file, err := cache.Open("tmp/foo.txt")

			assert.NoError(t, err)

			fileContent, err := io.ReadAll(file)
			assert.NoError(t, err)
			assert.Equal(t, "file content", string(fileContent))

			fileInfo, err := file.Stat()
			assert.NoError(t, err)
			assert.Equal(t, "tmp/foo.txt", fileInfo.Name())
			assert.Equal(t, 12, fileInfo.Size())

			_, ok := cache.GetFile("tmp/foo.txt")
			assert.True(t, ok)
		})

		t.Run("does not yet track the last modified time of cached documents.", func(t *testing.T) {
			cache := documentcache.NewCache(fstest.MapFS{})
			cache.SetFile("tmp/foo.txt", "file content")

			file, _ := cache.Open("tmp/foo.txt")
			fileInfo, _ := file.Stat()
			assert.Equal(t, time.Unix(0, 0), fileInfo.ModTime())
		})
	})
}
