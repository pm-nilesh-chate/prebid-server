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

	slots, slotMap, _, hw := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)
	for i, slot := range slots {
		slotMappingObj, ok := slotMap[strings.ToLower(slot)]
		if !ok {
			continue
		}

		bidderParams := slotMappingObj.SlotMappings
		for key, value := range partnerConfig {
			if !ignoreKeys[key] {
				bidderParams[key] = value
			}
		}

		h := hw[i][0]
		w := hw[i][1]
		params, err := adapters.PrepareBidParamJSONForPartner(&w, &h, slotMappingObj.SlotMappings, slot, partnerConfig[models.PREBID_PARTNER_NAME], partnerConfig[models.BidderCode], &impExt)
		if err != nil || params == nil {
			continue
		}
		return slot, params, nil
	}

	return "", nil, nil
}
