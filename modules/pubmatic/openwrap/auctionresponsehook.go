package openwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookanalytics"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) handleAuctionResponseHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	result := hookstage.HookResult[hookstage.AuctionResponsePayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.AuctionResponsePayload]{}

	// absence of rctx at this hook means the first hook failed!. Do nothing
	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	rctx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	defer func() {
		moduleCtx.ModuleContext["rctx"] = rctx
	}()

	// cache rctx for analytics
	result.AnalyticsTags.Activities = make([]hookanalytics.Activity, 1)
	result.AnalyticsTags.Activities[0].Name = "openwrap_request_ctx"
	result.AnalyticsTags.Activities[0].Results = make([]hookanalytics.Result, 1)
	values := make(map[string]interface{})
	values["request-ctx"] = &rctx
	result.AnalyticsTags.Activities[0].Results[0].Values = values

	winningBids := make(map[string]models.OwBid, 0)
	for _, seatBid := range payload.BidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impCtx, ok := rctx.ImpBidCtx[bid.ImpID]
			if !ok {
				result.Errors = append(result.Errors, "invalid impCtx.ID for bid"+bid.ImpID)
				continue
			}

			partnerID := 0
			if bidderMeta, ok := impCtx.Bidders[seatBid.Seat]; ok {
				partnerID = bidderMeta.PartnerID
			}

			revShare := models.GetRevenueShare(rctx.PartnerConfigMap[partnerID])
			price := bid.Price

			bidExt := &models.BidExt{}
			if len(bid.Ext) != 0 { //NYC_TODO: most of the fields should be filled even if unmarshal fails
				err := json.Unmarshal(bid.Ext, bidExt)
				if err != nil {
					result.Errors = append(result.Errors, "failed to unmarshal bid.ext for "+bid.ID)
					// continue
				}

				// NYC_TODO: fix this in PBS-Core or ExecuteAllProcessedBidResponsesStage
				if bidExt.Prebid != nil && bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration == 0 &&
					bidExt.Prebid.Video.PrimaryCategory == "" && bidExt.Prebid.Video.VASTTagID == "" {
					bidExt.Prebid.Video = nil
				}

				if v, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID]["refreshInterval"]; ok {
					n, err := strconv.Atoi(v)
					if err == nil {
						bidExt.RefreshInterval = n
					}
				}

				if bidExt.Prebid != nil {
					bidExt.CreativeType = string(bidExt.Prebid.Type)
				}
				if bidExt.CreativeType == "" {
					bidExt.CreativeType = models.GetAdFormat(bid.AdM)
				}

				if payload.BidResponse.Cur != "USD" {
					price = bidExt.OriginalBidCPMUSD
				}

				bidExt.NetECPM = models.GetNetEcpm(price, revShare)

				if impCtx.Video != nil && impCtx.Type == "video" && bidExt.CreativeType == "video" {
					if bidExt.Video == nil {
						bidExt.Video = &models.ExtBidVideo{}
					}
					if impCtx.Video.MaxDuration != 0 {
						bidExt.Video.MaxDuration = impCtx.Video.MaxDuration
					}
					if impCtx.Video.MinDuration != 0 {
						bidExt.Video.MinDuration = impCtx.Video.MinDuration
					}
					if impCtx.Video.Skip != nil {
						bidExt.Video.Skip = impCtx.Video.Skip
					}
					if impCtx.Video.SkipAfter != 0 {
						bidExt.Video.SkipAfter = impCtx.Video.SkipAfter
					}
					if impCtx.Video.SkipMin != 0 {
						bidExt.Video.SkipMin = impCtx.Video.SkipMin
					}
					bidExt.Video.BAttr = impCtx.Video.BAttr
					bidExt.Video.PlaybackMethod = impCtx.Video.PlaybackMethod
					if rctx.ClientConfigFlag == 1 {
						bidExt.Video.ClientConfig = adunitconfig.GetClientConfigForMediaType(rctx, bid.ImpID, "video")
					}
				} else if rctx.ImpBidCtx[bid.ImpID].Type == "banner" && bidExt.CreativeType == "banner" && rctx.ClientConfigFlag == 1 {
					cc := adunitconfig.GetClientConfigForMediaType(rctx, bid.ImpID, "banner")
					if len(cc) != 0 {
						if bidExt.Banner == nil {
							bidExt.Banner = &models.ExtBidBanner{}
						}
						bidExt.Banner.ClientConfig = cc
					}
				}
			}

			bidDealTierSatisfied := false
			if bidExt.Prebid != nil {
				bidDealTierSatisfied = bidExt.Prebid.DealTierSatisfied
			}

			owbid := models.OwBid{
				Bid:                  &bid,
				NetEcpm:              bidExt.NetECPM,
				BidDealTierSatisfied: bidDealTierSatisfied,
			}
			wbid, ok := winningBids[bid.ImpID]
			if !ok || isNewWinningBid(owbid, wbid, rctx.PreferDeals) {
				winningBids[owbid.ImpID] = owbid
			}

			// cache for bid details for logger and tracker
			if impCtx.BidCtx == nil {
				impCtx.BidCtx = make(map[string]models.BidCtx)
			}
			impCtx.BidCtx[bid.ID] = models.BidCtx{
				BidExt: *bidExt,
			}
			rctx.ImpBidCtx[bid.ImpID] = impCtx
		}
	}

	rctx.WinningBids = winningBids

	droppedBids, warnings := addPWTTargetingForBid(rctx, payload.BidResponse)
	if len(droppedBids) != 0 {
		rctx.DroppedBids = droppedBids
	}
	if len(warnings) != 0 {
		result.Warnings = append(result.Warnings, warnings...)
	}

	responseExt := make(map[string]interface{})
	// TODO use concrete structure
	if len(payload.BidResponse.Ext) != 0 {
		if err := json.Unmarshal(payload.BidResponse.Ext, &responseExt); err != nil {
			result.Errors = append(result.Errors, "failed to unmarshal response.ext err: "+err.Error())
		}
	}

	// TODO: PBS-Core should pass the hostcookie for module to usersync.ParseCookieFromRequest()
	if matchedImpression := getMatchedImpression(rctx); matchedImpression != nil {
		responseExt[models.MatchedImpression] = matchedImpression
	}

	if rctx.SendAllBids {
		responseExt[models.SendAllBidsFlagKey] = 1
	}

	var err error
	rctx.ResponseExt, err = json.Marshal(responseExt)
	if err != nil {
		result.Errors = append(result.Errors, "failed to marshal response.ext err: "+err.Error())
	}

	result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
		rctx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		ap.BidResponse, err = m.updateORTBV25Response(rctx, ap.BidResponse)
		if err != nil {
			return ap, err
		}
		ap.BidResponse, err = m.injectTrackers(rctx, ap.BidResponse)
		if err != nil {
			return ap, err
		}

		ap.BidResponse, err = m.addDefaultBids(rctx, ap.BidResponse)

		ap.BidResponse.Ext = rctx.ResponseExt
		return ap, err
	}, hookstage.MutationUpdate, "response-body-with-sshb-format")

	// TODO: move debug here
	// result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
	// }, hookstage.MutationUpdate, "response-body-with-sshb-format")

	return result, nil
}

