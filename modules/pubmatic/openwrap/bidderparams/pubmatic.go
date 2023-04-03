package bidderparams

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func PreparePubMaticParamsV25(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) (string, string, bool, []byte, error) {
	wrapExt := fmt.Sprintf(`{"%s":%d,"%s":%d}`, models.SS_PM_VERSION_ID, rctx.DisplayID, models.SS_PM_PROFILE_ID, rctx.ProfileID)
	extImpPubMatic := openrtb_ext.ExtImpPubmatic{
		PublisherId: strconv.Itoa(rctx.PubID),
		WrapExt:     json.RawMessage(wrapExt),
		Keywords:    getImpExtPubMaticKeyWords(impExt, rctx.PartnerConfigMap[partnerID][models.BidderCode]),
		DealTier:    getDealTier(impExt, rctx.PartnerConfigMap[partnerID][models.BidderCode]),
	}

	slots, slotMap, slotMappingInfo, _ := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)

	if rctx.IsTestRequest {
		extImpPubMatic.AdSlot = slots[0]
		params, err := json.Marshal(extImpPubMatic)
		return extImpPubMatic.AdSlot, "", false, params, err
	}

	kgpv := "" //regex match
	hash := ""
	var err error
	var paramMap map[string]interface{}
	selectSlot := ""
	isRegexSlot := false

	isRegexKGP := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN] == models.REGEX_KGP

	// simple+regex key match
	for _, slot := range slots {
		matchedSlot, matchedPattern := GetMatchingSlot(rctx, cache, slot, slotMap, slotMappingInfo, isRegexKGP, partnerID)
		if matchedSlot != "" {
			selectSlot = matchedSlot
			extImpPubMatic.AdSlot = matchedSlot

			if matchedPattern != "" {
				isRegexSlot = true
				kgpv = matchedPattern
				// imp.TagID = hash

				// TODO: handle kgpv case sensitivity in hashvaluemap
				if slotMappingInfo.HashValueMap != nil {
					if v, ok := slotMappingInfo.HashValueMap[kgpv]; ok {
						extImpPubMatic.AdSlot = v
						imp.TagID = hash // TODO, make imp pointer. But do other bidders accept hash as TagID?
					}
				}
			}

			break
		}
	}

	//overwrite
	if paramMap != nil {
		// use owSlotName to addres case insensitive slotname. EX slot= "/43743431/DMDEMO@300x250" and owSlotName="/43743431/DMDemo@300x250"
		if v, ok := paramMap[models.KEY_OW_SLOT_NAME]; ok {
			if owSlotName, ok := v.(string); ok {
				extImpPubMatic.AdSlot = owSlotName
			}
		}

		//Update slot key for PubMatic secondary flow
		if v, ok := paramMap[models.KEY_SLOT_NAME]; ok {
			if secondarySlotName, ok := v.(string); ok {
				extImpPubMatic.AdSlot = secondarySlotName
			}
		}
	}

	//last resort
	if extImpPubMatic.AdSlot == "" {
		var div string
		if impExt.Wrapper != nil {
			div = impExt.Wrapper.Div
		}
		unmappedKPG := getDefaultMappingKGP(rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN])
		extImpPubMatic.AdSlot = GenerateSlotName(0, 0, unmappedKPG, imp.TagID, div, rctx.Source)
		if len(slots) != 0 { // reuse this field for wt and wl in combination with isRegex
			kgpv = slots[0]
		}
	}

	params, err := json.Marshal(extImpPubMatic)
	return selectSlot, kgpv, isRegexSlot, params, err
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
