package openwrap

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	uuid "github.com/satori/go.uuid"
)

func (m OpenWrap) handleBeforeValidationHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	result := hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{}

	rCtx := result.ModuleContext["rctx"].(models.RequestCtx)

	rCtx.IsTestRequest = payload.BidRequest.Test == 2

	var partnerConfigMap map[int]map[string]string
	if rCtx.IsTestRequest {
		// NYC: can this be clubbed with profileid=0
		partnerConfigMap = getTestModePartnerConfigMap(payload.BidRequest, rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	} else if rCtx.ProfileID == 0 {
		partnerConfigMap = getDefaultPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	} else {
		partnerConfigMap = m.cache.GetPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	}
	if len(partnerConfigMap) == 0 {
		return result, errors.New("failed to get profile data")
	}

	rCtx.PartnerConfigMap = partnerConfigMap // keep a copy at module level as well
	result.ModuleContext["rctx"] = rCtx

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		//NYC_TODO: convert /2.5 redirect request to auction
		rctx := result.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		ep.BidRequest, err = m.updateORTBV25Request(rctx, ep.BidRequest)
		return ep, err
	}, hookstage.MutationUpdate, "request-body-with-profile-data")

	return result, nil
}

// updateORTBV25Request copies and updates BidRequest with required values from http header and partnetConfigMap
func (m *OpenWrap) updateORTBV25Request(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest) (*openrtb2.BidRequest, error) {
	if cur, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID][models.AdServerCurrency]; ok {
		bidRequest.Cur = []string{cur}
	}

	// 	adapterThrottleMap, err := GetAdapterThrottleMap(reqWrapper.ReqID, reqWrapper.PartnerConfigMap)
	// if err != nil {
	// 	return nil, err
	// }

	// adunitCfg := m.cache.GetAdunitConfigFromCache(bidRequest, rctx.PubID, rctx.ProfileID, rctx.DisplayID)
	platform, _ := rctx.PartnerConfigMap[models.VersionLevelConfigID][models.PLATFORM_KEY]

	// var videoEnabled bool
	// var allowedConnectionTypes []int
	// if platform == models.PLATFORM_APP || platform == models.PLATFORM_VIDEO || platform == models.PLATFORM_DISPLAY {
	// 	videoEnabled = getDefaultEnabledValueForMediaType(adunitCfg, models.AdunitConfigSlotVideoKey)
	// 	allowedConnectionTypes = getDefaultAllowedConnectionTypes(adunitCfg)
	// }
	// bannerEnabled := getDefaultEnabledValueForMediaType(adunitCfg, models.AdunitConfigSlotBannerKey)
	// if reqWrapper.AdUnitConfig != nil && len(reqWrapper.AdUnitConfig) != 0 && reqWrapper.AdUnitConfig[models.AdunitConfigRegex] != nil {
	// 	if regexEnabled, ok := reqWrapper.AdUnitConfig[models.AdunitConfigRegex].(bool); ok && regexEnabled {
	// 		// Logging erroneous regular expressions
	// 		populateAndLogRegex(reqWrapper.AdUnitConfig, reqWrapper.PubID, reqWrapper.ProfileID)
	// 	}
	// }

	// var isAdPodRequest bool
	// aliasgvlids := make(map[string]uint16)

	reqExt, err := models.GetRequestExtWrapper(bidRequest.Ext)
	if err != nil {
		return bidRequest, err
	}

	loggerID := getLoggerID(reqExt)

	reqExt.Prebid.SupportDeals = reqExt.Wrapper.SupportDeals && rctx.IsCTVRequest
	// AlternateBidderCodes: getMarketplaceBidders(newReq.Ext, reqWrapper.PartnerConfigMap),
	reqExt.Prebid.Debug = rctx.Debug

	for i := 0; i < len(bidRequest.Imp); i++ {
		var adpodExt *models.AdPod
		// var isAdPodImpression bool
		imp := &bidRequest.Imp[i]
		// //Wrapper
		// impWrapper := &models.ImpWrapper{
		// 	Imp:    eachImp,
		// 	Bidder: map[string]*models.BidderWrapper{},
		// }

		if imp.BidFloor == 0 {
			imp.BidFloorCur = ""
		} else if imp.BidFloorCur == "" {
			imp.BidFloorCur = "USD"
		}

		if nil != imp.Video && nil == reqExt.Prebid.Macros {
			// provide custom macros for video event trackers
			pubMaticPlatform := GetDevicePlatform(rctx.UA, *bidRequest, platform)
			bidderparams.SetVASTEventMacros(&reqExt, *bidRequest, "", strconv.Itoa(rctx.DisplayID), pubMaticPlatform)
		}

		// if reqWrapper.AdUnitConfig != nil && len(reqWrapper.AdUnitConfig) != 0 {

		// 	// Currently we are supporting Video config via Ad Unit config file for in-app / video / display profiles
		// 	if (reqWrapper.Platform == models.PLATFORM_APP || reqWrapper.Platform == models.PLATFORM_VIDEO || reqWrapper.Platform == models.PLATFORM_DISPLAY) && eachImp.Video != nil {
		// 		if reqWrapper.BidRequest.Site != nil && reqWrapper.BidRequest.Site.Content != nil {
		// 			stats.IncrReqImpsWithSiteContentCount(reqWrapper.PubID)
		// 		}
		// 		if reqWrapper.BidRequest.App != nil && reqWrapper.BidRequest.App.Content != nil {
		// 			stats.IncrReqImpsWithAppContentCount(reqWrapper.PubID)
		// 		}
		// 		if err := updateVideoObjectWithAdunitConfig(reqWrapper.ReqID, eachImp, reqWrapper.AdUnitConfig, videoEnabled,
		// 			allowedConnectionTypes, newReq.Device.ConnectionType, reqWrapper.PubID, reqWrapper.ProfileID, reqWrapper.RequestAPI); err != nil {
		// 			return nil, err
		// 		}
		// 	}

		// 	// We are supporting Banner config via Ad Unit config file for all platforms. Hence, there is no platform check for updating Banner object
		// 	if eachImp.Banner != nil {
		// 		updateBannerObjectWithAdunitConfig(reqWrapper.ReqID, eachImp, reqWrapper.AdUnitConfig, bannerEnabled, reqWrapper.PubID, reqWrapper.ProfileID)
		// 	}
		// 	if eachImp.Banner == nil && eachImp.Video == nil {
		// 		newReq.Imp = append(newReq.Imp[:i], newReq.Imp[i+1:]...)
		// 						i--
		// 		partnerChan <- *objects.GetErrorResponseForImpressionV25(reqWrapper.ReqID, eachImp, "", errorcodes.ErrBannerVideoDisabled)
		// 		continue
		// 	}
		// }

		// if eachImp.Video != nil && eachImp.Video.Ext != nil && IsCTVAPIRequest(reqWrapper.RequestAPI) {
		// 	ext, ok := eachImp.Video.Ext.(map[string]interface{})
		// 	if ok && ext[models.ORTBExtAdPod] != nil {
		// 		//TODO: Read AdPod Object
		// 		adpodExt, _ = ext[models.ORTBExtAdPod].(*openrtb.AdPod)
		// 		isAdPodImpression = true
		// 		if !isAdPodRequest {
		// 			isAdPodRequest = true
		// 			stats.IncrCTVReqCountWithAdPod(reqWrapper.PubID, reqWrapper.ProfileID)
		// 		}
		// 	}
		// }

		impExt := &models.ImpExtension{}
		if len(imp.Ext) != 0 {
			err = json.Unmarshal(imp.Ext, impExt)
			if err != nil {
				return bidRequest, err
			}
		}

		for _, partnerConfig := range rctx.PartnerConfigMap {
			if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
				continue
			}

			partneridstr, ok := partnerConfig[models.PARTNER_ID]
			if !ok {
				continue
			}
			partnerID, err := strconv.Atoi(partneridstr)
			if err != nil || partnerID == models.VersionLevelConfigID {
				continue
			}

			bidderCode := partnerConfig[models.BidderCode]

			// if adapterThrottleMap[bidderCode] {
			// 					errorResponse := objects.GetErrorResponseForImpressionV25(reqWrapper.ReqID, eachImp, bidderCode, errorcodes.ErrPartnerThrottled)
			// 	PopulateClientConfigForErrorResponse(reqWrapper.ClientConfigFlag, errorResponse, eachImp, reqWrapper.AdUnitConfig)
			// 	partnerChan <- *errorResponse
			// 	continue
			// }

			bidder := partnerConfig[models.PREBID_PARTNER_NAME]
			var bidderParams json.RawMessage
			switch bidder {
			case string(openrtb_ext.BidderPubmatic), models.BidderPubMaticSecondaryAlias:
				bidderParams, err = bidderparams.PreparePubMaticParamsV25(rctx, m.cache, *bidRequest, *imp, *impExt, partnerID)
			case models.BidderVASTBidder:
				bidderParams, err = bidderparams.PrepareVASTBidderParams(rctx, m.cache, *bidRequest, *imp, *impExt, partnerID, adpodExt)
			default:
				bidderParams, err = bidderparams.PrepareAdapterParamsV25(rctx, m.cache, *bidRequest, *imp, *impExt, partnerID)
			}

			if err != nil || len(bidderParams) == 0 {
				continue
			}

			impExt.Prebid.Bidder[bidder] = bidderParams

			if alias, ok := partnerConfig[models.IsAlias]; ok && alias == "1" {
				if reqExt.Prebid.Aliases == nil {
					reqExt.Prebid.Aliases = make(map[string]string)
				}
				if prebidPartnerName, ok := partnerConfig[models.PREBID_PARTNER_NAME]; ok {
					reqExt.Prebid.Aliases[bidderCode] = adapters.ResolveOWBidder(prebidPartnerName)
				}
			}
		} // for(rctx.PartnerConfigMap

		imp.Ext, err = json.Marshal(impExt)
		if err != nil {
			// NYC_TODO: mark and remove impression
			continue
		}

		if platform == models.PLATFORM_VIDEO {
			// set banner object back to nil
			imp.Banner = nil
		}

		// NYC_TODO:
		// setContentTransparencyObject(newReq, &requestExt, wtExt, eachImp, reqWrapper.AdUnitConfig, reqWrapper.PartnerConfigMap, adapterThrottleMap)
	} // for(imp

	reqExt.Prebid.BidderParams, err = updateRequestExtBidderParamsPubmatic(reqExt.Prebid.BidderParams, rctx.Cookies, loggerID, string(openrtb_ext.BidderPubmatic))
	if err != nil {
		// return bidRequest, err
	}

	// replaceAppObjectFromAdUnitConfig(reqWrapper.AdUnitConfig, newReq)

	// replaceDeviceTypeFromAdUnitConfig(reqWrapper.AdUnitConfig, newReq)

	// if !bidExist {
	// 	return nil, errorcodes.ErrInvalidImpression
	// }

	bidRequest.Source.TID = bidRequest.ID // NYC: is this needed

	// if reqWrapper.Platform == models.PLATFORM_APP || reqWrapper.Platform == models.PLATFORM_VIDEO {
	// 	sChainObj := getSChainObj(reqWrapper.PartnerConfigMap)
	// 	if sChainObj != nil {
	// 		setSchainInSourceObject(newReq.Source, sChainObj)
	// 	}
	// }

	// pg, priceGranularity, err := computePriceGranularity(reqWrapper.PartnerConfigMap, reqWrapper.RequestAPI, reqWrapper.BidRequest.Test)
	// if err != nil {
	// 	return nil, err
	// }

	// requestExt.Prebid.AliasGVLIDs = aliasgvlids

	reqExt.Prebid.Targeting = &openrtb_ext.ExtRequestTargeting{
		// PriceGranularity:  pg,
		IncludeBidderKeys: true,
		IncludeWinners:    true,
	}

	// setIncludeBrandCategory(wtExt, &requestExt.Prebid, reqWrapper.PartnerConfigMap, IsCTVAPIRequest(reqWrapper.RequestAPI))

	// if isAdPodRequest {
	// 	requestExt.AdPod = getPrebidExtRequestAdPod(reqWrapper)
	//  }
	// updateFloorsExtObjectFromAdUnitConfig(wtExt, reqWrapper.AdUnitConfig, newReq, &requestExt)
	// setPriceFloorFetchURL(&requestExt, reqWrapper.PartnerConfigMap)
	return bidRequest, nil
}

