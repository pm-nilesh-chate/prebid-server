package errorcodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		code IError
		want string
	}{
		{name: "ErrSuccess", code: ErrSuccess, want: ""},
		{name: "ErrGADSUnmappedSlot", code: ErrGADSUnmappedSlot, want: "Slot not mapped"},
		{name: "ErrGADSMissingConfig", code: ErrGADSMissingConfig, want: "Missing Configuration"},
		{name: "ErrTimeout", code: ErrTimeout, want: "Timeout Error"},
		{name: "ErrNoBidPrebid", code: ErrNoBidPrebid, want: "No Bid"},
		{name: "ErrPartnerTimeout", code: ErrPartnerTimeout, want: "Partner Timed out"},
		{name: "ErrInvalidConfiguration", code: ErrInvalidConfiguration, want: "Invalid Configuration"},
		{name: "ErrNoGDPRConsent", code: ErrNoGDPRConsent, want: "No Consent Present"},
		{name: "ErrInvalidCreative", code: ErrInvalidCreative, want: "Invalid Creative"},
		{name: "ErrCachePutFailed", code: ErrCachePutFailed, want: "Cache PUT Failed"},
		{name: "ErrInvalidParameter", code: ErrInvalidParameter, want: "Invalid Parameters"},
		{name: "ErrAllPartnerThrottled", code: ErrAllPartnerThrottled, want: "All partners throttled"},
		{name: "ErrPartnerThrottled", code: ErrPartnerThrottled, want: "Partner throttled"},
		{name: "ErrBannerVideoDisabled", code: ErrBannerVideoDisabled, want: "Banner/Video disabled through config"},
		{name: "ErrPartnerContextDeadlineExceeded", code: ErrPartnerContextDeadlineExceeded, want: "context deadline exceeded"},
		{name: "ErrPrebidDefaultTimeout", code: ErrPrebidDefaultTimeout, want: "Timed out"},
		{name: "ErrInvalidHTTPRequestMethod", code: ErrInvalidHTTPRequestMethod, want: "Invalid Request Method"},
		{name: "ErrInvalidImpression", code: ErrInvalidImpression, want: "No Valid Impression Found"},
		{name: "ErrBadRequest", code: ErrBadRequest, want: "BadRequest"},
		{name: "ErrPrebidUnknownError", code: ErrPrebidUnknownError, want: "Unknown error received from Prebid"},
		{name: "ErrInvalidAdUnitConfig", code: ErrInvalidAdUnitConfig, want: "Invalid adUnit config uploaded"},
		{name: "ErrInvalidDealTierExt", code: ErrInvalidDealTierExt, want: "Invalid deal tier info"},
		{name: "ErrInternalApp", code: ErrInternalApp, want: "API Error"},
		{name: "ErrMissingRequestID", code: ErrMissingRequestID, want: "Missing Request ID"},
		{name: "ErrMissingImpressions", code: ErrMissingImpressions, want: "Missing Impressions"},
		{name: "ErrMissingSiteApp", code: ErrMissingSiteApp, want: "Missing Site/App Object"},
		{name: "ErrSiteAppBothPresent", code: ErrSiteAppBothPresent, want: "Site App Both Present"},
		{name: "ErrMissingPublisherID", code: ErrMissingPublisherID, want: "Missing Publisher ID"},
		{name: "ErrMissingTagID", code: ErrMissingTagID, want: "Missing Tag ID"},
		{name: "ErrMissingImpressionID", code: ErrMissingImpressionID, want: "Missing Impression ID"},
		{name: "ErrMissingAdType", code: ErrMissingAdType, want: "Missing Ad Type(Banner/Video)"},
		{name: "ErrMissingAdSize", code: ErrMissingAdSize, want: "Missing Ad Sizes"},
		{name: "ErrMissingMIME", code: ErrMissingMIME, want: "Missing MIMEs"},
		{name: "ErrMissingNativeRequest", code: ErrMissingNativeRequest, want: "Missing Native Request Payload"},
		{name: "ErrMissingVideoObject", code: ErrMissingVideoObject, want: "Missing Video Object"},
		{name: "ErrInvalidRequestExtension", code: ErrInvalidRequestExtension, want: "Invalid Request Extension"},
		{name: "ErrInvalidCrossPodAdvertiserExclusionPercent", code: ErrInvalidCrossPodAdvertiserExclusionPercent, want: "request.ext.adpod.crosspodexcladv must be a number between 0 and 100"},
		{name: "ErrInvalidCrossPodIABCategoryExclusionPercent", code: ErrInvalidCrossPodIABCategoryExclusionPercent, want: "request.ext.adpod.crosspodexcliabcat must be a number between 0 and 100"},
		{name: "ErrInvalidIABCategoryExclusionWindow", code: ErrInvalidIABCategoryExclusionWindow, want: "request.ext.adpod.excliabcatwindow must be postive number"},
		{name: "ErrInvalidAdvertiserExclusionWindow", code: ErrInvalidAdvertiserExclusionWindow, want: "request.ext.adpod.excladvwindow must be postive number"},
		{name: "ErrInvalidVideoExtension", code: ErrInvalidVideoExtension, want: "Invalid Video Extensions"},
		{name: "ErrInvalidAdPodOffset", code: ErrInvalidAdPodOffset, want: "imp.video.ext.offset must be postive number"},
		{name: "ErrInvalidMinAds", code: ErrInvalidMinAds, want: "%key%.adpod.minads must be positive number"},
		{name: "ErrInvalidMaxAds", code: ErrInvalidMaxAds, want: "%key%.adpod.maxads must be positive number"},
		{name: "ErrInvalidMinDuration", code: ErrInvalidMinDuration, want: "%key%.adpod.adminduration must be positive number"},
		{name: "ErrInvalidMaxDuration", code: ErrInvalidMaxDuration, want: "%key%.adpod.admaxduration must be positive number"},
		{name: "ErrInvalidAdvertiserExclusionPercent", code: ErrInvalidAdvertiserExclusionPercent, want: "%key%.adpod.excladv must be number between 0 and 100"},
		{name: "ErrInvalidIABCategoryExclusionPercent", code: ErrInvalidIABCategoryExclusionPercent, want: "%key%.adpod.excliabcat must be number between 0 and 100"},
		{name: "ErrInvalidMinMaxAds", code: ErrInvalidMinMaxAds, want: "%key%.adpod.minads must be less than %key%.adpod.maxads"},
		{name: "ErrInvalidMinMaxDuration", code: ErrInvalidMinMaxDuration, want: "%key%.adpod.adminduration must be less than %key%.adpod.admaxduration"},
		{name: "ErrInvalidVideoMinDuration", code: ErrInvalidVideoMinDuration, want: "imp.video.minduration must be positive number"},
		{name: "ErrInvalidVideoMaxDuration", code: ErrInvalidVideoMaxDuration, want: "imp.video.maxduration must be positive number"},
		{name: "ErrInvalidVideoDurations", code: ErrInvalidVideoDurations, want: "imp.video.minduration must be less than imp.video.maxduration"},
		{name: "ErrInvalidMinMaxDurationRange", code: ErrInvalidMinMaxDurationRange, want: "adpod duration checks for adminduration,admaxduration,minads,maxads are not in video minduration and maxduration duration range"},
		{name: "ErrInvalidPublisherID", code: ErrInvalidPublisherID, want: "Invalid Publisher ID"},
		{name: "ErrMissingProfileID", code: ErrMissingProfileID, want: "Missing Profile ID"},
		{name: "ErrCacheWarmup", code: ErrCacheWarmup, want: "Cache Warmup"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.code.Error())
		})
	}
}

func TestNewError(t *testing.T) {
	code := NewError(123, "Message")
	assert.Equal(t, 123, code.Code())
	assert.Equal(t, "Message", code.Error())
}