func (m *OpenWrap) updateORTBV25Response(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	if len(bidResponse.SeatBid) == 0 {
		return bidResponse, nil
	}

	// remove non-winning bids if sendallbids=1
	for _, seatBid := range bidResponse.SeatBid {
		filteredBid := make([]openrtb2.Bid, 0, len(seatBid.Bid))
		for _, bid := range seatBid.Bid {
			if b, ok := rctx.WinningBids[bid.ImpID]; ok && b.ID == bid.ID && rctx.SendAllBids {
				filteredBid = append(filteredBid, bid)
			}
		}
		seatBid.Bid = filteredBid
	}

	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			impCtx, ok := rctx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			bidResponse.SeatBid[i].Bid[j].Ext, _ = json.Marshal(bidCtx.BidExt)
		}
	}

	// remove seats with empty bids (will add nobids later)
	filteredSeatBid := make([]openrtb2.SeatBid, 0, len(bidResponse.SeatBid))
	for _, seatBid := range bidResponse.SeatBid {
		if len(seatBid.Bid) > 0 {
			filteredSeatBid = append(filteredSeatBid, seatBid)
		}
	}
	bidResponse.SeatBid = filteredSeatBid

	if len(bidResponse.SeatBid) != 0 {
		// keep pubmatic 1st to handle automation failure.
		if bidResponse.SeatBid[0].Seat != "pubmatic" {
			for i := 0; i < len(bidResponse.SeatBid); i++ {
				if bidResponse.SeatBid[i].Seat == "pubmatic" {
					temp := bidResponse.SeatBid[0]
					bidResponse.SeatBid[0] = bidResponse.SeatBid[i]
					bidResponse.SeatBid[i] = temp
				}
			}
		}
	}

	return bidResponse, nil
}

