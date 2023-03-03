package cache

import (
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type Cache interface {
	GetPartnerConfigMap(bidRequest *openrtb2.BidRequest, pubid, profileid, displayversion int) map[int]map[string]string
	GetAdunitConfigFromCache(request *openrtb2.BidRequest, pubID int, profileID, displayVersion int) models.AdUnitConfig
	GetMappingsFromCacheV25(rctx models.RequestCtx, partnerID int) map[string]models.SlotMapping
	GetSlotToHashValueMapFromCacheV25(rctx models.RequestCtx, partnerID int) models.SlotMappingInfo
	GetPublisherVASTTagsFromCache(pubID int) models.PublisherVASTTags
}

// type internalCache interface {
// }