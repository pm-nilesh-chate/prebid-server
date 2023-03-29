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
		return result, nil
	}
	rctx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		return result, nil
	}

	result.AnalyticsTags.Activities = make([]hookanalytics.Activity, 1)
	result.AnalyticsTags.Activities[0].Name = "openwrap_request_ctx"
	result.AnalyticsTags.Activities[0].Results = make([]hookanalytics.Result, 1)
	values := make(map[string]interface{})
	values["request-ctx"] = &rctx
	result.AnalyticsTags.Activities[0].Results[0].Values = values

	result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
		rctx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		ap.BidResponse, err = m.updateORTBV25Response(rctx, ap.BidResponse)
		if err != nil {
			return ap, err
		}
		ap.BidResponse, err = m.injectTrackers(rctx, ap.BidResponse)
		return ap, err
	}, hookstage.MutationUpdate, "response-body-with-sshb-format")

	return result, nil
}

type owBid struct {
	*openrtb2.Bid
	netEcpm              float64
	bidDealTierSatisfied bool
}

func (m *OpenWrap) updateORTBV25Response(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	if len(bidResponse.SeatBid) == 0 {
		return bidResponse, nil
	}

	winningBids := make(map[string]owBid, 0)
	// winningBidsByBidder := make(map[string]map[openrtb_ext.BidderName]owBid, 0)
	partnerNameMap := make(map[string]map[string]string)

	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			// NYC_TODO maintain a global map of ow-partner-id to biddercode. Ex. 8->pubmatic
			// prepare partner name to partner config map

			for _, partnerConfig := range rctx.PartnerConfigMap {
				if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
					continue
				}
				partnerNameMap[partnerConfig[models.BidderCode]] = partnerConfig
			}

			bidExt := &models.BidExt{}
			if len(bid.Ext) != 0 { //NYC_TODO: most of the fields should be filled even if unmarshal fails
				err := json.Unmarshal(bid.Ext, bidExt)
				if err != nil {
					return bidResponse, err
				}

				if v, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID]["refreshInterval"]; ok {
					n, err := strconv.Atoi(v)
					if err == nil {
						bidExt.RefreshInterval = n
					}
				}

				bidExt.CreativeType = string(bidExt.Prebid.Type)
				if bidExt.CreativeType == "" {
					bidExt.CreativeType = models.GetAdFormat(bid.AdM)
				}

				// bidExt.Summary

				revShare := models.GetRevenueShare(partnerNameMap[seatBid.Seat])
				price := bid.Price
				if bidResponse.Cur != "USD" {
					price = bidExt.OriginalBidCPMUSD
				}

				// if platform == models.PLATFORM_APP {
				bidExt.NetECPM = models.GetNetEcpm(price, revShare)
				// bidExt.Prebid = addPWTTargetingForBid(*request.Id, eachBid, impExt.Prebid, *eachSeatBid.Seat, platform, winBidFlag, netEcpm)
				// }

				if rctx.ClientConfigFlag == 1 {
					if rctx.ImpBidCtx[bid.ImpID].Type == "banner" {
						if bidExt.Banner == nil {
							bidExt.Banner = &models.ExtBidBanner{}
						}
						bidExt.Banner.ClientConfig = adunitconfig.GetClientConfigForMediaType(rctx, bid.ImpID, rctx.AdUnitConfig, "banner")
					} else if rctx.ImpBidCtx[bid.ImpID].Type == "video" {
						if bidExt.Video == nil {
							bidExt.Video = &models.ExtBidVideo{}
						}
						bidExt.Video.ClientConfig = adunitconfig.GetClientConfigForMediaType(rctx, bid.ImpID, rctx.AdUnitConfig, "video")
					}
				}
			}

			owbid := owBid{&bid, bidExt.NetECPM, bidExt.Prebid.DealTierSatisfied}
			wbid, ok := winningBids[bid.ImpID]
			if !ok || isNewWinningBid(owbid, wbid, rctx.PreferDeals) {
				winningBids[owbid.ImpID] = owbid
			}
			// if bidMap, ok := winningBidsByBidder[owbid.ImpID]; ok {
			// 	bestSoFar, ok := bidMap[openrtb_ext.BidderName(seatBid.Seat)]
			// 	if !ok || cpm > bestSoFar.Bid.Price {
			// 		bidMap[bidderName] = bid
			// 	}
			// } else {
			// 	winningBidsByBidder[bid.Bid.ImpID] = make(map[openrtb_ext.BidderName]*entities.PbsOrtbBid)
			// 	winningBidsByBidder[bid.Bid.ImpID][bidderName] = bid
			// }

			var err error
			bidResponse.SeatBid[i].Bid[j].Ext, err = json.Marshal(bidExt)
			if err != nil {
				return bidResponse, err
			}
		}
	}

	//setTargeting
	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			bidExt := &models.BidExt{}
			if len(bid.Ext) != 0 {
				err := json.Unmarshal(bid.Ext, bidExt)
				if err != nil {
					return bidResponse, err
				}
			}

			revShare := models.GetRevenueShare(partnerNameMap[seatBid.Seat])
			netEcpm := models.GetNetEcpm(bid.Price, revShare)

			newTargeting := make(map[string]string)
			for key, value := range bidExt.Prebid.Targeting {
				if allowTargetingKey(key) {
					updatedKey := key
					if strings.HasPrefix(key, models.PrebidTargetingKeyPrefix) {
						updatedKey = strings.Replace(key, models.PrebidTargetingKeyPrefix, models.OWTargetingKeyPrefix, 1)
					}
					newTargeting[updatedKey] = value
				}
				delete(bidExt.Prebid.Targeting, key)
			}

			bidExt.Prebid.Targeting = newTargeting
			bidExt.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_SLOTID)] = bid.ID
			bidExt.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_SZ)] = models.GetSize(bid.W, bid.H)
			bidExt.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_PARTNERID)] = seatBid.Seat
			bidExt.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_ECPM)] = fmt.Sprintf("%.2f", netEcpm)
			bidExt.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_PLATFORM)] = getPlatformName(rctx.Platform)
			bidExt.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_BIDSTATUS)] = "1"
			if len(bid.DealID) != 0 {
				bidExt.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_DEALID)] = bid.DealID
			}

			if b, ok := winningBids[bid.ImpID]; ok && b.ID == bid.ID {
				// bidExt.Winner = ptrutil.ToPtr(1)
				bidExt.Winner = 1

				bidExt.Prebid.Targeting[models.PWT_SLOTID] = bid.ID
				bidExt.Prebid.Targeting[models.PWT_BIDSTATUS] = "1"
				bidExt.Prebid.Targeting[models.PWT_SZ] = models.GetSize(bid.W, bid.H)
				bidExt.Prebid.Targeting[models.PWT_PARTNERID] = seatBid.Seat
				bidExt.Prebid.Targeting[models.PWT_ECPM] = fmt.Sprintf("%.2f", netEcpm)
				bidExt.Prebid.Targeting[models.PWT_PLATFORM] = getPlatformName(rctx.Platform)
				if len(bid.DealID) != 0 {
					bidExt.Prebid.Targeting[models.PWT_DEALID] = bid.DealID
				}
			}

			var err error
			bidResponse.SeatBid[i].Bid[j].Ext, err = json.Marshal(bidExt)
			if err != nil {
				return bidResponse, err
			}
		}
	}

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

	return bidResponse, nil
}

// isNewWinningBid calculates if the new bid (nbid) will win against the current winning bid (wbid) given preferDeals.
func isNewWinningBid(bid, wbid owBid, preferDeals bool) bool {
	if preferDeals {
		//only wbid has deal
		if wbid.bidDealTierSatisfied && !bid.bidDealTierSatisfied {
			return false
		}
		//only bid has deal
		if !wbid.bidDealTierSatisfied && bid.bidDealTierSatisfied {
			return true
		}
	}
	//both have deal or both do not have deal
	return bid.netEcpm > wbid.netEcpm
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
