package documentcache_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/yeldiRium/hledger-language-server/documentcache"
)

func TestCache(t *testing.T) {
	t.Run("can add and retrieve files", func(t *testing.T) {
		cache := documentcache.NewCache()
		cache.SetFile("file:///tmp/foo.txt", "file content")

		content, ok := cache.GetFile("file:///tmp/foo.txt")

		assert.True(t, ok)
		assert.Equal(t, "file content", content)
	})

	t.Run("returns w/e, false for files that are not cached", func(t *testing.T) {
		cache := documentcache.NewCache()

		_, ok := cache.GetFile("file:///tmp/doesnt-exist")

		assert.False(t, ok)
	})

	t.Run("can overwrite a cached file", func(t *testing.T) {
		cache := documentcache.NewCache()
		cache.SetFile("file:///tmp/foo.txt", "file content")
		cache.SetFile("file:///tmp/foo.txt", "file content overwritten")

		content, ok := cache.GetFile("file:///tmp/foo.txt")

		assert.True(t, ok)
		assert.Equal(t, "file content overwritten", content)
	})

	t.Run("can delete a cached file", func(t *testing.T) {
		cache := documentcache.NewCache()
		cache.SetFile("file:///tmp/foo.txt", "file content")
		cache.DeleteFile("file:///tmp/foo.txt")

		_, ok := cache.GetFile("file:///tmp/foo.txt")

		assert.False(t, ok)
	})
}
