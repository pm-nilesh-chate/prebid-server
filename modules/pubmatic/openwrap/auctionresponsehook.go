package openwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) handleAuctionResponseHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	result := hookstage.HookResult[hookstage.AuctionResponsePayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.AuctionResponsePayload]{}
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

			revShare := GetRevenueShare(partnerNameMap[seatBid.Seat])
			netEcpm := GetNetEcpm(bid.Price, revShare)

			bidExt := &models.BidExt{}
			if len(bid.Ext) != 0 {
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
					bidExt.CreativeType = GetAdFormat(bid.AdM)
				}

				// bidExt.Summary

				// if platform == models.PLATFORM_APP {
				bidExt.NetECPM = netEcpm
				// bidExt.Prebid = addPWTTargetingForBid(*request.Id, eachBid, impExt.Prebid, *eachSeatBid.Seat, platform, winBidFlag, netEcpm)
				// }

				if rctx.ClientConfigFlag == 1 {
					if rctx.ImpBidCtx[bid.ImpID].Type == "banner" {
						bidExt.Banner.ClientConfig = GetClientConfigForMediaType(rctx, bid.ImpID, rctx.AdUnitConfig, "banner")
					} else if rctx.ImpBidCtx[bid.ImpID].Type == "video" {
						bidExt.Video.ClientConfig = GetClientConfigForMediaType(rctx, bid.ImpID, rctx.AdUnitConfig, "video")
					}
				}
			}

			owbid := owBid{&bid, netEcpm, bidExt.Prebid.DealTierSatisfied}
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

			revShare := GetRevenueShare(partnerNameMap[seatBid.Seat])
			netEcpm := GetNetEcpm(bid.Price, revShare)

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
			bidExt.Prebid.Targeting[CreatePartnerKey(seatBid.Seat, models.PWT_SLOTID)] = bid.ID
			bidExt.Prebid.Targeting[CreatePartnerKey(seatBid.Seat, models.PWT_SZ)] = GetSize(bid.W, bid.H)
			bidExt.Prebid.Targeting[CreatePartnerKey(seatBid.Seat, models.PWT_PARTNERID)] = seatBid.Seat
			bidExt.Prebid.Targeting[CreatePartnerKey(seatBid.Seat, models.PWT_ECPM)] = fmt.Sprintf("%.2f", netEcpm)
			bidExt.Prebid.Targeting[CreatePartnerKey(seatBid.Seat, models.PWT_PLATFORM)] = getPlatformName(rctx.Platform)
			bidExt.Prebid.Targeting[CreatePartnerKey(seatBid.Seat, models.PWT_BIDSTATUS)] = "1"
			if len(bid.DealID) != 0 {
				bidExt.Prebid.Targeting[CreatePartnerKey(seatBid.Seat, models.PWT_DEALID)] = bid.DealID
			}

			if _, ok := winningBids[bid.ImpID]; ok {
				// bidExt.Winner = ptrutil.ToPtr(1)
				bidExt.Winner = 1

				bidExt.Prebid.Targeting[models.PWT_SLOTID] = bid.ID
				bidExt.Prebid.Targeting[models.PWT_BIDSTATUS] = "1"
				bidExt.Prebid.Targeting[models.PWT_SZ] = GetSize(bid.W, bid.H)
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

// CreatePartnerKey returns key with partner appended
func CreatePartnerKey(partner, key string) string {
	if partner == "" {
		return key
	}
	return key + "_" + partner
}

func GetSize(width, height int64) string {
	return fmt.Sprintf("%dx%d", width, height)
}

// GetAdFormat gets adformat from creative(adm) of the bid
func GetAdFormat(adm string) string {
	adFormat := models.Banner
	videoRegex, _ := regexp.Compile("<VAST\\s+")

	if videoRegex.MatchString(adm) {
		adFormat = models.Video
	} else {
		var admJSON map[string]interface{}
		err := json.Unmarshal([]byte(strings.Replace(adm, "/\\/g", "", -1)), &admJSON)
		if err == nil && admJSON != nil && admJSON["native"] != nil {
			adFormat = models.Native
		}
	}
	return adFormat
}

func GetRevenueShare(partnerConfig map[string]string) float64 {
	var revShare float64

	if val, ok := partnerConfig[models.REVSHARE]; ok {
		revShare, _ = strconv.ParseFloat(val, 64)
	}
	return revShare
}

func GetNetEcpm(price float64, revShare float64) float64 {
	if revShare == 0 {
		return toFixed(price, models.BID_PRECISION)
	}
	price = price * (1 - revShare/100)
	return toFixed(price, models.BID_PRECISION)
}

func GetGrossEcpm(price float64) float64 {
	return toFixed(price, models.BID_PRECISION)
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
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
