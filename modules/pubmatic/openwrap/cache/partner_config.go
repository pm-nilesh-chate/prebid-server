package cache

import (
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// NYC_TODO: Return DB level errors for module logging
func (c *cache) GetPartnerConfigMap(bidRequest *openrtb2.BidRequest, pubid, profileid, displayversion int) map[int]map[string]string {
	if int(bidRequest.Test) == 2 {
		return getTestModePartnerConfigMap(bidRequest, pubid, profileid, displayversion)
	}

	if profileid == 0 {
		return getDefaultPartnerConfigMap(bidRequest, pubid, profileid, displayversion)
	}

	return c.getActivePartnerConfigMap(bidRequest, pubid, profileid, displayversion)
}

func getTestModePartnerConfigMap(bidRequest *openrtb2.BidRequest, pubid, profileid, displayversion int) map[int]map[string]string {
	var platform string
	if bidRequest.Site != nil {
		platform = models.PLATFORM_DISPLAY
	} else if bidRequest.App != nil {
		platform = models.PLATFORM_APP
	}

	return map[int]map[string]string{
		1: {
			models.PARTNER_ID:          models.PUBMATIC_PARTNER_ID_STRING,
			models.PREBID_PARTNER_NAME: string(openrtb_ext.BidderPubmatic),
			models.BidderCode:          string(openrtb_ext.BidderPubmatic),
			models.SERVER_SIDE_FLAG:    models.PUBMATIC_SS_FLAG,
			models.KEY_GEN_PATTERN:     models.ADUNIT_SIZE_KGP,
			// models.TIMEOUT:             strconv.Itoa(m.cfg.OpenWrap.Timeout.HBTimeout),
		},
		-1: {
			models.PLATFORM_KEY:     platform,
			models.DisplayVersionID: strconv.Itoa(displayversion),
		},
	}
}

// NYC_TODO: cache, etc
func getDefaultPartnerConfigMap(bidRequest *openrtb2.BidRequest, pubid, profileid, displayversion int) map[int]map[string]string {
	return map[int]map[string]string{
		1: {
			models.PARTNER_ID:          models.PUBMATIC_PARTNER_ID_STRING,
			models.PREBID_PARTNER_NAME: string(openrtb_ext.BidderPubmatic),
			models.BidderCode:          string(openrtb_ext.BidderPubmatic),
			models.PROTOCOL:            models.PUBMATIC_PROTOCOL,
			models.SERVER_SIDE_FLAG:    models.PUBMATIC_SS_FLAG,
			models.KEY_GEN_PATTERN:     models.ADUNIT_SIZE_KGP,
			models.LEVEL:               models.PUBMATIC_LEVEL,
			// models.TIMEOUT:             strconv.Itoa(m.cfg.OpenWrap.Timeout.HBTimeout),
		},
	}
}

func (c *cache) getActivePartnerConfigMap(bidRequest *openrtb2.BidRequest, pubid, profileid, displayversion int) map[int]map[string]string {
	c.populateCacheWithPubSlotNameHash(pubid)
	c.populatePublisherVASTTags(pubid)

	cacheKey := key(PUB_HB_PARTNER, pubid, profileid, displayversion)
	if obj, ok := c.cache.Get(cacheKey); ok {
		return obj.(map[int]map[string]string)
	}

	partnerConfigMap := c.db.GetActivePartnerConfigurations(pubid, profileid, displayversion)
	if len(partnerConfigMap) != 0 {
		c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))

		c.populateCacheWithWrapperSlotMappings(pubid, partnerConfigMap, profileid, displayversion)
		c.populateCacheWithAdunitConfig(pubid, profileid, displayversion)
	}
	return partnerConfigMap
}