// isNewWinningBid calculates if the new bid (nbid) will win against the current winning bid (wbid) given preferDeals.
func isNewWinningBid(bid, wbid models.OwBid, preferDeals bool) bool {
	if preferDeals {
		//only wbid has deal
		if wbid.BidDealTierSatisfied && !bid.BidDealTierSatisfied {
			return false
		}
		//only bid has deal
		if !wbid.BidDealTierSatisfied && bid.BidDealTierSatisfied {
			return true
		}
	}
	//both have deal or both do not have deal
	return bid.NetEcpm > wbid.NetEcpm
}

func getPlatformName(platform string) string {
	if platform == models.PLATFORM_APP {
		return models.PlatformAppTargetingKey
	}
	return platform
}

func getIntPtr(i int) *int {
	return &i
}

func addPWTTargetingForBid(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (droppedBids map[string][]openrtb2.Bid, warnings []string) {
	if rctx.Platform != models.PLATFORM_APP {
		return
	}

	if !rctx.SendAllBids {
		droppedBids = make(map[string][]openrtb2.Bid)
	}

	//setTargeting needs a seperate loop as final winner would be decided after all the bids are processed by auction
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impCtx, ok := rctx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			isWinningBid := false
			if b, ok := rctx.WinningBids[bid.ImpID]; ok && b.ID == bid.ID {
				isWinningBid = true
			}

			if !(isWinningBid || rctx.SendAllBids) {
				droppedBids[seatBid.Seat] = append(droppedBids[seatBid.Seat], bid)
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			newTargeting := make(map[string]string)
			for key, value := range bidCtx.Prebid.Targeting {
				if allowTargetingKey(key) {
					updatedKey := key
					if strings.HasPrefix(key, models.PrebidTargetingKeyPrefix) {
						updatedKey = strings.Replace(key, models.PrebidTargetingKeyPrefix, models.OWTargetingKeyPrefix, 1)
					}
					newTargeting[updatedKey] = value
				}
				delete(bidCtx.Prebid.Targeting, key)
			}

			bidCtx.Prebid.Targeting = newTargeting
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_SLOTID)] = bid.ID
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_SZ)] = models.GetSize(bid.W, bid.H)
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_PARTNERID)] = seatBid.Seat
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_ECPM)] = fmt.Sprintf("%.2f", bidCtx.NetECPM)
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_PLATFORM)] = getPlatformName(rctx.Platform)
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_BIDSTATUS)] = "1"
			if len(bid.DealID) != 0 {
				bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_DEALID)] = bid.DealID
			}

			if isWinningBid {
				if rctx.SendAllBids {
					bidCtx.Winner = 1
				}

				bidCtx.Prebid.Targeting[models.PWT_SLOTID] = bid.ID
				bidCtx.Prebid.Targeting[models.PWT_BIDSTATUS] = "1"
				bidCtx.Prebid.Targeting[models.PWT_SZ] = models.GetSize(bid.W, bid.H)
				bidCtx.Prebid.Targeting[models.PWT_PARTNERID] = seatBid.Seat
				bidCtx.Prebid.Targeting[models.PWT_ECPM] = fmt.Sprintf("%.2f", bidCtx.NetECPM)
				bidCtx.Prebid.Targeting[models.PWT_PLATFORM] = getPlatformName(rctx.Platform)
				if len(bid.DealID) != 0 {
					bidCtx.Prebid.Targeting[models.PWT_DEALID] = bid.DealID
				}
			} else if !rctx.SendAllBids {
				warnings = append(warnings, "dropping bid "+bid.ID+" as sendAllBids is disabled")
			}

			// cache for bid details for logger and tracker
			if impCtx.BidCtx == nil {
				impCtx.BidCtx = make(map[string]models.BidCtx)
			}
			impCtx.BidCtx[bid.ID] = bidCtx
			rctx.ImpBidCtx[bid.ImpID] = impCtx
		}
	}
	return
}

