package cache

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// PopulatePublisherVASTTags will put publisher level VAST Tag details into cache
func (c *cache) populatePublisherVASTTags(pubid int) {
	cacheKey := key(PubVASTTags, pubid)
	if _, ok := c.cache.Get(cacheKey); ok {
		return
	}

	//get publisher level vast tag details from DB
	publisherVASTTags, err := c.db.GetPublisherVASTTags(pubid)
	if err != nil {
		return
	}

	if publisherVASTTags == nil {
		publisherVASTTags = models.PublisherVASTTags{}
	}

	c.cache.Set(cacheKey, publisherVASTTags, c.cfg.VASTTagCacheExpiry)
}
