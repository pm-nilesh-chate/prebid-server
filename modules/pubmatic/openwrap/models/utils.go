package models

import (
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
)

// IsCTVAPIRequest will return true if reqAPI is from CTV EndPoint
func IsCTVAPIRequest(api string) bool {
	// return reqAPI == OpenRTB_VIDEO_JSON_API || reqAPI == OpenRTB_VIDEO_VAST_API || reqAPI == constant.OpenRTB_VIDEO_OPENRTB_API
	return api != "/2.5"
	// NYC_TODO: fix this temporary change
}

func GetWrapperExt(request []byte) (RequestExtWrapper, error) {
	extWrapper := RequestExtWrapper{}

	// NYC_TODO: if /2.5 redirect check ext.wrapper else check ext.prebid.bidderparams.pubmatic.wrapper
	extWrapperBytes, _, _, err := jsonparser.Get(request, "ext", "wrapper")
	if err != nil {
		return extWrapper, fmt.Errorf("request.ext.wrapper not found: %v", err)
	}

	err = json.Unmarshal(extWrapperBytes, &extWrapper)
	if err != nil {
		return extWrapper, fmt.Errorf("failed to decode request.ext.wrapper : %v", err)
	}

	return extWrapper, nil
}

func GetTest(request []byte) (int64, error) {
	test, err := jsonparser.GetInt(request, "test")
	if err != nil {
		return test, fmt.Errorf("request.test not found: %v", err)
	}
	return test, nil
}
