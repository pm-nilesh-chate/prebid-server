package openwrap

import (
	"encoding/json"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/adcom1"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// TODO use this
func getMatchedSlotName(rCtx models.RequestCtx, imp openrtb2.Imp, impExt models.ImpExtension) (slotAdUnitConfig *adunitconfig.AdConfig, isRegex bool) {
	div := ""
	height := imp.Video.H
	width := imp.Video.W
	tagID := imp.TagID

	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	slotName := bidderparams.GenerateSlotName(height, width, rCtx.AdUnitConfig.ConfigPattern, tagID, div, rCtx.Source)

	var ok bool
	slotAdUnitConfig, ok = rCtx.AdUnitConfig.Config[slotName]
	if ok {
		return
	}

	// for slot, adUnitConfig := range rCtx.AdUnitConfig.Config {

	// }

	return
}

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

func updateVideoObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, impExt models.ImpExtension, connectionType *adcom1.ConnectionType) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	if (rCtx.Platform != models.PLATFORM_APP) && (rCtx.Platform != models.PLATFORM_VIDEO) && (rCtx.Platform != models.PLATFORM_DISPLAY) {
		return
	}

	if imp.Video == nil || rCtx.AdUnitConfig == nil || len(rCtx.AdUnitConfig.Config) == 0 {
		return
	}

	allowedConnectionTypes := getDefaultAllowedConnectionTypes(rCtx.AdUnitConfig)

	usingDefaultConfig := false
	var videoConfig *adunitconfig.Video

	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil {
		usingDefaultConfig = true
		videoConfig = defaultAdUnitConfig.Video

		if defaultAdUnitConfig.Video != nil && defaultAdUnitConfig.Video.Enabled != nil && !*defaultAdUnitConfig.Video.Enabled {
			imp.Video = nil
			return
		}
	}

	div := ""
	height := imp.Video.H
	width := imp.Video.W
	tagID := imp.TagID

	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	slotName := bidderparams.GenerateSlotName(height, width, rCtx.AdUnitConfig.ConfigPattern, tagID, div, rCtx.Source)
	slotAdUnitConfig, ok := rCtx.AdUnitConfig.Config[slotName]

	if ok && slotAdUnitConfig != nil && slotAdUnitConfig.Exp != nil {
		imp.Exp = int64(*slotAdUnitConfig.Exp)
	}

	getFinalSlotAdUnitConfig(slotAdUnitConfig, defaultAdUnitConfig)

	if slotAdUnitConfig.Video != nil {
		usingDefaultConfig = false
		videoConfig = slotAdUnitConfig.Video
	}

	if videoConfig == nil {
		return
	}

	//check if video is disabled, if yes then remove video from imp object
	if videoConfig.Enabled != nil && !*videoConfig.Enabled {
		imp.Video = nil
		return
	}

	if videoConfig.Config == nil {
		return
	}
	configObjInVideoConfig := videoConfig.Config

	// disable video if connection type is not present in allowed connection types from config
	if connectionType != nil {
		//check connection type in slot config
		if !usingDefaultConfig && len(configObjInVideoConfig.ConnectionType) > 0 {
			allowedConnectionTypes = configObjInVideoConfig.ConnectionType
		}

		if allowedConnectionTypes != nil && !CheckValuePresentInArray(allowedConnectionTypes, int(*connectionType)) {
			imp.Video = nil
			return
		}
	}

	if imp.BidFloor == 0 && slotAdUnitConfig.BidFloor != nil {
		imp.BidFloor = *slotAdUnitConfig.BidFloor
	}

	if len(imp.BidFloorCur) == 0 && slotAdUnitConfig.BidFloorCur != nil {
		imp.BidFloorCur = *slotAdUnitConfig.BidFloorCur
	}

	if len(imp.Video.MIMEs) == 0 {
		imp.Video.MIMEs = configObjInVideoConfig.MIMEs
	}

	if imp.Video.MinDuration == 0 {
		imp.Video.MinDuration = configObjInVideoConfig.MinDuration
	}

	if imp.Video.MaxDuration == 0 {
		imp.Video.MaxDuration = configObjInVideoConfig.MaxDuration
	}

	if imp.Video.Skip == nil {
		imp.Video.Skip = configObjInVideoConfig.Skip
	}

	if imp.Video.SkipMin == 0 {
		imp.Video.SkipMin = configObjInVideoConfig.SkipMin
	}

	if imp.Video.SkipAfter == 0 {
		imp.Video.SkipAfter = configObjInVideoConfig.SkipAfter
	}

	if len(configObjInVideoConfig.BAttr) == 0 {
		imp.Video.BAttr = configObjInVideoConfig.BAttr
	}

	if imp.Video.MinBitRate == 0 {
		imp.Video.MinBitRate = configObjInVideoConfig.MinBitRate
	}

	if imp.Video.MaxBitRate == 0 {
		imp.Video.MaxBitRate = configObjInVideoConfig.MaxBitRate
	}

	if imp.Video.MaxExtended == 0 {
		imp.Video.MaxExtended = configObjInVideoConfig.MaxExtended
	}

	if imp.Video.StartDelay == nil {
		imp.Video.StartDelay = configObjInVideoConfig.StartDelay
	}

	if imp.Video.Placement == 0 {
		imp.Video.Placement = configObjInVideoConfig.Placement
	}

	if imp.Video.Linearity == 0 {
		imp.Video.Linearity = configObjInVideoConfig.Linearity
	}

	if imp.Video.Protocol == 0 {
		imp.Video.Protocol = configObjInVideoConfig.Protocol
	}

	if len(configObjInVideoConfig.Protocols) == 0 {
		imp.Video.Protocols = configObjInVideoConfig.Protocols
	}

	if imp.Video.W == 0 {
		imp.Video.W = configObjInVideoConfig.W
	}

	if imp.Video.H == 0 {
		imp.Video.H = configObjInVideoConfig.H
	}

	if imp.Video.Sequence == 0 {
		imp.Video.Sequence = configObjInVideoConfig.Sequence
	}

	if imp.Video.BoxingAllowed == 0 {
		imp.Video.BoxingAllowed = configObjInVideoConfig.BoxingAllowed
	}

	if imp.Video.PlaybackMethod == nil && len(configObjInVideoConfig.PlaybackMethod) > 0 {
		imp.Video.PlaybackMethod = configObjInVideoConfig.PlaybackMethod
	}

	if imp.Video.PlaybackEnd == 0 {
		imp.Video.PlaybackEnd = configObjInVideoConfig.PlaybackEnd
	}

	if imp.Video.Delivery == nil {
		imp.Video.Delivery = configObjInVideoConfig.Delivery
	}

	if imp.Video.Pos == nil {
		imp.Video.Pos = configObjInVideoConfig.Pos
	}

	if len(configObjInVideoConfig.API) > 0 {
		imp.Video.API = configObjInVideoConfig.API
	}

	if len(configObjInVideoConfig.CompanionType) > 0 {
		imp.Video.CompanionType = configObjInVideoConfig.CompanionType
	}

	if imp.Video.CompanionAd == nil {
		imp.Video.CompanionAd = configObjInVideoConfig.CompanionAd
	}
}

