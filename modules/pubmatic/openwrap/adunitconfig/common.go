package adunitconfig

import (
	"encoding/json"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func selectSlot(rCtx models.RequestCtx, h, w int64, kgp, tagid, div string) (slotAdUnitConfig *adunitconfig.AdConfig, slotName string, isRegex bool, matchedRegex string) {
	slotName = bidderparams.GenerateSlotName(h, w, rCtx.AdUnitConfig.ConfigPattern, tagid, div, rCtx.Source)

	if slotAdUnitConfig, ok := rCtx.AdUnitConfig.Config[slotName]; ok {
		return slotAdUnitConfig, slotName, false, ""
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

	if cfg, ok := adUnitCfgMap.Config[impData.TagID]; ok {
		if mediaType == models.AdunitConfigSlotBannerKey {
			if cfg.Banner != nil && cfg.Banner.Config != nil {
				return cfg.Banner.Config.ClientConfig
			}
		} else if mediaType == models.AdunitConfigSlotVideoKey {
			if cfg.Video != nil && cfg.Video.Config != nil {
				return cfg.Video.Config.ClientConfig
			}
		}
	}

	if cfg, ok := adUnitCfgMap.Config[models.AdunitConfigDefaultKey]; ok {
		if mediaType == models.AdunitConfigSlotBannerKey {
			if cfg.Banner != nil && cfg.Banner.Config != nil {
				return cfg.Banner.Config.ClientConfig
			}
		} else if mediaType == models.AdunitConfigSlotVideoKey {
			if cfg.Video != nil && cfg.Video.Config != nil {
				return cfg.Video.Config.ClientConfig
			}
		}
	}

	return nil
}
