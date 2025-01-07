package documentcache_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.lsp.dev/uri"

	"github.com/yeldiRium/hledger-language-server/documentcache"
)

func TestCache(t *testing.T) {
	t.Run("can add and retrieve files", func(t *testing.T) {
		cache := documentcache.NewCache()
		cache.SetFile(uri.New("file:///tmp/foo.txt"), "file content")

		content, ok := cache.GetFile(uri.New("file:///tmp/foo.txt"))

		assert.True(t, ok)
		assert.Equal(t, "file content", content)
	})

	t.Run("returns w/e, false for files that are not cached", func(t *testing.T) {
		cache := documentcache.NewCache()

		_, ok := cache.GetFile(uri.New("file:///tmp/doesnt-exist"))

		assert.False(t, ok)
	})

	t.Run("can overwrite a cached file", func(t *testing.T) {
		cache := documentcache.NewCache()
		cache.SetFile(uri.New("file:///tmp/foo.txt"), "file content")
		cache.SetFile(uri.New("file:///tmp/foo.txt"), "file content overwritten")

		content, ok := cache.GetFile(uri.New("file:///tmp/foo.txt"))

		assert.True(t, ok)
		assert.Equal(t, "file content overwritten", content)
	})

	t.Run("can delete a cached file", func(t *testing.T) {
		cache := documentcache.NewCache()
		cache.SetFile(uri.New("file:///tmp/foo.txt"), "file content")
		cache.DeleteFile(uri.New("file:///tmp/foo.txt"))

		_, ok := cache.GetFile(uri.New("file:///tmp/foo.txt"))

		assert.False(t, ok)
	})
}