func CheckValuePresentInArray(intArray []int, value int) bool {
	for _, eachVal := range intArray {
		if eachVal == value {
			return true
		}
	}
	return false
}

// update slotConfig with final AdUnit config to apply with
func getFinalSlotAdUnitConfig(slotConfig, defaultConfig *adunitconfig.AdConfig) {
	if slotConfig == nil && defaultConfig == nil {
		return
	}

	if slotConfig == nil {
		slotConfig = defaultConfig
		return
	}

	if (slotConfig.BidFloor == nil || *slotConfig.BidFloor == 0.0) && slotConfig.BidFloor != nil {
		slotConfig.BidFloor = defaultConfig.BidFloor

		slotConfig.BidFloorCur = func() *string { s := "USD"; return &s }()
		if defaultConfig.BidFloorCur != nil {
			slotConfig.BidFloorCur = defaultConfig.BidFloorCur
		}
	}
}

func updateBannerObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, impExt models.ImpExtension) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	if imp.Banner == nil || rCtx.AdUnitConfig == nil || len(rCtx.AdUnitConfig.Config) == 0 {
		return
	}

	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil {
		if defaultAdUnitConfig.Banner != nil && defaultAdUnitConfig.Banner.Enabled != nil &&
			defaultAdUnitConfig.Banner.Enabled != nil && !*defaultAdUnitConfig.Banner.Enabled {
			imp.Banner = nil
			return
		}
	}

	div := ""
	tagID := imp.TagID

	var height, width int64
	if imp.Banner.H != nil {
		height = *imp.Banner.H
	}
	if imp.Banner.H != nil {
		width = *imp.Banner.W
	}

	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	slotName := bidderparams.GenerateSlotName(height, width, rCtx.AdUnitConfig.ConfigPattern, tagID, div, rCtx.Source)
	slotAdUnitConfig, ok := rCtx.AdUnitConfig.Config[slotName]
	if ok && slotAdUnitConfig != nil && slotAdUnitConfig.Exp != nil {
		imp.Exp = int64(*slotAdUnitConfig.Exp)
	}

	getFinalSlotAdUnitConfig(slotAdUnitConfig, defaultAdUnitConfig)

	if imp.BidFloor == 0 && slotAdUnitConfig.BidFloor != nil {
		imp.BidFloor = *slotAdUnitConfig.BidFloor
	}

	if len(imp.BidFloorCur) == 0 && slotAdUnitConfig.BidFloorCur != nil {
		imp.BidFloorCur = *slotAdUnitConfig.BidFloorCur
	}

	if slotAdUnitConfig == nil || slotAdUnitConfig.Banner == nil {
		return
	}

	//check if video is disabled, if yes then remove video from imp object
	if slotAdUnitConfig.Banner.Enabled != nil && !*slotAdUnitConfig.Banner.Enabled {
		imp.Video = nil
		return
	}
}
