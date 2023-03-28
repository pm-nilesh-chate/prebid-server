package cache

// NYC_TODO: Return DB level errors for module logging
func (c *cache) GetPartnerConfigMap(pubid, profileid, displayversion int) (map[int]map[string]string, error) {
	c.populateCacheWithPubSlotNameHash(pubid)
	c.populatePublisherVASTTags(pubid)

	cacheKey := key(PUB_HB_PARTNER, pubid, profileid, displayversion)
	if obj, ok := c.cache.Get(cacheKey); ok {
		return obj.(map[int]map[string]string), nil
	}

	partnerConfigMap, err := c.db.GetActivePartnerConfigurations(pubid, profileid, displayversion)
	if len(partnerConfigMap) != 0 {
		c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))

		c.populateCacheWithWrapperSlotMappings(pubid, partnerConfigMap, profileid, displayversion)
		c.populateCacheWithAdunitConfig(pubid, profileid, displayversion)
	}
	return partnerConfigMap, err
}
