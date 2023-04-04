package openwrap

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m *OpenWrap) createTrackers(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) map[string]models.Tracker {
	trackers := make(map[string]models.Tracker)

	// pubmatic's KGP details per impression
	type pubmaticMarketplaceMeta struct {
		PubmaticKGP, PubmaticKGPV, PubmaticKGPSV string
	}
	pmMkt := make(map[string]pubmaticMarketplaceMeta)

	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			tracker := models.Tracker{
				PubID:     rctx.PubID,
				ProfileID: fmt.Sprintf("%d", rctx.ProfileID),
				VersionID: fmt.Sprintf("%d", rctx.DisplayID),
				PageURL:   rctx.PageURL,
				Timestamp: rctx.StartTime,
				IID:       rctx.LoggerImpressionID,
				Platform:  int(rctx.DevicePlatform),
				SSAI:      rctx.SSAI,
				ImpID:     bid.ImpID,
			}

			tagid := ""
			netECPM := float64(0)
			matchedSlot := ""
			price := bid.Price
			isRewardInventory := 0
			partnerID := seatBid.Seat

			var isRegex bool
			var kgp, kgpv, kgpsv string

			if impCtx, ok := rctx.ImpBidCtx[bid.ImpID]; ok {
				if bidderMeta, ok := impCtx.Bidders[seatBid.Seat]; ok {
					matchedSlot = bidderMeta.MatchedSlot
					partnerID = bidderMeta.PrebidBidderCode
				}

				if bidCtx, ok := impCtx.BidCtx[bid.ID]; ok {
					if bidResponse.Cur != "USD" {
						price = bidCtx.OriginalBidCPMUSD
					}
					netECPM = bidCtx.NetECPM

					// TODO do most calculation in wt
					// marketplace/alternatebiddercodes feature
					bidExt := bidCtx.BidExt
					if bidExt.Prebid != nil && bidExt.Prebid.Meta != nil && len(bidExt.Prebid.Meta.AdapterCode) != 0 && seatBid.Seat != bidExt.Prebid.Meta.AdapterCode {
						partnerID = bidExt.Prebid.Meta.AdapterCode

						if aliasSeat, ok := rctx.PrebidBidderCode[partnerID]; ok {
							if bidderMeta, ok := impCtx.Bidders[aliasSeat]; ok {
								matchedSlot = bidderMeta.MatchedSlot
							}
						}
					}
				}

				_ = matchedSlot
				// --------------------------------------------------------------------------------------------------
				// Move this code to a function. Confirm the kgp, kgpv, kgpsv relation in wt and wl.
				// --------------------------------------------------------------------------------------------------
				// var kgp, kgpv, kgpsv string

				if bidderMeta, ok := impCtx.Bidders[seatBid.Seat]; ok {
					partnerID = bidderMeta.PrebidBidderCode
					kgp = bidderMeta.KGP
					kgpv = bidderMeta.KGPV
					kgpsv = bidderMeta.MatchedSlot
					isRegex = bidderMeta.IsRegex
				}

				// 1. nobid
				if bid.Price == 0 && bid.H == 0 && bid.W == 0 {
					//NOTE: kgpsv = bidderMeta.MatchedSlot above. Use the same
					kgpv = kgpsv
				} else if !isRegex {
					// 2. valid bid
					// kgpv has regex, do not generate slotName again
					// kgpsv could be unmapped or mapped slot, generate slotName again based on bid.H and bid.W
					kgpsv := bidderparams.GenerateSlotName(bid.H, bid.W, kgp, impCtx.TagID, impCtx.Div, rctx.Source)
					kgpv = kgpsv
				}
				// --------------------------------------------------------------------------------------------------

				tagid = impCtx.TagID
				tracker.Secure = impCtx.Secure
				isRewardInventory = getRewardedInventoryFlag(rctx.ImpBidCtx[bid.ImpID].IsRewardInventory)
			}

			if seatBid.Seat == "pubmatic" {
				pmMkt[bid.ImpID] = pubmaticMarketplaceMeta{
					PubmaticKGP:   kgp,
					PubmaticKGPV:  kgpv,
					PubmaticKGPSV: kgpsv,
				}
			}

			tracker.Adunit = tagid
			tracker.SlotID = fmt.Sprintf("%s_%s", bid.ImpID, tagid)
			tracker.RewardedInventory = isRewardInventory
			tracker.PartnerInfo = &models.Partner{
				PartnerID:  partnerID,
				BidderCode: seatBid.Seat,
				BidID:      bid.ID,
				OrigBidID:  bid.ID,
				KGPV:       kgpv,
				NetECPM:    float64(netECPM),
				GrossECPM:  models.GetGrossEcpm(price),
			}

			if len(bid.ADomain) != 0 {
				if domain, err := models.ExtractDomain(bid.ADomain[0]); err == nil {
					tracker.Advertiser = domain
				}
			}

			trackers[bid.ID] = tracker
		}
	}

	for bidID, tracker := range trackers {
		if tracker.PartnerInfo != nil {
			if _, ok := rctx.MarketPlaceBidders[tracker.PartnerInfo.BidderCode]; ok {
				if v, ok := pmMkt[tracker.ImpID]; ok {
					tracker.PartnerInfo.KGPV = v.PubmaticKGPV
				}
			}
		}
		trackers[bidID] = tracker
	}

	return trackers
}

