package cache

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// PopulateCacheWithPubSlotNameHash will put the slot names and hashes for a publisher in cache
func (c *cache) populateCacheWithPubSlotNameHash(pubid int) {
	cacheKey := key(PubSlotNameHash, pubid)
	if _, ok := c.cache.Get(cacheKey); ok {
		return
	}

	publisherSlotNameHashMap := c.db.GetPublisherSlotNameHash(pubid)
	if publisherSlotNameHashMap != nil {
		c.cache.Set(cacheKey, publisherSlotNameHashMap, time.Duration(c.cfg.CacheDefaultExpiry))
	}
}

// PopulateCacheWithWrapperSlotMappings will get the SlotMappings from database and put them in cache.
func (c *cache) populateCacheWithWrapperSlotMappings(pubid int, partnerConfigMap map[int]map[string]string, profileId, displayVersion int) {
	partnerSlotMappingMap := c.db.GetWrapperSlotMappings(partnerConfigMap, profileId, displayVersion)

	//put a version level dummy entry in cache denoting mappings are present for this version
	cacheKey := key(PUB_SLOT_INFO, pubid, profileId, displayVersion, 0)
	c.cache.Set(cacheKey, make(map[string]models.SlotMapping, 0), time.Duration(c.cfg.CacheDefaultExpiry))

	if len(partnerSlotMappingMap) == 0 {
		for _, partnerConf := range partnerConfigMap {
			partnerID, _ := strconv.Atoi(partnerConf[models.PARTNER_ID])
			cacheKey = key(PUB_SLOT_INFO, pubid, profileId, displayVersion, partnerID)
			c.cache.Set(cacheKey, make(map[string]models.SlotMapping, 0), time.Duration(c.cfg.CacheDefaultExpiry))
		}
		return
	}

	var nameHashMap map[string]string
	obj, ok := c.cache.Get(key(PubSlotNameHash, pubid))
	if ok && obj != nil {
		nameHashMap = obj.(map[string]string)
	}

	for partnerID, slotMappingList := range partnerSlotMappingMap {
		slotNameToMappingMap := make(map[string]models.SlotMapping, len(slotMappingList))
		slotNameToHashValueMap := make(map[string]string, len(slotMappingList))
		slotNameOrderedList := make([]string, 0)
		sort.Slice(slotMappingList, func(i, j int) bool {
			return slotMappingList[i].OrderID < slotMappingList[j].OrderID
		})
		for _, slotMapping := range slotMappingList {
			slotMapping.Hash = nameHashMap[slotMapping.SlotName]

			var mappingJSONObj map[string]interface{}
			if err := json.Unmarshal([]byte(slotMapping.MappingJson), &mappingJSONObj); err != nil {
				continue
			}

			cfgMap := partnerConfigMap[partnerID]
			bidderCode := cfgMap[models.BidderCode]
			if bidderCode == string(openrtb_ext.BidderPubmatic) || bidderCode == string(openrtb_ext.BidderGroupm) {
				//Adding slotName from DB in fieldMap for PubMatic, as slotName from DB should be sent to PubMatic instead of slotName from request
				//This is required for case in-sensitive mapping
				mappingJSONObj[models.KEY_OW_SLOT_NAME] = slotMapping.SlotName
			}

			slotMapping.SlotMappings = mappingJSONObj
			slotNameToMappingMap[strings.ToLower(slotMapping.SlotName)] = slotMapping
			slotNameToHashValueMap[slotMapping.SlotName] = slotMapping.Hash
			slotNameOrderedList = append(slotNameOrderedList, slotMapping.SlotName)
		}
		cacheKey = key(PUB_SLOT_INFO, pubid, profileId, displayVersion, partnerID)
		c.cache.Set(cacheKey, slotNameToMappingMap, time.Duration(c.cfg.CacheDefaultExpiry))

		slotMappingInfoObj := models.SlotMappingInfo{
			OrderedSlotList: slotNameOrderedList,
			HashValueMap:    slotNameToHashValueMap,
		}
		cacheKey = key(PubSlotHashInfo, pubid, profileId, displayVersion, partnerID)
		c.cache.Set(cacheKey, slotMappingInfoObj, time.Duration(c.cfg.CacheDefaultExpiry))
	}
}
