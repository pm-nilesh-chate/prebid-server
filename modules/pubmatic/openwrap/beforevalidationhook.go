package openwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (m OpenWrap) handleBeforeValidationHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	result := hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{}
	rCtx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)

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

	rCtx.Platform, _ = rCtx.GetVersionLevelKey(models.PLATFORM_KEY)
	rCtx.PageURL = getPageURL(payload.BidRequest)
	rCtx.DevicePlatform = GetDevicePlatform(rCtx.UA, payload.BidRequest, rCtx.Platform)

	if payload.BidRequest.Site != nil {
		if len(payload.BidRequest.Site.Domain) != 0 {
			rCtx.Source = payload.BidRequest.Site.Domain
		} else if len(payload.BidRequest.Site.Page) != 0 {
			rCtx.Source = getDomainFromUrl(payload.BidRequest.Site.Page)
		}
	} else if payload.BidRequest.App != nil {
		rCtx.Source = payload.BidRequest.App.Bundle
	}

	var err error
	rCtx.AdapterThrottleMap, err = GetAdapterThrottleMap(rCtx.PartnerConfigMap)
	if err != nil {
		return result, err
	}

	var videoEnabled bool
	var bannerEnabled bool
	var allowedConnectionTypes []int
	rCtx.AdUnitConfig = m.cache.GetAdunitConfigFromCache(payload.BidRequest, rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	if rCtx.Platform == models.PLATFORM_APP || rCtx.Platform == models.PLATFORM_VIDEO || rCtx.Platform == models.PLATFORM_DISPLAY {
		videoEnabled = getDefaultEnabledValueForMediaType(rCtx.AdUnitConfig, models.AdunitConfigSlotVideoKey)
		allowedConnectionTypes = getDefaultAllowedConnectionTypes(rCtx.AdUnitConfig)
	}
	bannerEnabled = getDefaultEnabledValueForMediaType(rCtx.AdUnitConfig, models.AdunitConfigSlotBannerKey)
	if rCtx.AdUnitConfig != nil && rCtx.AdUnitConfig.Config[models.AdunitConfigRegex] != nil {
		if v, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigRegex]; ok && v.Regex != nil && *v.Regex == true {
			// Logging erroneous regular expressions
			populateAndLogRegex(rCtx.AdUnitConfig)
		}
	}

	_ = videoEnabled
	_ = bannerEnabled
	_ = allowedConnectionTypes

	// var isAdPodRequest bool

	requestExt := &openrtb_ext.ExtOWRequest{}
	if len(payload.BidRequest.Ext) != 0 {
		err := json.Unmarshal(payload.BidRequest.Ext, requestExt)
		if err != nil {
			return result, err
		}
	}
	requestExt.Prebid.SupportDeals = rCtx.PreferDeals // && IsCTVAPIRequest(reqWrapper.RequestAPI),
	// requestExt.Prebid.AlternateBidderCodes = getMarketplaceBidders(reqWrapper.BidRequest.Ext, reqWrapper.PartnerConfigMap),

	for i := 0; i < len(payload.BidRequest.Imp); i++ {
		var adpodExt *models.AdPod
		// var isAdPodImpression bool
		imp := payload.BidRequest.Imp[i]

		if len(requestExt.Prebid.Macros) == 0 && imp.Video != nil {
			// provide custom macros for video event trackers
			requestExt.Prebid.Macros = getVASTEventMacros(rCtx)
		}

		// if rCtx.AdUnitConfig != nil {
		// 	// Currently we are supporting Video config via Ad Unit config file for in-app / video / display profiles
		// 	if (rCtx.Platform == models.PLATFORM_APP || rCtx.Platform == models.PLATFORM_VIDEO || rCtx.Platform == models.PLATFORM_DISPLAY) && imp.Video != nil {
		// 		if err := updateVideoObjectWithAdunitConfig(reqWrapper.ReqID, eachImp, reqWrapper.AdUnitConfig, videoEnabled,
		// 			allowedConnectionTypes, newReq.Device.ConnectionType, reqWrapper.PubID, reqWrapper.ProfileID, reqWrapper.RequestAPI); err != nil {
		// 			return nil, err
		// 		}
		// 	}

		// 	// We are supporting Banner config via Ad Unit config file for all platforms. Hence, there is no platform check for updating Banner object
		// 	if eachImp.Banner != nil {
		// 		updateBannerObjectWithAdunitConfig(reqWrapper.ReqID, eachImp, reqWrapper.AdUnitConfig, bannerEnabled, reqWrapper.PubID, reqWrapper.ProfileID)
		// 	}
		// 	if eachImp.Banner == nil && eachImp.Video == nil && eachImp.Native == nil {
		// 		newReq.Imp = append(newReq.Imp[:i], newReq.Imp[i+1:]...)
		// 		logger.DebugWithBid(reqWrapper.ReqID, "No Valid Banner/Video/Native present for impID: %v ", *eachImp.Id)
		// 		i--
		// 		partnerChan <- *objects.GetErrorResponseForImpressionV25(reqWrapper.ReqID, eachImp, "", errorcodes.ErrBannerVideoDisabled)
		// 		continue
		// 	}
		// }

		impExt := &models.ImpExtension{}
		if len(imp.Ext) != 0 {
			err := json.Unmarshal(imp.Ext, impExt)
			if err != nil {
				return result, err
			}
		}

		for _, partnerConfig := range rCtx.PartnerConfigMap {
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

			bidder := partnerConfig[models.PREBID_PARTNER_NAME]
			var slot string
			var bidderParams json.RawMessage
			switch bidder {
			case string(openrtb_ext.BidderPubmatic), models.BidderPubMaticSecondaryAlias:
				slot, bidderParams, err = bidderparams.PreparePubMaticParamsV25(rCtx, m.cache, *payload.BidRequest, imp, *impExt, partnerID)
			case models.BidderVASTBidder:
				slot, bidderParams, err = bidderparams.PrepareVASTBidderParams(rCtx, m.cache, *payload.BidRequest, imp, *impExt, partnerID, adpodExt)
			default:
				slot, bidderParams, err = bidderparams.PrepareAdapterParamsV25(rCtx, m.cache, *payload.BidRequest, imp, *impExt, partnerID)
			}

			if err != nil || len(bidderParams) == 0 {
				continue
			}

			slotType := "banner"
			if imp.Video != nil {
				slotType = "video"
			}

			if _, ok := rCtx.ImpBidCtx[imp.ID]; !ok {
				rCtx.ImpBidCtx[imp.ID] = models.ImpCtx{
					TagID:             imp.TagID,
					IsRewardInventory: impExt.Reward,
					MatchedSlot:       slot,
					Type:              slotType,
					KGPV:              rCtx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN],
					Bidders:           make(map[string]json.RawMessage),
					BidCtx:            make(map[string]models.BidCtx),
				}
			}

			rCtx.ImpBidCtx[imp.ID].Bidders[bidderCode] = bidderParams

			if alias, ok := partnerConfig[models.IsAlias]; ok && alias == "1" {
				if prebidPartnerName, ok := partnerConfig[models.PREBID_PARTNER_NAME]; ok {
					rCtx.Aliases[bidderCode] = adapters.ResolveOWBidder(prebidPartnerName)
				}
			}
		} // for(rctx.PartnerConfigMap

		// NYC_TODO:
		if cto := setContentTransparencyObject(rCtx, requestExt, imp.ID, rCtx.AdUnitConfig); cto != nil {
			requestExt.Prebid.Transparency = cto
		}
	} // for(imp

	moduleCtx.ModuleContext["rctx"] = rCtx

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		//NYC_TODO: convert /2.5 redirect request to auction
		rctx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
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

	reqExt, err := models.GetRequestExtWrapper(bidRequest.Ext)
	if err != nil {
		return bidRequest, err
	}

	// loggerID := getLoggerID(reqExt)

	requestExt := &openrtb_ext.ExtOWRequest{}
	if len(bidRequest.Ext) != 0 {
		err := json.Unmarshal(bidRequest.Ext, requestExt)
		if err != nil {
			return bidRequest, err
		}
	}
	requestExt.Prebid.SupportDeals = rctx.PreferDeals // && IsCTVAPIRequest(reqWrapper.RequestAPI),
	// requestExt.Prebid.AlternateBidderCodes = getMarketplaceBidders(reqWrapper.BidRequest.Ext, reqWrapper.PartnerConfigMap),

	reqExt.Prebid.SupportDeals = reqExt.Wrapper.SupportDeals && rctx.IsCTVRequest
	// AlternateBidderCodes: getMarketplaceBidders(newReq.Ext, reqWrapper.PartnerConfigMap),
	reqExt.Prebid.Debug = rctx.Debug

	aliasgvlids := make(map[string]uint16)

	for i := 0; i < len(bidRequest.Imp); i++ {
		//var adpodExt *models.AdPod
		// var isAdPodImpression bool
		imp := bidRequest.Imp[i]

		if imp.BidFloor == 0 {
			imp.BidFloorCur = ""
		} else if imp.BidFloorCur == "" {
			imp.BidFloorCur = "USD"
		}

		if len(requestExt.Prebid.Macros) == 0 && imp.Video != nil {
			// provide custom macros for video event trackers
			requestExt.Prebid.Macros = getVASTEventMacros(rctx)
		}

		// if rCtx.AdUnitConfig != nil {
		// 	// Currently we are supporting Video config via Ad Unit config file for in-app / video / display profiles
		// 	if (rCtx.Platform == models.PLATFORM_APP || rCtx.Platform == models.PLATFORM_VIDEO || rCtx.Platform == models.PLATFORM_DISPLAY) && imp.Video != nil {
		// 		if err := updateVideoObjectWithAdunitConfig(reqWrapper.ReqID, eachImp, reqWrapper.AdUnitConfig, videoEnabled,
		// 			allowedConnectionTypes, newReq.Device.ConnectionType, reqWrapper.PubID, reqWrapper.ProfileID, reqWrapper.RequestAPI); err != nil {
		// 			return nil, err
		// 		}
		// 	}

		// 	// We are supporting Banner config via Ad Unit config file for all platforms. Hence, there is no platform check for updating Banner object
		// 	if eachImp.Banner != nil {
		// 		updateBannerObjectWithAdunitConfig(reqWrapper.ReqID, eachImp, reqWrapper.AdUnitConfig, bannerEnabled, reqWrapper.PubID, reqWrapper.ProfileID)
		// 	}
		// 	if eachImp.Banner == nil && eachImp.Video == nil && eachImp.Native == nil {
		// 		newReq.Imp = append(newReq.Imp[:i], newReq.Imp[i+1:]...)
		// 		logger.DebugWithBid(reqWrapper.ReqID, "No Valid Banner/Video/Native present for impID: %v ", *eachImp.Id)
		// 		i--
		// 		partnerChan <- *objects.GetErrorResponseForImpressionV25(reqWrapper.ReqID, eachImp, "", errorcodes.ErrBannerVideoDisabled)
		// 		continue
		// 	}
		// }

		impExt := &models.ImpExtension{}
		if len(imp.Ext) != 0 {
			err := json.Unmarshal(imp.Ext, impExt)
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

			bidderParams, ok := rctx.ImpBidCtx[imp.ID].Bidders[bidderCode]
			if !ok {
				continue
			}

			if impExt.Prebid.Bidder == nil {
				impExt.Prebid.Bidder = make(map[string]json.RawMessage)
			}
			impExt.Prebid.Bidder[bidderCode] = bidderParams

			if partnerConfig[models.PREBID_PARTNER_NAME] == models.BidderVASTBidder {
				updateAliasGVLIds(aliasgvlids, bidderCode, partnerConfig)
			}
		} // for(rctx.PartnerConfigMap

		// NYC_TODO:
		if cto := setContentTransparencyObject(rctx, requestExt, imp.ID, rctx.AdUnitConfig); cto != nil {
			requestExt.Prebid.Transparency = cto
		}

		bidRequest.Imp[i].Ext, err = json.Marshal(impExt)
		if err != nil {
			return bidRequest, err
		}

	} // for(imp

	for k, v := range rctx.Aliases {
		reqExt.Prebid.Aliases[k] = v
	}

	reqExt.Prebid.BidderParams, err = updateRequestExtBidderParamsPubmatic(reqExt.Prebid.BidderParams, rctx.Cookies, rctx.LoggerImpressionID, string(openrtb_ext.BidderPubmatic))
	if err != nil {
		// return bidRequest, err
	}

	// replaceAppObjectFromAdUnitConfig(rCtx.AdUnitConfig, newReq)

	// replaceDeviceTypeFromAdUnitConfig(rCtx.AdUnitConfig, newReq)

	// if !bidExist {
	// 	return nil, errorcodes.ErrInvalidImpression
	// }

	bidRequest.Source.TID = bidRequest.ID // NYC: is this needed

	// if rCtx.Platform == models.PLATFORM_APP || rCtx.Platform == models.PLATFORM_VIDEO {
	// 	sChainObj := getSChainObj(reqWrapper.PartnerConfigMap)
	// 	if sChainObj != nil {
	// 		setSchainInSourceObject(newReq.Source, sChainObj)
	// 	}
	// }

	// pg, priceGranularity, err := computePriceGranularity(reqWrapper.PartnerConfigMap, reqWrapper.RequestAPI, reqWrapper.BidRequest.Test)
	// if err != nil {
	// 	return nil, err
	// }

	requestExt.Prebid.AliasGVLIDs = aliasgvlids

	reqExt.Prebid.Targeting = &openrtb_ext.ExtRequestTargeting{
		// PriceGranularity:  pg,
		IncludeBidderKeys: true,
		IncludeWinners:    true,
	}

	// setIncludeBrandCategory(wtExt, &requestExt.Prebid, reqWrapper.PartnerConfigMap, IsCTVAPIRequest(reqWrapper.RequestAPI))

	// if isAdPodRequest {
	// 	requestExt.AdPod = getPrebidExtRequestAdPod(reqWrapper)
	//  }
	// updateFloorsExtObjectFromAdUnitConfig(wtExt, rCtx.AdUnitConfig, newReq, &requestExt)
	// setPriceFloorFetchURL(&requestExt, reqWrapper.PartnerConfigMap)
	bidRequest.Ext, err = json.Marshal(reqExt)
	return bidRequest, err
}

