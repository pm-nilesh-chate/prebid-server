package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
)

// IsCTVAPIRequest will return true if reqAPI is from CTV EndPoint
func IsCTVAPIRequest(api string) bool {
	// return reqAPI == OpenRTB_VIDEO_JSON_API || reqAPI == OpenRTB_VIDEO_VAST_API || reqAPI == constant.OpenRTB_VIDEO_OPENRTB_API
	return api != "/2.5"
	// NYC_TODO: fix this temporary change
}

func GetRequestExt(ext []byte) (ExtRequest, error) {
	extRequest := ExtRequest{}

	err := json.Unmarshal(ext, &extRequest)
	if err != nil {
		return extRequest, fmt.Errorf("failed to decode request.ext : %v", err)
	}

	return extRequest, nil
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

func GetAccountID(request []byte) (int, error) {
	pubid := 0
	pubIdStr, _, err := searchAccountId(request)
	if err != nil {
		return pubid, fmt.Errorf("failed to get publisher id : %v", err)
	}

	pubid, err = strconv.Atoi(pubIdStr)
	if err != nil {
		return pubid, fmt.Errorf("invalid publisher id : %v", err)
	}

	return pubid, nil
}

// NYC_TODO: Export searchAccountId() from PBS-Core for reuse

var accountIdSearchPath = [...]struct {
	isApp bool
	key   []string
}{
	{true, []string{"app", "publisher", "ext", "prebid", "parentAccount"}},
	{true, []string{"app", "publisher", "id"}},
	{false, []string{"site", "publisher", "ext", "prebid", "parentAccount"}},
	{false, []string{"site", "publisher", "id"}},
}

func searchAccountId(request []byte) (string, bool, error) {
	for _, path := range accountIdSearchPath {
		accountId, exists, err := getStringValueFromRequest(request, path.key)
		if err != nil {
			return "", path.isApp, err
		}
		if exists {
			return accountId, path.isApp, nil
		}
	}
	return "", false, nil
}

func getStringValueFromRequest(request []byte, key []string) (string, bool, error) {
	val, dataType, _, err := jsonparser.Get(request, key...)
	if dataType == jsonparser.NotExist {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if dataType != jsonparser.String {
		return "", true, fmt.Errorf("%s must be a string", strings.Join(key, "."))
	}
	return string(val), true, nil
}
