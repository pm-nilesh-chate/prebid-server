package openwrap

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/request"
	ow_request "github.com/prebid/prebid-server/modules/pubmatic/openwrap/request"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// updateORTBV25Request copies and updates BidRequest with required values from http header and partnetConfigMap
func (m *OpenWrap) updateORTBV25Request(rctx models.RequestCtx, body []byte) ([]byte, error) {
	bidRequest := &openrtb2.BidRequest{}
	err := json.Unmarshal(body, bidRequest)
	if err != nil {
		return body, fmt.Errorf("failed to decode request %v", err)
	}

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

	reqExt, err := request.GetRequestExt(bidRequest.Ext)
	if err != nil {
		return body, err
	}

	reqExt.Prebid.SupportDeals = reqExt.Wrapper.SupportDeals && rctx.IsCTVRequest
	// AlternateBidderCodes: getMarketplaceBidders(newReq.Ext, reqWrapper.PartnerConfigMap),

	for i := 0; i < len(bidRequest.Imp); i++ {
		var adpodExt *ow_request.AdPod
		// var isAdPodImpression bool
		eachImp := &bidRequest.Imp[i]
		// //Wrapper
		// impWrapper := &request.ImpWrapper{
		// 	Imp:    eachImp,
		// 	Bidder: map[string]*request.BidderWrapper{},
		// }

		updateImpFloorDefaultCurrency(eachImp)

		if nil != eachImp.Video && nil == reqExt.Prebid.Macros {
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

		impExt := &request.ImpExtension{}
		if len(eachImp.Ext) != 0 {
			err = json.Unmarshal(eachImp.Ext, impExt)
			if err != nil {
				return body, err
			}
		}

		prebidBidderParams := make(map[string]json.RawMessage)
		for _, partnerConfig := range rctx.PartnerConfigMap {
			if partnerConfig[models.SERVER_SIDE_FLAG] == "" || partnerConfig[models.SERVER_SIDE_FLAG] == "0" {
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
			case models.BidderPubMatic, models.BidderPubMaticSecondaryAlias:
				bidderParams, err = bidderparams.PreparePubMaticParamsV25(rctx, m.cache, *bidRequest, *eachImp, *impExt, partnerID)
			case models.BidderVASTBidder:
				bidderParams, err = bidderparams.PrepareVASTBidderParams(rctx, m.cache, *bidRequest, *eachImp, *impExt, partnerID, adpodExt)
			default:
				bidderParams, err = bidderparams.PrepareAdapterParamsV25(rctx, m.cache, *bidRequest, *eachImp, *impExt, partnerID)
			}

			if err != nil || len(bidderParams) == 0 {
				continue
			}

			prebidBidderParams[bidder] = bidderParams

			if alias, ok := partnerConfig[models.IsAlias]; ok && alias == "1" {
				if reqExt.Prebid.Aliases == nil {
					reqExt.Prebid.Aliases = make(map[string]string)
				}
				if prebidPartnerName, ok := partnerConfig[models.PREBID_PARTNER_NAME]; ok {
					reqExt.Prebid.Aliases[bidderCode] = adapters.ResolveOWBidder(prebidPartnerName)
				}
			}
		} // rctx.PartnerConfigMap

		if impExt.Prebid == nil {
			impExt.Prebid = &openrtb_ext.ExtImpPrebid{}
		}

		impExt.Prebid.Bidder = prebidBidderParams
		impExt.Prebid.IsRewardedInventory = impExt.Reward

		eachImp.Ext, err = json.Marshal(impExt)
		if err != nil {
			// NYC_TODO: mark and remove impression
			continue
		}

		if platform == models.PLATFORM_VIDEO {
			// set banner object back to nil
			eachImp.Banner = nil
		}

		// NYC_TODO:
		// setContentTransparencyObject(newReq, &requestExt, wtExt, eachImp, reqWrapper.AdUnitConfig, reqWrapper.PartnerConfigMap, adapterThrottleMap)
	}

	bidderparams.UpdateRequestExtBidderParamsForPubmatic(&reqExt.Prebid.BidderParams, rctx.Cookies, "NYC-loggerImpID", "NYC-loggerImpID", platform, string(openrtb_ext.BidderPubmatic))

	// replaceAppObjectFromAdUnitConfig(reqWrapper.AdUnitConfig, newReq)

	// replaceDeviceTypeFromAdUnitConfig(reqWrapper.AdUnitConfig, newReq)

	// if !bidExist {
	// 	return nil, errorcodes.ErrInvalidImpression
	// }

	// if reqWrapper.BidRequest.Id != nil {
	// 	newReq.Source.TID = new(string)
	// 	*newReq.Source.TID = reqWrapper.ReqID
	// }

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

	// requestExt.Prebid.Targeting = &openrtb_ext.ExtRequestTargeting{
	// 	PriceGranularity:  pg,
	// 	IncludeBidderKeys: true,
	// 	IncludeWinners:    true,
	// }

	// setIncludeBrandCategory(wtExt, &requestExt.Prebid, reqWrapper.PartnerConfigMap, IsCTVAPIRequest(reqWrapper.RequestAPI))

	// if isAdPodRequest {
	// 	requestExt.AdPod = getPrebidExtRequestAdPod(reqWrapper)
	//  }
	// if reqWrapper.Debug || reqWrapper.WakandaDebug.Enabled {
	// 	requestExt.Prebid.Debug = true
	// }
	// updateFloorsExtObjectFromAdUnitConfig(wtExt, reqWrapper.AdUnitConfig, newReq, &requestExt)
	// setPriceFloorFetchURL(&requestExt, reqWrapper.PartnerConfigMap)
	bidRequest.Ext, err = json.Marshal(reqExt)
	if err != nil {
		return body, err
	}

	return json.Marshal(bidRequest)
}

// updateImpFloorDefaultCurrency updates default currency to USD if only bidfloor values is provided,
// if only bidfloorCur is provided then bidfloor and bidfloorcur are resetted
func updateImpFloorDefaultCurrency(imp *openrtb2.Imp) {
	if imp.BidFloor == 0 {
		imp.BidFloorCur = ""
	} else if imp.BidFloorCur == "" {
		imp.BidFloorCur = "USD"
	}
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
