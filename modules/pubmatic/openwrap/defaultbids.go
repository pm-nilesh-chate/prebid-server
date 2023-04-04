package openwrap

import (
	"encoding/json"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m *OpenWrap) addDefaultBids(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) map[string]map[string][]openrtb2.Bid {
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

	// consider responded but dropped bids to avoid false nobid entries
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

	// add nobids for throttled adapter to all the impressions (how do we set profile with custom list of bidders at impression level?)
	for bidder := range rctx.AdapterThrottleMap {
		for impID := range rctx.ImpBidCtx { // ImpBidCtx is used only for list of impID, it does not have data of throttled adapters
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

	return noSeatBids
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

func (m *OpenWrap) applyDefaultBids(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	// update nobids in final response
	for i, seatBid := range bidResponse.SeatBid {
		for impID, noSeatBid := range rctx.NoSeatBids {
			for seat, bids := range noSeatBid {
				if seatBid.Seat == seat {
					bidResponse.SeatBid[i].Bid = append(bidResponse.SeatBid[i].Bid, bids...)
					delete(noSeatBid, seat)
					rctx.NoSeatBids[impID] = noSeatBid
				}
			}
		}
	}

	// no-seat case
	for _, noSeatBid := range rctx.NoSeatBids {
		for seat, bids := range noSeatBid {
			bidResponse.SeatBid = append(bidResponse.SeatBid, openrtb2.SeatBid{
				Bid:  bids,
				Seat: seat,
			})
		}
	}

	return bidResponse, nil
}
