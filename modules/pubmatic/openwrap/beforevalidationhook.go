package openwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (m OpenWrap) handleBeforeValidationHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	result := hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
		Reject: true,
	}

	requestExt, err := models.GetRequestExt(payload.BidRequest.Ext)
	if err != nil {
		result.NbrCode = errorcodes.ErrInvalidRequestExtension.Code()
		result.Errors = append(result.Errors, errorcodes.ErrInvalidRequestExtension.Error())
		return result, err
	}

	rCtx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	defer func() {
		moduleCtx.ModuleContext["rctx"] = rCtx
	}()

	rCtx.IsTestRequest = payload.BidRequest.Test == 2

	partnerConfigMap, err := m.getProfileData(rCtx)
	if err != nil || len(partnerConfigMap) == 0 {
		result.NbrCode = errorcodes.ErrInvalidConfiguration.Code()
		result.DebugMessages = append(result.Errors, errorcodes.ErrInvalidConfiguration.Error())
		return result, errors.New("failed to get profile data")
	}

	rCtx.PartnerConfigMap = partnerConfigMap // keep a copy at module level as well
	rCtx.Platform, _ = rCtx.GetVersionLevelKey(models.PLATFORM_KEY)
	rCtx.PageURL = getPageURL(payload.BidRequest)
	rCtx.DevicePlatform = GetDevicePlatform(rCtx.UA, payload.BidRequest, rCtx.Platform)
	rCtx.SendAllBids = isSendAllBids(rCtx)

	if payload.BidRequest.Site != nil {
		if len(payload.BidRequest.Site.Domain) != 0 {
			rCtx.Source = payload.BidRequest.Site.Domain
		} else if len(payload.BidRequest.Site.Page) != 0 {
			rCtx.Source = getDomainFromUrl(payload.BidRequest.Site.Page)
		}
	} else if payload.BidRequest.App != nil {
		rCtx.Source = payload.BidRequest.App.Bundle
	}

	rCtx.AdapterThrottleMap, err = GetAdapterThrottleMap(rCtx.PartnerConfigMap)
	if err != nil {
		result.NbrCode = errorcodes.ErrAllPartnerThrottled.Code()
		result.DebugMessages = append(result.Errors, err.Error())
		return result, err
	}

	priceGranularity, err := computePriceGranularity(rCtx)
	if err != nil {
		result.NbrCode = errorcodes.ErrAllPartnerThrottled.Code()
		result.DebugMessages = append(result.Errors, err.Error())
		return result, err
	}

	rCtx.AdUnitConfig = m.cache.GetAdunitConfigFromCache(payload.BidRequest, rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	if rCtx.AdUnitConfig != nil && rCtx.AdUnitConfig.Config[models.AdunitConfigRegex] != nil {
		if v, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigRegex]; ok && v.Regex != nil && *v.Regex == true {
			errs := populateAndLogRegex(rCtx.AdUnitConfig)
			for _, err := range errs {
				result.Errors = append(result.Errors, err.Error())
			}
		}
	}

	requestExt.Prebid.Debug = rCtx.Debug
	requestExt.Prebid.SupportDeals = rCtx.PreferDeals // && IsCTVAPIRequest(reqWrapper.RequestAPI),
	requestExt.Prebid.AlternateBidderCodes = getMarketplaceBidders(requestExt.Prebid.AlternateBidderCodes, partnerConfigMap)
	requestExt.Prebid.Targeting = &openrtb_ext.ExtRequestTargeting{
		PriceGranularity:  priceGranularity,
		IncludeBidderKeys: true,
		IncludeWinners:    true,
	}
	requestExt.Prebid.BidderParams, _ = updateRequestExtBidderParamsPubmatic(requestExt.Prebid.BidderParams, rCtx.Cookies, rCtx.LoggerImpressionID, string(openrtb_ext.BidderPubmatic))

	aliasgvlids := make(map[string]uint16)
	for i := 0; i < len(payload.BidRequest.Imp); i++ {
		var adpodExt *models.AdPod
		imp := payload.BidRequest.Imp[i]

		if len(requestExt.Prebid.Macros) == 0 && imp.Video != nil {
			// provide custom macros for video event trackers
			requestExt.Prebid.Macros = getVASTEventMacros(rCtx)
		}

		impExt := &models.ImpExtension{}
		if len(imp.Ext) != 0 {
			err := json.Unmarshal(imp.Ext, impExt)
			if err != nil {
				return result, err
			}
		}

		var videoAdUnitCtx, bannerAdUnitCtx models.AdUnitCtx
		if rCtx.AdUnitConfig != nil {
			div := ""
			if impExt.Wrapper != nil {
				div = impExt.Wrapper.Div
			}

			videoAdUnitCtx = adunitconfig.UpdateVideoObjectWithAdunitConfig(rCtx, imp, div, payload.BidRequest.Device.ConnectionType)
			bannerAdUnitCtx = adunitconfig.UpdateBannerObjectWithAdunitConfig(rCtx, imp, div)
		}

		if imp.Banner == nil && imp.Video == nil && imp.Native == nil {
			payload.BidRequest.Imp = append(payload.BidRequest.Imp[:i], payload.BidRequest.Imp[i+1:]...)
			result.Errors = append(result.Errors, fmt.Sprintf("no Valid Banner/Video/Native present for imp: %+v", imp.ID))
			i--
			continue
		}

		slotType := "banner"
		if imp.Video != nil {
			slotType = "video"
		}

		bidderMeta := make(map[string]models.PartnerData)
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

			if alias, ok := partnerConfig[models.IsAlias]; ok && alias == "1" {
				if prebidPartnerName, ok := partnerConfig[models.PREBID_PARTNER_NAME]; ok {
					rCtx.Aliases[bidderCode] = adapters.ResolveOWBidder(prebidPartnerName)
				}
			}

			if partnerConfig[models.PREBID_PARTNER_NAME] == models.BidderVASTBidder {
				updateAliasGVLIds(aliasgvlids, bidderCode, partnerConfig)
			}

			bidderMeta[bidderCode] = models.PartnerData{
				PartnerID:   partnerID,
				MatchedSlot: slot,
				Params:      bidderParams,
				KGPV:        rCtx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN],
			}
		} // for(rctx.PartnerConfigMap

		if cto := setContentTransparencyObject(rCtx, requestExt, imp.ID, rCtx.AdUnitConfig); cto != nil {
			requestExt.Prebid.Transparency = cto
		}

		// update the imp.ext with bidder params for this
		if impExt.Prebid.Bidder == nil {
			impExt.Prebid.Bidder = make(map[string]json.RawMessage)
		}
		for bidder, meta := range bidderMeta {
			impExt.Prebid.Bidder[bidder] = meta.Params
		}

		newExt, err := json.Marshal(impExt)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to update bidder params for impression %s", imp.ID))
		}

		// cache the details for further processing
		if _, ok := rCtx.ImpBidCtx[imp.ID]; !ok {
			rCtx.ImpBidCtx[imp.ID] = models.ImpCtx{
				TagID:             imp.TagID,
				IsRewardInventory: impExt.Reward,
				Type:              slotType,
				Bidders:           make(map[string]models.PartnerData),
				BidCtx:            make(map[string]models.BidCtx),
				NewExt:            json.RawMessage(newExt),
			}
		}

		impCtx := rCtx.ImpBidCtx[imp.ID]
		impCtx.Bidders = bidderMeta
		impCtx.VideoAdUnitCtx = videoAdUnitCtx
		impCtx.BannerAdUnitCtx = bannerAdUnitCtx
		rCtx.ImpBidCtx[imp.ID] = impCtx
	} // for(imp

	adunitconfig.UpdateFloorsExtObjectFromAdUnitConfig(rCtx, &requestExt)
	setPriceFloorFetchURL(&requestExt, rCtx.PartnerConfigMap)

	if len(rCtx.Aliases) != 0 && requestExt.Prebid.Aliases == nil {
		requestExt.Prebid.Aliases = make(map[string]string)
	}
	for k, v := range rCtx.Aliases {
		requestExt.Prebid.Aliases[k] = v
	}

	requestExt.Prebid.AliasGVLIDs = aliasgvlids

	rCtx.NewReqExt, err = json.Marshal(requestExt)
	if err != nil {
		result.Errors = append(result.Errors, "failed to update request.ext "+err.Error())
	}

	newImp, _ := json.Marshal(rCtx.ImpBidCtx)
	result.DebugMessages = append(result.DebugMessages, "new imp: "+string(newImp))
	result.DebugMessages = append(result.DebugMessages, "new request.ext: "+string(rCtx.NewReqExt))

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		rctx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		ep.BidRequest, err = m.applyProfileChanges(rctx, ep.BidRequest)
		return ep, err
	}, hookstage.MutationUpdate, "request-body-with-profile-data")

	result.Reject = false
	return result, nil
}

