package cache

import (
	"sync"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
)

// NYC_TODO: refactor this to inject any kind of cache,replace cache with freecache library
// any db or cache should be injectable
type cache struct {
	cache *gocache.Cache
	cfg   config.Cache
	db    database.Database
}

var c *cache
var cOnce sync.Once

func New(goCache *gocache.Cache, database database.Database, cfg config.Cache) *cache {
	cOnce.Do(
		func() {
			c = &cache{
				cache: goCache,
				db:    database,
				cfg:   cfg,
			}
		})
	return c
}
