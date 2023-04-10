package openwrap

import (
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) getProfileData(rCtx models.RequestCtx, bidRequest openrtb2.BidRequest) (map[int]map[string]string, error) {
	if rCtx.IsTestRequest {
		//get platform from request, since test mode can be enabled for display and app platform only
		var platform string
		if bidRequest.Site != nil {
			platform = models.PLATFORM_DISPLAY
		} else if bidRequest.App != nil {
			platform = models.PLATFORM_APP
		}

		// NYC: can this be clubbed with profileid=0
		return getTestModePartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID, platform), nil
	} else if rCtx.ProfileID == 0 {
		return getDefaultPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID), nil
	}

	return m.cache.GetPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
}
