package adapters

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/request"

	"github.com/prebid/prebid-server/openrtb_ext"
)

func PrepareVASTBidderParamJSON(request *openrtb2.BidRequest, imp *openrtb2.Imp,
	pubVASTTags models.PublisherVASTTags,
	matchedSlotKeys []string, slotMap map[string]models.SlotMapping,
	adpod *request.AdPod,
	db database.Database) string {

	if nil == imp.Video {
		return ""
	}

	bidderExt := openrtb_ext.ExtImpVASTBidder{}
	bidderExt.Tags = make([]*openrtb_ext.ExtImpVASTBidderTag, len(matchedSlotKeys))
	var tagIndex int = 0
	for _, slotKey := range matchedSlotKeys {
		vastTagID := getVASTTagID(slotKey)
		if 0 == vastTagID {
			continue
		}

		vastTag, ok := pubVASTTags[vastTagID]
		if false == ok {
			continue
		}

		mapping, err := db.GetMappings(slotKey, slotMap)
		if err != nil {
			continue
		}

		//adding mapping parameters as it is in ext.bidder
		params := mapping
		/*
			params := make(map[string]interface{})
			// Copy from the original map of for slot key to the target map
			for key, value := range mapping {
				params[key] = value
			}
		*/

		//prepare bidder ext json here
		bidderExt.Tags[tagIndex] = &openrtb_ext.ExtImpVASTBidderTag{
			//TagID:    strconv.Itoa(vastTag.ID),
			TagID:    slotKey,
			URL:      vastTag.URL,
			Duration: vastTag.Duration,
			Price:    vastTag.Price,
			Params:   params,
		}
		tagIndex++
	}

	if tagIndex > 0 {
		//If any vast tags found then create impression ext for vast bidder.
		bidderExt.Tags = bidderExt.Tags[:tagIndex]
		bidParamBuf, _ := json.Marshal(bidderExt)
		return string(bidParamBuf)
	}
	return ""
}

// getVASTTagID returns VASTTag ID details from slot key
func getVASTTagID(key string) int {
	index := strings.LastIndex(key, "@")
	if -1 == index {
		return 0
	}
	id, _ := strconv.Atoi(key[index+1:])
	return id
}