func getDomainFromUrl(pageUrl string) string {
	u, err := url.Parse(pageUrl)
	if err != nil {
		return ""
	}

	return u.Host
}

func getDefaultAllowedConnectionTypes(adUnitConfigMap *adunitconfig.AdUnitConfig) []int {
	if adUnitConfigMap == nil {
		return nil
	}

	if v, ok := adUnitConfigMap.Config[models.AdunitConfigDefaultKey]; ok && v.Video != nil {
		return v.Video.Config.ConnectionType
	}

	return nil
}

// always perfer rCtx.LoggerImpressionID received in request. Create a new once if it is not availble.
// func getLoggerID(reqExt models.ExtRequestWrapper) string {
// 	if reqExt.Wrapper.LoggerImpressionID != "" {
// 		return reqExt.Wrapper.LoggerImpressionID
// 	}
// 	return uuid.NewV4().String()
// }

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

func getPageURL(bidRequest *openrtb2.BidRequest) string {
	if bidRequest.App != nil && bidRequest.App.StoreURL != "" {
		return bidRequest.App.StoreURL
	} else if bidRequest.Site != nil && bidRequest.Site.Page != "" {
		return bidRequest.Site.Page
	}
	return ""
}

// getVASTEventMacros populates macros with PubMatic specific macros
// These marcros is used in replacing with actual values of Macros in case of Video Event tracke URLs
// If this function fails to determine value of any macro then it continues with next macro setup
// returns true when at least one macro is added to map
func getVASTEventMacros(rctx models.RequestCtx) map[string]string {
	return map[string]string{
		string(models.MacroProfileID):           strconv.Itoa(rctx.ProfileID),
		string(models.MacroUnixTimeStamp):       strconv.FormatInt(time.Now().Unix(), 10),
		string(models.MacroPlatform):            fmt.Sprintf("%d", rctx.DevicePlatform),
		string(models.MacroWrapperImpressionID): rctx.LoggerImpressionID,
		string(models.MacroSSAI):                rctx.SSAI,
		string(models.MacroProfileVersionID):    fmt.Sprintf("%d", rctx.DisplayID),
	}
}

func updateAliasGVLIds(aliasgvlids map[string]uint16, bidderCode string, partnerConfig map[string]string) {
	if vendorID, ok := partnerConfig[models.VENDORID]; ok && vendorID != "" {
		vid, err := strconv.ParseUint(vendorID, 10, 64)
		if err != nil {
			return
		}

		if vid == 0 {
			return
		}
		aliasgvlids[bidderCode] = uint16(vid)
	}
}
