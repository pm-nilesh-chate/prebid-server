package models

// IsCTVAPIRequest will return true if reqAPI is from CTV EndPoint
func IsCTVAPIRequest(api string) bool {
	// return reqAPI == OpenRTB_VIDEO_JSON_API || reqAPI == OpenRTB_VIDEO_VAST_API || reqAPI == constant.OpenRTB_VIDEO_OPENRTB_API
	return api != "/2.5"
	// NYC_TODO: fix this temporary change
}