// applyProfileChanges copies and updates BidRequest with required values from http header and partnetConfigMap
func (m *OpenWrap) applyProfileChanges(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest) (*openrtb2.BidRequest, error) {
	if cur, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID][models.AdServerCurrency]; ok {
		bidRequest.Cur = []string{cur}
	}

	if bidRequest.TMax == 0 {
		bidRequest.TMax = m.setTimeout(rctx)
	}

	if bidRequest.Source == nil {
		bidRequest.Source = &openrtb2.Source{}
	}

	bidRequest.Source.TID = bidRequest.ID

	for i := 0; i < len(bidRequest.Imp); i++ {
		m.applyBannerAdUnitConfig(rctx, &bidRequest.Imp[i])
		m.applyVideoAdUnitConfig(rctx, &bidRequest.Imp[i])
		bidRequest.Imp[i].Ext = rctx.ImpBidCtx[bidRequest.Imp[i].ID].NewExt
	}

	if rctx.Platform == models.PLATFORM_APP || rctx.Platform == models.PLATFORM_VIDEO {
		sChainObj := getSChainObj(rctx.PartnerConfigMap)
		if sChainObj != nil {
			setSchainInSourceObject(bidRequest.Source, sChainObj)
		}
	}

	adunitconfig.ReplaceAppObjectFromAdUnitConfig(rctx, bidRequest.App)
	adunitconfig.ReplaceDeviceTypeFromAdUnitConfig(rctx, bidRequest.Device)
	bidRequest.Device.IP = rctx.IP

	bidRequest.Ext = rctx.NewReqExt
	return bidRequest, nil
}

