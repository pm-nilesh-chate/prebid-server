package adunitconfig

import (
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/adcom1"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func UpdateVideoObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, div string, connectionType *adcom1.ConnectionType) (adUnitCtx models.AdUnitCtx) {
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

	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil {
		adUnitCtx.UsingDefaultConfig = true

		if defaultAdUnitConfig.Video != nil && defaultAdUnitConfig.Video.Enabled != nil && !*defaultAdUnitConfig.Video.Enabled {
			f := false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Video: &adunitconfig.Video{Enabled: &f}}
			return
		}
	}

	height := int64(imp.Video.H)
	width := int64(imp.Video.W)

	adUnitCtx.SelectedSlotAdUnitConfig, adUnitCtx.MatchedSlot, adUnitCtx.IsRegex, adUnitCtx.MatchedRegex = selectSlot(rCtx, height, width, imp.TagID, div, rCtx.Source)
	if adUnitCtx.SelectedSlotAdUnitConfig != nil && adUnitCtx.SelectedSlotAdUnitConfig.Video != nil {
		adUnitCtx.UsingDefaultConfig = false
	}

	adUnitCtx.AppliedSlotAdUnitConfig = getFinalSlotAdUnitConfig(adUnitCtx.SelectedSlotAdUnitConfig, defaultAdUnitConfig)
	if adUnitCtx.AppliedSlotAdUnitConfig == nil {
		return
	}

	adUnitCtx.AllowedConnectionTypes = getDefaultAllowedConnectionTypes(rCtx.AdUnitConfig)

	// updateAllowedConnectionTypes := !adUnitCtx.UsingDefaultConfig
	// if adUnitCtx.AppliedSlotAdUnitConfig != nil && adUnitCtx.AppliedSlotAdUnitConfig.Video != nil &&
	// 	adUnitCtx.AppliedSlotAdUnitConfig.Video.Config != nil && len(adUnitCtx.AppliedSlotAdUnitConfig.Video.Config.ConnectionType) != 0 {
	// 	updateAllowedConnectionTypes = updateAllowedConnectionTypes && true
	// }

	// // disable video if connection type is not present in allowed connection types from config
	// if connectionType != nil {
	// 	//check connection type in slot config
	// 	if updateAllowedConnectionTypes {
	// 		adUnitCtx.AllowedConnectionTypes = configObjInVideoConfig.ConnectionType
	// 	}

	// 	if allowedConnectionTypes != nil && !checkValuePresentInArray(allowedConnectionTypes, int(*connectionType)) {
	// 		f := false
	// 		adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Video: &adunitconfig.Video{Enabled: &f}}
	// 		return
	// 	}
	// }

	return
}
