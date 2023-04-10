package openrtb2

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/analytics/openwrap"
)

func getLogInfo(requestExt, responseExt []byte, ao *analytics.AuctionObject) []byte {
	isLogInfo, err := jsonparser.GetBoolean(requestExt, "wrapper", "loginfo")
	if err == nil && isLogInfo {
		responseExtMap := make(map[string]interface{})
		if err = json.Unmarshal(responseExt, &responseExtMap); err == nil && responseExtMap["loginfo"] != nil {
			if logInfo, ok := responseExtMap["loginfo"].(map[string]interface{}); ok {
				logInfo["logger"] = openwrap.GetLogAuctionObjectAsURL(ao, true)
			}
			responseExt, _ = json.Marshal(responseExtMap)
		}

		return responseExt
	}
	return responseExt
}