func (m *OpenWrap) applyVideoAdUnitConfig(rCtx models.RequestCtx, imp *openrtb2.Imp) {
	adUnitCfg := rCtx.ImpBidCtx[imp.ID].VideoAdUnitCtx.AppliedSlotAdUnitConfig
	if adUnitCfg == nil {
		return
	}

	if imp.BidFloor == 0 && adUnitCfg.BidFloor != nil {
		imp.BidFloor = *adUnitCfg.BidFloor
	}

	if len(imp.BidFloorCur) == 0 && adUnitCfg.BidFloorCur != nil {
		imp.BidFloorCur = *adUnitCfg.BidFloorCur
	}

	if adUnitCfg.Exp != nil {
		imp.Exp = int64(*adUnitCfg.Exp)
	}

	if adUnitCfg.Video == nil {
		return
	}

	//check if video is disabled, if yes then remove video from imp object
	if adUnitCfg.Video.Enabled != nil && !*adUnitCfg.Video.Enabled {
		imp.Video = nil
		return
	}

	if adUnitCfg.Video.Config == nil {
		return
	}

	configObjInVideoConfig := adUnitCfg.Video.Config

	if len(imp.Video.MIMEs) == 0 {
		imp.Video.MIMEs = configObjInVideoConfig.MIMEs
	}

	if imp.Video.MinDuration == 0 {
		imp.Video.MinDuration = configObjInVideoConfig.MinDuration
	}

	if imp.Video.MaxDuration == 0 {
		imp.Video.MaxDuration = configObjInVideoConfig.MaxDuration
	}

	if imp.Video.Skip == nil {
		imp.Video.Skip = configObjInVideoConfig.Skip
	}

	if imp.Video.SkipMin == 0 {
		imp.Video.SkipMin = configObjInVideoConfig.SkipMin
	}

	if imp.Video.SkipAfter == 0 {
		imp.Video.SkipAfter = configObjInVideoConfig.SkipAfter
	}

	if len(configObjInVideoConfig.BAttr) == 0 {
		imp.Video.BAttr = configObjInVideoConfig.BAttr
	}

	if imp.Video.MinBitRate == 0 {
		imp.Video.MinBitRate = configObjInVideoConfig.MinBitRate
	}

	if imp.Video.MaxBitRate == 0 {
		imp.Video.MaxBitRate = configObjInVideoConfig.MaxBitRate
	}

	if imp.Video.MaxExtended == 0 {
		imp.Video.MaxExtended = configObjInVideoConfig.MaxExtended
	}

	if imp.Video.StartDelay == nil {
		imp.Video.StartDelay = configObjInVideoConfig.StartDelay
	}

	if imp.Video.Placement == 0 {
		imp.Video.Placement = configObjInVideoConfig.Placement
	}

	if imp.Video.Linearity == 0 {
		imp.Video.Linearity = configObjInVideoConfig.Linearity
	}

	if imp.Video.Protocol == 0 {
		imp.Video.Protocol = configObjInVideoConfig.Protocol
	}

	if len(configObjInVideoConfig.Protocols) == 0 {
		imp.Video.Protocols = configObjInVideoConfig.Protocols
	}

	if imp.Video.W == 0 {
		imp.Video.W = configObjInVideoConfig.W
	}

	if imp.Video.H == 0 {
		imp.Video.H = configObjInVideoConfig.H
	}

	if imp.Video.Sequence == 0 {
		imp.Video.Sequence = configObjInVideoConfig.Sequence
	}

	if imp.Video.BoxingAllowed == 0 {
		imp.Video.BoxingAllowed = configObjInVideoConfig.BoxingAllowed
	}

	if imp.Video.PlaybackMethod == nil && len(configObjInVideoConfig.PlaybackMethod) > 0 {
		imp.Video.PlaybackMethod = configObjInVideoConfig.PlaybackMethod
	}

	if imp.Video.PlaybackEnd == 0 {
		imp.Video.PlaybackEnd = configObjInVideoConfig.PlaybackEnd
	}

	if imp.Video.Delivery == nil {
		imp.Video.Delivery = configObjInVideoConfig.Delivery
	}

	if imp.Video.Pos == nil {
		imp.Video.Pos = configObjInVideoConfig.Pos
	}

	if len(configObjInVideoConfig.API) > 0 {
		imp.Video.API = configObjInVideoConfig.API
	}

	if len(configObjInVideoConfig.CompanionType) > 0 {
		imp.Video.CompanionType = configObjInVideoConfig.CompanionType
	}

	if imp.Video.CompanionAd == nil {
		imp.Video.CompanionAd = configObjInVideoConfig.CompanionAd
	}
}

