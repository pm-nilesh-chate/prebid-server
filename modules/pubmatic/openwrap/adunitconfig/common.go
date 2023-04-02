package adunitconfig

import (
	"encoding/json"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func selectSlot(rCtx models.RequestCtx, h, w int64, tagid, div, source string) (slotAdUnitConfig *adunitconfig.AdConfig, slotName string, isRegex bool, matchedRegex string) {
	slotName = bidderparams.GenerateSlotName(h, w, rCtx.AdUnitConfig.ConfigPattern, tagid, div, rCtx.Source)

	if slotAdUnitConfig, ok := rCtx.AdUnitConfig.Config[slotName]; ok {
		return slotAdUnitConfig, slotName, false, ""
	} else if rCtx.AdUnitConfig.Regex {
		if matchedRegex = getRegexMatch(rCtx, slotName); matchedRegex != "" {
			return rCtx.AdUnitConfig.Config[matchedRegex], slotName, true, matchedRegex
		}
	}

	return nil, "", false, ""
}

/*GetClientConfigForMediaType function fetches the client config data from the ad unit config JSON for the given media type*/
func GetClientConfigForMediaType(rctx models.RequestCtx, impID string, adUnitCfgMap *adunitconfig.AdUnitConfig, mediaType string) json.RawMessage {
	if adUnitCfgMap == nil || adUnitCfgMap.Config == nil {
		return nil
	}

	impData, ok := rctx.ImpBidCtx[impID]
	if !ok {
		return nil
	}

	if mediaType == "banner" {
		if impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig != nil && impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner != nil &&
			impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Config != nil {
			return impData.BannerAdUnitCtx.AppliedSlotAdUnitConfig.Banner.Config.ClientConfig
		}
	} else if mediaType == "video" {
		if impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig != nil && impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
			impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config != nil {
			return impData.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Video.Config.ClientConfig
		}
	}

	return nil
}
