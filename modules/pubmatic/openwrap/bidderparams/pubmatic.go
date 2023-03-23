package bidderparams

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func PreparePubMaticParamsV25(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) ([]byte, errorcodes.IError) {
	slots, slotMap, slotMappingInfo, _ := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)
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

func prepareBidParamForPubmaticV25(rctx models.RequestCtx, slot string, slotMap map[string]models.SlotMapping, slotMappingInfo models.SlotMappingInfo, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int, isRegex bool) ([]byte, error) {
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

func getDealTier(impExt models.ImpExtension, bidderCode string) *openrtb_ext.DealTier {
	if len(impExt.Bidder) != 0 {
		if bidderExt, ok := impExt.Bidder[bidderCode]; ok && bidderExt != nil && bidderExt.DealTier != nil {
			return bidderExt.DealTier
		}
	}
	return nil
}

func getImpExtPubMaticKeyWords(impExt models.ImpExtension, bidderCode string) []*openrtb_ext.ExtImpPubmaticKeyVal {
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