func (m *OpenWrap) applyBannerAdUnitConfig(rCtx models.RequestCtx, imp *openrtb2.Imp) {
	adUnitCfg := rCtx.ImpBidCtx[imp.ID].VideoAdUnitCtx.AppliedSlotAdUnitConfig
	if adUnitCfg == nil {
		return
	}

	if imp.BidFloor == 0 && adUnitCfg.BidFloor != nil {
		imp.BidFloor = *adUnitCfg.BidFloor
	}

	if len(imp.BidFloorCur) == 0 && adUnitCfg.BidFloorCur != nil {
		imp.BidFloorCur = *adUnitCfg.BidFloorCur
	}

	if adUnitCfg.Exp != nil {
		imp.Exp = int64(*adUnitCfg.Exp)
	}

	if adUnitCfg.Banner == nil {
		return
	}

	if adUnitCfg.Banner.Enabled != nil && !*adUnitCfg.Banner.Enabled {
		imp.Banner = nil
		return
	}
}

func getDomainFromUrl(pageUrl string) string {
	u, err := url.Parse(pageUrl)
	if err != nil {
		return ""
	}

	return u.Host
}

// always perfer rCtx.LoggerImpressionID received in request. Create a new once if it is not availble.
// func getLoggerID(reqExt models.ExtRequestWrapper) string {
// 	if reqExt.Wrapper.LoggerImpressionID != "" {
// 		return reqExt.Wrapper.LoggerImpressionID
// 	}
// 	return uuid.NewV4().String()
// }

