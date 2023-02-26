package bidderparams

import (
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/request"
)

func PrepareAdapterParamsV25(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt request.ImpExtension, partnerID int) ([]byte, errorcodes.IError) {
	slots, slotMap, _ := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)
	for _, slot := range slots {
		slotMappingObj, ok := slotMap[strings.ToLower(slot)]
		if !ok {
			continue
		}

		params, err := adapters.PrepareBidParamJSONForPartner(imp.Banner.W, imp.Banner.H, slotMappingObj.SlotMappings, slot, rctx.PartnerConfigMap[partnerID][models.PREBID_PARTNER_NAME], rctx.PartnerConfigMap[partnerID][models.BidderCode], &impExt)
		if err != nil || params == nil {
			continue
		}
		return params, nil
	}

	return nil, nil
}
