package bidderparams

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/request"
)

// setVASTEventMacros populates reqExt.Prebid.Macros with PubMatic specific macros
// These marcros is used in replacing with actual values of Macros in case of Video Event tracke URLs
// If this function fails to determine value of any macro then it continues with next macro setup
// returns true when at least one macro is added to map
func SetVASTEventMacros(reqExt *request.ExtRequest, bidReq openrtb2.BidRequest, wrapperImpressionID, displayVersionID string, pubmaticPlatform models.DevicePlatform) (bool, error) {
	reqExt.Prebid.Macros = make(map[string]string)
	errMsgs := []string{}
	reqExt.Prebid.Macros[string(models.MacroProfileID)] = strconv.Itoa(reqExt.Wrapper.ProfileId)

	if displayVersionID == "" {
		errMsgs = append(errMsgs, "version id")
	} else {
		reqExt.Prebid.Macros[string(models.MacroProfileVersionID)] = displayVersionID
	}

	reqExt.Prebid.Macros[string(models.MacroUnixTimeStamp)] = strconv.FormatInt(time.Now().Unix(), 10)
	reqExt.Prebid.Macros[string(models.MacroPlatform)] = strconv.Itoa(int(pubmaticPlatform))
	if wrapperImpressionID == "" {
		errMsgs = append(errMsgs, "wrapper impression id")
	} else {
		reqExt.Prebid.Macros[string(models.MacroWrapperImpressionID)] = wrapperImpressionID
	}

	if len(reqExt.Wrapper.SSAI) != 0 {
		reqExt.Prebid.Macros[string(models.MacroSSAI)] = reqExt.Wrapper.SSAI
	}

	var err error
	if len(errMsgs) > 0 {
		err = fmt.Errorf("invalid '%s'. Not able to add to custom macro map", strings.Join(errMsgs, ","))
	}
	return true, err
}

func PrepareVASTBidderParams(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt request.ImpExtension, partnerID int, adpodExt *request.AdPod) (json.RawMessage, errorcodes.IError) {
	if imp.Video == nil {
		return nil, nil
	}

	slots, slotMap, _ := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)
	if len(slots) == 0 {
		return nil, nil
	}

	pubVASTTags := cache.GetPublisherVASTTagsFromCache(rctx.PubID)
	if len(pubVASTTags) == 0 {
		return nil, nil
	}

	matchedSlotKeys, err := getVASTBidderSlotKeys(&imp, slots[0], slotMap, pubVASTTags, adpodExt)
	if len(matchedSlotKeys) == 0 {
		return nil, err
	}

	// NYC_TODO:
	//setting flagmap
	// bidderWrapper := &BidderWrapper{VASTagFlags: make(map[string]bool)}
	// for _, key := range matchedSlotKeys {
	// 	bidderWrapper.VASTagFlags[key] = false
	// }
	// impWrapper.Bidder[bidderCode] = bidderWrapper

	bidParams := adapters.PrepareVASTBidderParamJSON(&bidRequest, &imp, pubVASTTags, matchedSlotKeys, slotMap, adpodExt)

	/*
		Sample Values
			//slotkey:/15671365/DMDemo1@com.pubmatic.openbid.app@
			//slotKeys:[/15671365/DMDemo1@com.pubmatic.openbid.app@101]
			//slotMap:map[/15671365/DMDemo1@com.pubmatic.openbid.app@101:map[param1:6005 param2:test param3:example]]
			//Ext:{"tags":[{"tagid":"101","url":"sample_url_1","dur":15,"price":"15","params":{"param1":"6005","param2":"test","param3":"example"}}]}
	*/
	return bidParams, nil
}