// NYC: make this generic. Do we need this?. PBS now has auto_gen_source_tid generator. We can make it to wiid for pubmatic adapter in pubmatic.go
func updateRequestExtBidderParamsPubmatic(bidderParams json.RawMessage, cookie, loggerID, bidderCode string) (json.RawMessage, error) {
	bidderParamsMap := make(map[string]map[string]interface{})
	_ = json.Unmarshal(bidderParams, &bidderParamsMap) // ignore error, incoming might be nil for now but we still have data to put

	bidderParamsMap[bidderCode] = map[string]interface{}{
		models.WrapperLoggerImpID: loggerID,
	}

	if len(cookie) != 0 {
		bidderParamsMap[bidderCode][models.COOKIE] = cookie
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
		string(models.MacroProfileID):           fmt.Sprintf("%d", rctx.ProfileID),
		string(models.MacroProfileVersionID):    fmt.Sprintf("%d", rctx.DisplayID),
		string(models.MacroUnixTimeStamp):       fmt.Sprintf("%d", rctx.StartTime),
		string(models.MacroPlatform):            fmt.Sprintf("%d", rctx.DevicePlatform),
		string(models.MacroSSAI):                rctx.SSAI,
		string(models.MacroWrapperImpressionID): rctx.LoggerImpressionID,
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

// setTimeout - This utility returns timeout applicable for a profile
func (m OpenWrap) setTimeout(rCtx models.RequestCtx) int64 {
	var auctionTimeout int64

	//check for ssTimeout in the partner config
	ssTimeout := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.SSTimeoutKey)
	if ssTimeout != "" {
		ssTimeoutDB, err := strconv.Atoi(ssTimeout)
		if err == nil {
			auctionTimeout = int64(ssTimeoutDB)
		}
	}

	// found tmax value in request or db
	if auctionTimeout != 0 {
		if auctionTimeout < m.cfg.OpenWrap.Timeout.MinTimeout {
			return m.cfg.OpenWrap.Timeout.MinTimeout
		} else if auctionTimeout > m.cfg.OpenWrap.Timeout.MaxTimeout {
			return m.cfg.OpenWrap.Timeout.MaxTimeout
		}
		return auctionTimeout
	}

	//Below piece of code is applicable for older profiles where ssTimeout is not set
	//Here we will check the partner timeout and select max timeout considering timeout range
	auctionTimeout = m.cfg.OpenWrap.Timeout.MinTimeout
	for _, partnerConfig := range rCtx.PartnerConfigMap {
		partnerTO, _ := strconv.Atoi(partnerConfig[models.TIMEOUT])
		if int64(partnerTO) > m.cfg.OpenWrap.Timeout.MaxTimeout {
			auctionTimeout = m.cfg.OpenWrap.Timeout.MaxTimeout
			break
		}
		if int64(partnerTO) >= m.cfg.OpenWrap.Timeout.MinTimeout {
			if auctionTimeout < int64(partnerTO) {
				auctionTimeout = int64(partnerTO)
			}
		}
	}
	return auctionTimeout
}

// isSendAllBids returns true in below cases:
// if ssauction flag is set 0 in the request
// if ssauction flag is not set and platform is dislay, then by default send all bids
// if ssauction flag is not set and platform is in-app, then check if profile setting sendAllBids is set to 1
func isSendAllBids(rctx models.RequestCtx) bool {

	//if ssauction is set to 0 in the request
	if rctx.SSAuction == 0 {
		return true
	} else if rctx.SSAuction == -1 {
		//when ssauction flag is not present in request

		//if platform is dislay, then by default send all bids is enabled
		if rctx.Platform == models.PLATFORM_DISPLAY {
			return true
		}

		//if platform is in-app, then check if profile setting sendAllBids is set to 1
		if rctx.Platform == models.PLATFORM_APP {
			if models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.SendAllBidsKey) == "1" {
				return true
			}
		}
	}
	return false
}
