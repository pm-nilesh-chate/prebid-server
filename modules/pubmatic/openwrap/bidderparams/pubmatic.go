package bidderparams

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/request"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func getSlotMeta(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt request.ImpExtension, partnerID int) ([]string, map[string]models.SlotMapping, models.SlotMappingInfo) {
	var slotMap map[string]models.SlotMapping
	var slotMappingInfo models.SlotMappingInfo

	//don't read mappings from cache in case of test=2
	if rctx.IsTestRequest {
		slotMap = cache.GetMappingsFromCacheV25(rctx, partnerID)
		if slotMap == nil {
			return nil, nil, models.SlotMappingInfo{}
		}
		slotMappingInfo = cache.GetSlotToHashValueMapFromCacheV25(rctx, partnerID)
		if len(slotMappingInfo.OrderedSlotList) == 0 {
			return nil, nil, models.SlotMappingInfo{}
		}
	}

	var wh [][2]int64
	if imp.Banner != nil {
		if imp.Banner.W != nil && imp.Banner.H != nil {
			wh = append(wh, [2]int64{*imp.Banner.H, *imp.Banner.W})
		}

		for _, format := range imp.Banner.Format {
			wh = append(wh, [2]int64{format.H, format.W})
		}
	}

	if imp.Video != nil {
		wh = append(wh, [2]int64{0, 0})
	}

	kgp := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN]

	var src string
	if bidRequest.Site != nil {
		if bidRequest.Site.Domain != "" {
			src = bidRequest.Site.Domain
		} else if bidRequest.Site.Page != "" {
			src = bidRequest.Site.Page
		}
	} else if bidRequest.App != nil && bidRequest.App.Bundle != "" {
		src = bidRequest.App.Bundle
	}

	var slots []string
	for _, format := range wh {
		slot := generateSlotName(format[0], format[1], kgp, imp.TagID, impExt.Wrapper.Div, src)
		if slot != "" {
			slots = append(slots, slot)
			// NYC_TODO: break at i=0 for pubmatic?
		}
	}

	return slots, slotMap, slotMappingInfo
}

func PreparePubMaticParamsV25(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt request.ImpExtension, partnerID int) ([]byte, errorcodes.IError) {
	slots, slotMap, slotMappingInfo := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)
	for _, slot := range slots {
		params, err := prepareBidParamForPubmaticV25(rctx, slot, slotMap, slotMappingInfo, bidRequest, imp, impExt, partnerID, false)
		if err != nil || params == nil {
			continue
		}
		return params, nil
	}
	// isRegex := kgp == models.REGEX_KGP
	// _ = isRegex
	//isRegex
	// for _, slot := range slots {
	// 	slot =
	// 	prepareBidParamForPubmaticV25(rctx, slot, slotMap, slotMappingInfo, bidRequest, imp, impExt, partnerID, true)
	// }

	return nil, nil
}

func prepareBidParamForPubmaticV25(rctx models.RequestCtx, slot string, slotMap map[string]models.SlotMapping, slotMappingInfo models.SlotMappingInfo, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt request.ImpExtension, partnerID int, isRegex bool) ([]byte, error) {
	var fieldMap map[string]interface{}
	var err error
	if !rctx.IsTestRequest && !isRegex {
		fieldMap, err = CheckSlotName(slot, isRegex, slotMap)
		if err != nil {
			return nil, err
		}
	}

	wrapExt := fmt.Sprintf(`{"%s":%d,"%s":%d}`, models.SS_PM_VERSION_ID, rctx.DisplayID, models.SS_PM_PROFILE_ID, rctx.ProfileID)
	extImpPubMatic := openrtb_ext.ExtImpPubmatic{
		PublisherId: strconv.Itoa(rctx.PubID),
		WrapExt:     json.RawMessage(wrapExt),
		AdSlot:      slot,
		Keywords:    getImpExtPubMaticKeyWords(impExt, rctx.PartnerConfigMap[partnerID][models.BidderCode]),
		DealTier:    getDealTier(impExt, rctx.PartnerConfigMap[partnerID][models.BidderCode]),
	}

	// NYC_TODO: check with translator if this is required.
	// if partnerConf[constant.KEY_GEN_PATTERN] == constant.REGEX_KGP {
	// 	slotKey = hashValue
	// } else if value, ok := fieldMap[constant.KEY_OW_SLOT_NAME]; ok && nil != value {
	// 	slotKey = fmt.Sprintf("%v", value)
	// }
	//
	// Update slot key for PubMatic secondary flow
	// if value, ok := fieldMap[constant.KEY_SLOT_NAME]; ok && nil != value {
	// 	slotKey = fmt.Sprintf("%v", value)
	// }
	_ = fieldMap

	return json.Marshal(extImpPubMatic)
}

func getDealTier(impExt request.ImpExtension, bidderCode string) *openrtb_ext.DealTier {
	if len(impExt.Bidder) != 0 {
		if bidderExt, ok := impExt.Bidder[bidderCode]; ok && bidderExt != nil && bidderExt.DealTier != nil {
			return bidderExt.DealTier
		}
	}
	return nil
}

func getImpExtPubMaticKeyWords(impExt request.ImpExtension, bidderCode string) []*openrtb_ext.ExtImpPubmaticKeyVal {
	if len(impExt.Bidder) != 0 {
		if bidderExt, ok := impExt.Bidder[bidderCode]; ok && bidderExt != nil && len(bidderExt.KeyWords) != 0 {
			keywords := make([]*openrtb_ext.ExtImpPubmaticKeyVal, 0)
			for _, keyVal := range bidderExt.KeyWords {
				//ignore key values pair with no values
				if len(keyVal.Values) == 0 {
					continue
				}
				keyValPair := openrtb_ext.ExtImpPubmaticKeyVal{
					Key:    keyVal.Key,
					Values: keyVal.Values,
				}
				keywords = append(keywords, &keyValPair)
			}
			if len(keywords) != 0 {
				return keywords
			}
		}
	}
	return nil
}

func UpdateRequestExtBidderParamsForPubmatic(bidderParams *json.RawMessage, cookie string, reqLoggerImpID, loggerImpID, platform, bidderCode string) {
	bidderParamsMap := make(map[string]map[string]interface{})
	err := json.Unmarshal(*bidderParams, &bidderParamsMap)
	if err != nil {
		return
	}

	params := map[string]interface{}{
		models.COOKIE: cookie,
	}

	//if platform is display set wiid as req.ext.wrapper.wiid, otherwise use
	if platform == models.PLATFORM_DISPLAY {
		params[models.WrapperLoggerImpID] = reqLoggerImpID
	} else {
		params[models.WrapperLoggerImpID] = loggerImpID
	}
	bidderParamsMap[bidderCode] = params

	*bidderParams, _ = json.Marshal(bidderParamsMap)
}
