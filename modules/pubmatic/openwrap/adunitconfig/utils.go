package adunitconfig

import (
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// TODO use this
func GetMatchedSlotName(rCtx models.RequestCtx, imp openrtb2.Imp, impExt models.ImpExtension) (slotAdUnitConfig *adunitconfig.AdConfig, isRegex bool) {
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

func getDefaultAllowedConnectionTypes(adUnitConfigMap *adunitconfig.AdUnitConfig) []int {
	if adUnitConfigMap == nil {
		return nil
	}

	if v, ok := adUnitConfigMap.Config[models.AdunitConfigDefaultKey]; ok && v.Video != nil && v.Video.Config != nil && len(v.Video.Config.CompanionType) != 0 {
		return v.Video.Config.ConnectionType
	}

	return nil
}

func checkValuePresentInArray(intArray []int, value int) bool {
	for _, eachVal := range intArray {
		if eachVal == value {
			return true
		}
	}
	return false
}

// update slotConfig with final AdUnit config to apply with
func getFinalSlotAdUnitConfig(slotConfig, defaultConfig *adunitconfig.AdConfig) *adunitconfig.AdConfig {
	// nothing available
	if slotConfig == nil && defaultConfig == nil {
		return nil
	}

	// only default available
	if slotConfig == nil {
		return defaultConfig
	}

	// only slot available
	if defaultConfig == nil {
		return slotConfig
	}

	// both available, merge both with priority to slot

	if (slotConfig.BidFloor == nil || *slotConfig.BidFloor == 0.0) && defaultConfig.BidFloor != nil {
		slotConfig.BidFloor = defaultConfig.BidFloor

		slotConfig.BidFloorCur = func() *string { s := "USD"; return &s }()
		if defaultConfig.BidFloorCur != nil {
			slotConfig.BidFloorCur = defaultConfig.BidFloorCur
		}
	}

	if slotConfig.Banner == nil {
		slotConfig.Banner = defaultConfig.Banner
	}

	if slotConfig.Video == nil {
		slotConfig.Video = defaultConfig.Video
	}

	if slotConfig.Floors == nil {
		slotConfig.Floors = defaultConfig.Floors
	}

	return slotConfig
}