func getDomainFromUrl(pageUrl string) string {
	u, err := url.Parse(pageUrl)
	if err != nil {
		return ""
	}

	return u.Host
}

func getDefaultAllowedConnectionTypes(adUnitConfigMap models.AdUnitConfig) []int {
	if len(adUnitConfigMap) == 0 {
		return nil
	}

	var defaultAdUnitConfig map[string]interface{} = nil
	if adUnitConfigMap[models.AdunitConfigDefaultKey] != nil {
		defaultAdUnitConfig = adUnitConfigMap[models.AdunitConfigDefaultKey].(map[string]interface{})
	}

	if defaultAdUnitConfig == nil || defaultAdUnitConfig[models.AdunitConfigSlotVideoKey] == nil {
		return nil
	}

	videoConfig, ok := defaultAdUnitConfig[models.AdunitConfigSlotVideoKey].(map[string]interface{})
	if videoConfig[models.AdunitConfigConfigKey] == nil {
		return nil
	}
	var configObjInVideoConfig map[string]interface{} = nil
	configObjInVideoConfig = videoConfig[models.AdunitConfigConfigKey].(map[string]interface{})

	if ok && configObjInVideoConfig[models.VideoConnectionType] != nil {
		allowedConnectionTypes := GetIntArray(configObjInVideoConfig[models.VideoConnectionType])
		return allowedConnectionTypes
	}

	return nil
}

func getLoggerID(reqExt models.ExtRequestWrapper) string {
	if reqExt.Wrapper.LoggerImpressionID != "" {
		return reqExt.Wrapper.LoggerImpressionID
	}
	return uuid.NewV4().String()
}

// NYC: make this generic
func updateRequestExtBidderParamsPubmatic(bidderParams json.RawMessage, cookie, loggerID, bidderCode string) (json.RawMessage, error) {
	bidderParamsMap := make(map[string]map[string]interface{})
	err := json.Unmarshal(bidderParams, &bidderParamsMap)
	if err != nil {
		return nil, err
	}

	bidderParamsMap[bidderCode] = map[string]interface{}{
		models.COOKIE:             cookie,
		models.WrapperLoggerImpID: loggerID,
	}

	return json.Marshal(bidderParamsMap)
}
