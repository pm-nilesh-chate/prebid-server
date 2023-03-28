package openwrap

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// setContentObjectTransparencyObject from request or AdUnit Object
// setContentObjectTransparencyObject from request or AdUnit Object
func setContentTransparencyObject(rctx models.RequestCtx, reqExt models.RequestExt, impID string, adUnitConfigMap *adunitconfig.AdUnitConfig) (prebidTransparency *openrtb_ext.TransparencyExt) {
	if reqExt.Prebid.Transparency != nil {
		return
	}

	if adUnitConfigMap == nil {
		return
	}

	impData, ok := rctx.ImpBidCtx[impID]
	if !ok {
		return
	}

	var contentMappings map[string]openrtb_ext.TransparencyRule

	if v, ok := adUnitConfigMap.Config[impData.TagID]; ok && v.Transparency != nil {
		contentMappings = v.Transparency.Content.Mappings
	} else if v, ok := adUnitConfigMap.Config[models.AdunitConfigDefaultKey]; ok && v.Transparency != nil {
		contentMappings = v.Transparency.Content.Mappings
	}

	if len(contentMappings) == 0 {
		return
	}

	prebidTransparency = &openrtb_ext.TransparencyExt{
		Content: map[string]openrtb_ext.TransparencyRule{},
	}

	for _, partnerConfig := range rctx.PartnerConfigMap {
		_, ok := rctx.AdapterThrottleMap[partnerConfig[models.BidderCode]]
		if ok || partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		for _, rule := range getRules(rctx.Source, partnerConfig[models.BidderCode]) {
			if transparencyRule, ok := contentMappings[rule]; ok {
				prebidTransparency.Content[partnerConfig[models.BidderCode]] = transparencyRule
				break
			}
		}
	}

	// NYC: This result overwrites previous o/p. Update code to append bidders
	return prebidTransparency
}

func getRules(source, bidder string) []string {
	return []string{source + "|" + bidder, "*|" + bidder, source + "|*", "*|*"}
}
