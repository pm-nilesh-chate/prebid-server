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
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (m OpenWrap) handleBeforeValidationHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	result := hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{}

	requestExt, err := models.GetRequestExt(payload.BidRequest.Ext)
	if err != nil {
		result.Reject = true
		result.NbrCode = errorcodes.ErrInvalidRequestExtension.Code()
		result.Errors = append(result.Errors, errorcodes.ErrInvalidRequestExtension.Error())
		return result, err
	}

	rCtx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	rCtx.IsTestRequest = payload.BidRequest.Test == 2

	partnerConfigMap, err := m.getProfileData(rCtx)
	if err != nil || len(partnerConfigMap) == 0 {
		result.Reject = true
		result.NbrCode = errorcodes.ErrInvalidConfiguration.Code()
		result.DebugMessages = append(result.Errors, errorcodes.ErrInvalidConfiguration.Error())
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

	rCtx.AdapterThrottleMap, err = GetAdapterThrottleMap(rCtx.PartnerConfigMap)
	if err != nil {
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

	requestExt.Prebid.SupportDeals = rCtx.PreferDeals // && IsCTVAPIRequest(reqWrapper.RequestAPI),
	requestExt.Prebid.AlternateBidderCodes = getMarketplaceBidders(requestExt.Prebid.AlternateBidderCodes, partnerConfigMap)
	requestExt.Prebid.Targeting = &openrtb_ext.ExtRequestTargeting{
		IncludeBidderKeys: true,
		IncludeWinners:    true,
	}
	requestExt.Prebid.BidderParams, _ = updateRequestExtBidderParamsPubmatic(requestExt.Prebid.BidderParams, rCtx.Cookies, rCtx.LoggerImpressionID, string(openrtb_ext.BidderPubmatic))

	newImps := make(map[string]openrtb2.Imp)
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

		if rCtx.AdUnitConfig != nil {
			// NYC TODO
			// rCtx.AdUnitConfigMatchedSlot, deducedAdUnitConfig = getMatchedSlotName(rCtx, imp, *impExt)
			updateVideoObjectWithAdunitConfig(rCtx, imp, *impExt, payload.BidRequest.Device.ConnectionType)
			updateBannerObjectWithAdunitConfig(rCtx, imp, *impExt)
		}

		if imp.Banner == nil && imp.Video == nil && imp.Native == nil {
			payload.BidRequest.Imp = append(payload.BidRequest.Imp[:i], payload.BidRequest.Imp[i+1:]...)
			result.Errors = append(result.Errors, fmt.Sprintf("no Valid Banner/Video/Native present for imp: %+v", imp.ID))
			i--
			continue
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
					Bidders:           make(map[string]models.PartnerData),
					BidCtx:            make(map[string]models.BidCtx),
				}
			}

			rCtx.ImpBidCtx[imp.ID].Bidders[bidderCode] = models.PartnerData{
				Params:    bidderParams,
				PartnerID: partnerID,
			}

			if alias, ok := partnerConfig[models.IsAlias]; ok && alias == "1" {
				if prebidPartnerName, ok := partnerConfig[models.PREBID_PARTNER_NAME]; ok {
					rCtx.Aliases[bidderCode] = adapters.ResolveOWBidder(prebidPartnerName)
				}
			}

			if partnerConfig[models.PREBID_PARTNER_NAME] == models.BidderVASTBidder {
				updateAliasGVLIds(aliasgvlids, bidderCode, partnerConfig)
			}
		} // for(rctx.PartnerConfigMap

		if cto := setContentTransparencyObject(rCtx, requestExt, imp.ID, rCtx.AdUnitConfig); cto != nil {
			requestExt.Prebid.Transparency = cto
		}
	} // for(imp

	// replaceAppObjectFromAdUnitConfig(reqWrapper.AdUnitConfig, newReq)

	if len(rCtx.Aliases) != 0 && requestExt.Prebid.Aliases == nil {
		requestExt.Prebid.Aliases = make(map[string]string)
	}
	for k, v := range rCtx.Aliases {
		requestExt.Prebid.Aliases[k] = v
	}

	requestExt.Prebid.AliasGVLIDs = aliasgvlids

	result.ChangeSet.AddMutation(func(ep hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		rctx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		ep.BidRequest, err = m.applyProfileChanges(rctx, ep.BidRequest, requestExt, newImps)
		return ep, err
	}, hookstage.MutationUpdate, "request-body-with-profile-data")

	moduleCtx.ModuleContext["rctx"] = rCtx
	return result, nil
}

// applyProfileChanges copies and updates BidRequest with required values from http header and partnetConfigMap
func (m *OpenWrap) applyProfileChanges(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest, requestExt models.RequestExt, imp map[string]openrtb2.Imp) (*openrtb2.BidRequest, error) {
	if cur, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID][models.AdServerCurrency]; ok {
		bidRequest.Cur = []string{cur}
	}

	var err error
	for i := 0; i < len(bidRequest.Imp); i++ {

		bidRequest.Imp[i].Ext, err = json.Marshal(imp[bidRequest.Imp[i].ID].Ext)
		if err != nil {
			return bidRequest, err
		}
	}

	bidRequest.Ext, err = json.Marshal(requestExt)
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

	if v, ok := adUnitConfigMap.Config[models.AdunitConfigDefaultKey]; ok && v.Video != nil && v.Video.Config != nil && len(v.Video.Config.CompanionType) != 0 {
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
		string(models.MacroUnixTimeStamp):       fmt.Sprintf("%d", rctx.StartTime, 10),
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
