package bidderparams

import (
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
)

func PrepareAdapterParamsV25(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) (string, []byte, errorcodes.IError) {
	partnerConfig, ok := rctx.PartnerConfigMap[partnerID]
	if !ok {
		return "", nil, errorcodes.ErrBidderParamsValidationError
	}

	isRegex := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN] == models.REGEX_KGP
	slots, slotMap, slotMappingInfo, hw := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)

	for i, slot := range slots {
		matchedSlot, matchedPattern := GetMatchingSlot(rctx, cache, slot, slotMap, slotMappingInfo, isRegex, partnerID)
		if matchedSlot == "" {
			continue
		}

		slotMappingObj, ok := slotMap[strings.ToLower(matchedSlot)]
		if !ok {
			slotMappingObj, _ = slotMap[strings.ToLower(matchedPattern)]
		}
		bidderParams := slotMappingObj.SlotMappings
		for key, value := range partnerConfig {
			if !ignoreKeys[key] {
				bidderParams[key] = value
			}
		}

		h := hw[i][0]
		w := hw[i][1]
		params, err := adapters.PrepareBidParamJSONForPartner(&w, &h, bidderParams, slot, partnerConfig[models.PREBID_PARTNER_NAME], partnerConfig[models.BidderCode], &impExt)
		if err != nil || params == nil {
			continue
		}
		return slot, params, nil
	}

	return "", nil, nil
}