// getVASTBidderSlotKeys returns all slot keys which are matching to vast tag slot key
func getVASTBidderSlotKeys(imp *openrtb2.Imp,
	slotKey string,
	slotMap map[string]models.SlotMapping,
	pubVASTTags models.PublisherVASTTags,
	adpodExt *request.AdPod) ([]string, errorcodes.IError) {

	//TODO: Optimize this function
	var (
		result, defaultMapping []string
		validationErr          errorcodes.IError
		isValidationError      bool
	)

	for _, sm := range slotMap {
		//formCaseInsensitiveVASTSlotKey forms slotKey for case in-sensitive comparison.
		//It converts AdUnit part of slot key to lower case for comparison.
		//Currently it only supports slot-keys of the format <AdUnit>@<bundle-id>@<VAST ID>
		//This function needs to be enhanced to support different slot-key formats.
		key := formCaseInsensitiveVASTSlotKey(sm.SlotName)
		tempSlotKey := formCaseInsensitiveVASTSlotKey(slotKey)
		isDefaultMappingSelected := false

		index := strings.Index(key, "@@")
		if -1 != index {
			//prefix check only for `/15671365/MG_VideoAdUnit@`
			if false == strings.HasPrefix(tempSlotKey, key[:index+1]) {
				continue
			}

			//getting slot key `/15671365/MG_VideoAdUnit@@`
			tempSlotKey = key[:index+2]
			isDefaultMappingSelected = true
		} else if false == strings.HasPrefix(key, tempSlotKey) {
			continue
		}

		//get vast tag id and slotkey
		vastTagID, _ := strconv.Atoi(key[len(tempSlotKey):])
		if 0 == vastTagID {
			continue
		}

		//check pubvasttag details
		vastTag, ok := pubVASTTags[vastTagID]
		if false == ok {
			continue
		}

		//validate vast tag details
		if err := validateVASTTag(vastTag, imp.Video.MinDuration, imp.Video.MaxDuration, adpodExt); nil != err {
			isValidationError = true
			continue
		}

		if isDefaultMappingSelected {
			defaultMapping = append(defaultMapping, sm.SlotName)
		} else {
			result = append(result, sm.SlotName)
		}
	}

	if len(result) == 0 && len(defaultMapping) == 0 && isValidationError {
		validationErr = errorcodes.ErrInvalidVastTag
	}

	if len(result) == 0 {
		return defaultMapping, validationErr
	}

	return result, validationErr
}

// formCaseInsensitiveVASTSlotKey forms slotKey for case in-sensitive comparison.
// It converts AdUnit part of slot key to lower case for comparison.
// Currently it only supports slot-keys of the format <AdUnit>@<bundle-id>@<VAST ID>
// This function needs to be enhanced to support different slot-key formats.
func formCaseInsensitiveVASTSlotKey(key string) string {
	index := strings.Index(key, "@")
	caseInsensitiveSlotKey := key
	if index != -1 {
		caseInsensitiveSlotKey = strings.ToLower(key[:index]) + key[index:]
	}
	return caseInsensitiveSlotKey
}

func validateVASTTag(
	vastTag *models.VASTTag,
	videoMinDuration, videoMaxDuration int64,
	adpod *request.AdPod) error {

	if nil == vastTag {
		return fmt.Errorf("Empty vast tag")
	}

	//TODO: adding checks for Duration and URL
	if len(vastTag.URL) == 0 {
		return fmt.Errorf("VAST tag mandatory parameter 'url' missing: %v", vastTag.ID)
	}

	if vastTag.Duration <= 0 {
		return fmt.Errorf("VAST tag mandatory parameter 'duration' missing: %v", vastTag.ID)
	}

	if vastTag.Duration > int(videoMaxDuration) {
		return fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration > video.maxduration' vastTagID:%v, tag.duration:%v, video.maxduration:%v", vastTag.ID, vastTag.Duration, videoMaxDuration)
	}

	if nil == adpod {
		//non-adpod request
		if vastTag.Duration < int(videoMinDuration) {
			return fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration < video.minduration' vastTagID:%v, tag.duration:%v, video.minduration:%v", vastTag.ID, vastTag.Duration, videoMinDuration)
		}

	} else {
		//adpod request
		if nil != adpod.MinDuration && vastTag.Duration < *adpod.MinDuration {
			return fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration < adpod.minduration' vastTagID:%v, tag.duration:%v, adpod.minduration:%v", vastTag.ID, vastTag.Duration, *adpod.MinDuration)
		}

		if nil != adpod.MaxDuration && vastTag.Duration > *adpod.MaxDuration {
			return fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration > adpod.maxduration' vastTagID:%v, tag.duration:%v, adpod.maxduration:%v", vastTag.ID, vastTag.Duration, *adpod.MaxDuration)
		}
	}

	return nil
}
