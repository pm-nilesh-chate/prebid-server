package openwrap

import "github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

func (m OpenWrap) getProfileData(rCtx models.RequestCtx) (map[int]map[string]string, error) {
	if rCtx.IsTestRequest {
		// NYC: can this be clubbed with profileid=0
		return getTestModePartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID, rCtx.Platform), nil
	} else if rCtx.ProfileID == 0 {
		return getDefaultPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID), nil
	} else {
		return m.cache.GetPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	}

	return nil, nil
}
