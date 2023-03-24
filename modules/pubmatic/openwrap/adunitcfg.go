package openwrap

import (
	"encoding/json"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// getEnabledValue reads default "enabled" value for banner and video in the adunit config
func getDefaultEnabledValueForMediaType(adUnitConfigMap *adunitconfig.AdUnitConfig, mediaType string) bool {
	enabled := true
	if adUnitConfigMap == nil || adUnitConfigMap.Config == nil {
		return enabled
	}

	if v, ok := adUnitConfigMap.Config[models.AdunitConfigDefaultKey]; ok {
		if mediaType == models.Banner && v.Banner.Enabled != nil {
			return *v.Banner.Enabled
		}

		if mediaType == models.Video && v.Video.Enabled != nil {
			return *v.Video.Enabled
		}
	}
	return enabled
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

	if cfg, ok := adUnitCfgMap.Config[impData.MatchedSlot]; ok {
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