func (m *OpenWrap) addDefaultBids(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	// responded bidders per impression
	seatBids := make(map[string]map[string]struct{}, len(bidResponse.SeatBid))
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if seatBids[bid.ImpID] == nil {
				seatBids[bid.ImpID] = make(map[string]struct{})
			}
			seatBids[bid.ImpID][seatBid.Seat] = struct{}{}
		}
	}

	// included dropped bids in responded to avoid false nobid entry.
	for seat, bids := range rctx.DroppedBids {
		for _, bid := range bids {
			if seatBids[bid.ImpID] == nil {
				seatBids[bid.ImpID] = make(map[string]struct{})
			}
			seatBids[bid.ImpID][seat] = struct{}{}
		}
	}

	// bids per bidders per impression that did not respond
	noSeatBids := make(map[string]map[string][]openrtb2.Bid, 0)
	for impID, impCtx := range rctx.ImpBidCtx {
		for bidder := range impCtx.Bidders {
			noBid := false
			if bidders, ok := seatBids[impID]; ok {
				if _, ok := bidders[bidder]; !ok {
					noBid = true
				}
			} else {
				noBid = true
			}

			if noBid {
				if noSeatBids[impID] == nil {
					noSeatBids[impID] = make(map[string][]openrtb2.Bid)
				}

				noSeatBids[impID][bidder] = append(noSeatBids[impID][bidder], openrtb2.Bid{
					ID:    impID,
					ImpID: impID,
					Ext:   newNoBidExt(rctx, impID),
				})
			}
		}
	}

	// add nobids for throttled adapter to all the impressions (how do we change bidders at impression level?)
	for bidder := range rctx.AdapterThrottleMap {
		for impID, _ := range rctx.ImpBidCtx { // ImpBidCtx is used only for list of impID, it does not have data of throttled adapters
			if noSeatBids[impID] == nil {
				noSeatBids[impID] = make(map[string][]openrtb2.Bid)
			}

			noSeatBids[impID][bidder] = []openrtb2.Bid{
				{
					ID:    impID,
					ImpID: impID,
					Ext:   newNoBidExt(rctx, impID),
				},
			}
		}
	}

	// add nobids for non-mapped bidders
	for impID, impCtx := range rctx.ImpBidCtx {
		for bidder := range impCtx.NonMapped {
			if noSeatBids[impID] == nil {
				noSeatBids[impID] = make(map[string][]openrtb2.Bid)
			}

			noSeatBids[impID][bidder] = []openrtb2.Bid{
				{
					ID:    impID,
					ImpID: impID,
					Ext:   newNoBidExt(rctx, impID),
				},
			}
		}
	}

	// update nobids in final response
	for i, seatBid := range bidResponse.SeatBid {
		for impID, noSeatBid := range noSeatBids {
			for seat, bids := range noSeatBid {
				if seatBid.Seat == seat {
					bidResponse.SeatBid[i].Bid = append(bidResponse.SeatBid[i].Bid, bids...)
					delete(noSeatBid, seat)
					noSeatBids[impID] = noSeatBid
				}
			}
		}
	}

	// no-seat case
	for _, noSeatBid := range noSeatBids {
		for seat, bids := range noSeatBid {
			bidResponse.SeatBid = append(bidResponse.SeatBid, openrtb2.SeatBid{
				Bid:  bids,
				Seat: seat,
			})
		}
	}

	return bidResponse, nil
}

func newNoBidExt(rctx models.RequestCtx, impID string) json.RawMessage {
	bidExt := models.BidExt{
		NetECPM: 0,
	}
	if rctx.ClientConfigFlag == 1 {
		bidExt.Banner = &models.ExtBidBanner{
			ClientConfig: adunitconfig.GetClientConfigForMediaType(rctx, impID, "banner"),
		}
		bidExt.Video = &models.ExtBidVideo{
			ClientConfig: adunitconfig.GetClientConfigForMediaType(rctx, impID, "video"),
		}
	}

	if v, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID]["refreshInterval"]; ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			bidExt.RefreshInterval = n
		}
	}

	newBidExt, err := json.Marshal(bidExt)
	if err != nil {
		return nil
	}

	return json.RawMessage(newBidExt)
}
