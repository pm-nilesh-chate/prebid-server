package openwrap

import (
	"context"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (m OpenWrap) handleAuctionResponseHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	result := hookstage.HookResult[hookstage.AuctionResponsePayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.AuctionResponsePayload]{}
	result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
		rctx := result.ModuleContext["rctx"].(models.RequestCtx)
		err := m.updateORTBV25Response(rctx, ap.BidResponse)
		return ap, err
	}, hookstage.MutationUpdate, "response-body-with-sshb-format")

	return result, nil
}

type bidExtOW struct {
	openrtb_ext.ExtBid
}

func (m *OpenWrap) updateORTBV25Response(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) error {
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if len(bid.Ext) != 0 {
				bidExt := &models.BidExt{}
				err := json.Unmarshal(bid.Ext, bidExt)
				if err != nil {
					return err
				}

				if v, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID]["refreshInterval"]; ok {
					n, err := strconv.Atoi(v)
					if err == nil {
						bidExt.RefreshInterval = n
					}
				}

				bidExt.CreativeType = GetAdFormat(bid.AdM)

				// bidExt.Summary

				// NYC_TODO maintain a global map of ow-partner-id to biddercode. Ex. 8->pubmatic
				// prepare partner name to partner config map
				partnerNameMap := make(map[string]map[string]string)
				for _, partnerConfig := range rctx.PartnerConfigMap {
					if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
						continue
					}

					partnerNameMap[partnerConfig[models.BidderCode]] = partnerConfig
				}

				// if platform == models.PLATFORM_APP {
				revShare := GetRevenueShare(partnerNameMap[seatBid.Seat])
				netEcpm := GetNetEcpm(bid.Price, revShare)

				bidExt.NetECPM = netEcpm
				// bidExt.Prebid = addPWTTargetingForBid(*request.Id, eachBid, impExt.Prebid, *eachSeatBid.Seat, platform, winBidFlag, netEcpm)
				// }
			}
		}
	}
	return nil
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