func (m *OpenWrap) injectTrackers(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			trackerURL := ConstructTrackerURL(rctx, seatBid.Seat, bid.ID, m.cfg.OpenWrap.Tracker.Endpoint)
			trackURL, err := url.Parse(trackerURL)
			if err == nil {
				trackURL.Scheme = models.HTTPSProtocol
				bidResponse.SeatBid[i].Bid[j].AdM += strings.Replace(models.TrackerCallWrap, "${escapedUrl}", trackURL.String(), 1)
			}
		}
	}
	return bidResponse, nil

}

func getRewardedInventoryFlag(reward *int8) int {
	if reward != nil {
		return int(*reward)
	}
	return 0
}

// ConstructTrackerURL constructing tracker url for impression
func ConstructTrackerURL(rctx models.RequestCtx, seat, bidID string, trackerURLString string) string {
	tracker := rctx.Trackers[bidID]

	trackerURL, err := url.Parse(trackerURLString)
	if err != nil {
		return ""
	}

	v := url.Values{}
	v.Set(models.TRKPubID, strconv.Itoa(tracker.PubID))
	v.Set(models.TRKPageURL, tracker.PageURL)
	v.Set(models.TRKTimestamp, strconv.FormatInt(tracker.Timestamp, 10))
	v.Set(models.TRKIID, tracker.IID)
	v.Set(models.TRKProfileID, tracker.ProfileID)
	v.Set(models.TRKVersionID, tracker.VersionID)
	v.Set(models.TRKSlotID, tracker.SlotID)
	v.Set(models.TRKAdunit, tracker.Adunit)
	if tracker.RewardedInventory == 1 {
		v.Set(models.TRKRewardedInventory, strconv.Itoa(tracker.RewardedInventory))
	}
	partner := tracker.PartnerInfo
	v.Set(models.TRKPartnerID, partner.PartnerID)
	v.Set(models.TRKBidderCode, partner.BidderCode)
	v.Set(models.TRKKGPV, partner.KGPV)
	v.Set(models.TRKGrossECPM, fmt.Sprint(partner.GrossECPM))
	v.Set(models.TRKNetECPM, fmt.Sprint(partner.NetECPM))
	v.Set(models.TRKBidID, partner.BidID)
	if tracker.SSAI != "" {
		v.Set(models.TRKSSAI, tracker.SSAI)
	}
	v.Set(models.TRKOrigBidID, partner.OrigBidID)
	queryString := v.Encode()

	//Code for making tracker call http/https based on secure flag for in-app platform
	//TODO change platform to models.PLATFORM_APP once in-app platform starts populating from wrapper UI
	if rctx.Platform == models.PLATFORM_DISPLAY {
		if tracker.Secure == 1 {
			trackerURL.Scheme = "https"
		} else {
			trackerURL.Scheme = "http"
		}

	}
	trackerQueryStr := trackerURL.String() + models.TRKQMARK + queryString
	return trackerQueryStr
}
