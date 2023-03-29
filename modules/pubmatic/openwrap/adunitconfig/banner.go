package adunitconfig

import (
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func UpdateBannerObjectWithAdunitConfig(rCtx models.RequestCtx, imp openrtb2.Imp, div string) (adUnitCtx models.AdUnitCtx) {
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
		if defaultAdUnitConfig.Banner != nil && defaultAdUnitConfig.Banner.Enabled != nil && !*defaultAdUnitConfig.Banner.Enabled {
			f := false
			adUnitCtx.AppliedSlotAdUnitConfig = &adunitconfig.AdConfig{Banner: &adunitconfig.Banner{Enabled: &f}}
			return
		}
	}

	var height, width int64
	if imp.Banner.H != nil {
		height = *imp.Banner.H
	}
	if imp.Banner.H != nil {
		width = *imp.Banner.W
	}

	adUnitCtx.SelectedSlotAdUnitConfig, adUnitCtx.MatchedSlot, adUnitCtx.IsRegex, adUnitCtx.MatchedRegex = selectSlot(rCtx, height, width, imp.TagID, div, rCtx.Source)

	adUnitCtx.AppliedSlotAdUnitConfig = getFinalSlotAdUnitConfig(adUnitCtx.SelectedSlotAdUnitConfig, defaultAdUnitConfig)

	return
}
