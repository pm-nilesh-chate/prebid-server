package openwrap

import (
	"encoding/json"
	"fmt"
	"net/url"
	"runtime/debug"
	"strconv"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func GetLogAuctionObjectAsURL(ao *analytics.AuctionObject) string {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	// TODO filter by name
	// (*stageOutcomes[8].Groups[0].InvocationResults[0].AnalyticsTags.Activities[0].Results[0].Values["request-ctx"].(data))
	rCtx := func() *models.RequestCtx {
		for _, stageOutcome := range ao.HookExecutionOutcome {
			for _, groups := range stageOutcome.Groups {
				for _, invocationResult := range groups.InvocationResults {
					for _, activity := range invocationResult.AnalyticsTags.Activities {
						for _, result := range activity.Results {
							if result.Values != nil {
								if irctx, ok := result.Values["request-ctx"]; ok {
									rctx, ok := irctx.(*models.RequestCtx)
									if !ok {
										return nil
									}
									return rctx
								}
							}
						}
					}
				}
			}
		}
		return nil
	}()

	if rCtx == nil {
		return ""
	}

	wlog := WloggerRecord{
		record: record{
			PubID:             rCtx.PubID,
			ProfileID:         fmt.Sprintf("%d", rCtx.ProfileID),
			VersionID:         fmt.Sprintf("%d", rCtx.DisplayID),
			Origin:            rCtx.Source,
			PageURL:           rCtx.PageURL,
			IID:               rCtx.LoggerImpressionID,
			Timestamp:         rCtx.StartTime,
			ServerLogger:      1,
			TestConfigApplied: rCtx.ABTestConfigApplied,
			Timeout:           int(ao.Request.TMax),
		},
	}

	extWrapper := models.RequestExtWrapper{}
	err := json.Unmarshal(ao.Request.Ext, &extWrapper)
	if err != nil {
		return ""
	}

	if ao.Request.App != nil {
		wlog.Origin = ao.Request.App.Bundle
	} else if ao.Request.Site != nil {
		if len(ao.Request.Site.Domain) != 0 {
			wlog.Origin = ao.Request.Site.Domain
		} else {
			pageURL, err := url.Parse(ao.Request.Site.Page)
			if err == nil && pageURL != nil {
				wlog.Origin = pageURL.Host
			}
		}
	}

	if ao.Request.User != nil {
		extUser := openrtb_ext.ExtUser{}
		_ = json.Unmarshal(ao.Request.User.Ext, &extUser)
		wlog.ConsentString = extUser.Consent
	}

	if ao.Request.Device != nil {
		wlog.IP = ao.Request.Device.IP
		wlog.UserAgent = ao.Request.Device.UA
	}

	if ao.Request.Regs != nil {
		extReg := openrtb_ext.ExtRegs{}
		_ = json.Unmarshal(ao.Request.Regs.Ext, &extReg)
		if extReg.GDPR != nil {
			wlog.GDPR = *extReg.GDPR
		}
	}

	//log device object
	wlog.logDeviceObject(*rCtx, rCtx.UA, ao.Request, rCtx.Platform)

	//log content object
	if nil != ao.Request.Site {
		wlog.logContentObject(ao.Request.Site.Content)
	} else if nil != ao.Request.App {
		wlog.logContentObject(ao.Request.App.Content)
	}

	// impID-partnerRecords: partner records per impression
	ipr := make(map[string][]PartnerRecord)

	loggerSeat := make(map[string][]openrtb2.Bid)
	for _, seatBid := range ao.Response.SeatBid {
		loggerSeat[seatBid.Seat] = append(loggerSeat[seatBid.Seat], seatBid.Bid...)
	}
	for seat, Bids := range rCtx.DroppedBids {
		loggerSeat[seat] = append(loggerSeat[seat], Bids...)
	}

	// pubmatic's KGP details per impression
	type pubmaticMarketplaceMeta struct {
		PubmaticKGP, PubmaticKGPV, PubmaticKGPSV string
	}
	pmMkt := make(map[string]pubmaticMarketplaceMeta)

	for seat, bids := range loggerSeat {
		if seat == string(openrtb_ext.BidderOWPrebidCTV) {
			continue
		}

		if _, ok := rCtx.AdapterThrottleMap[seat]; ok {
			continue
		}

		for _, bid := range bids {
			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			if _, ok := impCtx.NonMapped[seat]; ok {
				break
			}

			revShare := 0.0
			partnerID := seat
			var isRegex bool
			var kgp, kgpv, kgpsv string

			if bidderMeta, ok := impCtx.Bidders[seat]; ok {
				revShare, _ = strconv.ParseFloat(rCtx.PartnerConfigMap[bidderMeta.PartnerID][models.REVSHARE], 64)
				partnerID = bidderMeta.PrebidBidderCode
				kgp = bidderMeta.KGP
				kgpv = bidderMeta.KGPV
				kgpsv = bidderMeta.MatchedSlot
				isRegex = bidderMeta.IsRegex
			}

			// 1. nobid
			if bid.Price == 0 && bid.H == 0 && bid.W == 0 {
				//NOTE: kgpsv = bidderMeta.MatchedSlot above. Use the same
				if !isRegex && kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv
				} else if !isRegex {
					kgpv = kgpsv
				}
			} else if !isRegex {
				if kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv
				} else if bid.H != 0 && bid.W != 0 { // Check when bid.H and bid.W will be zero with Price !=0. Ex: MobileInApp-MultiFormat-OnlyBannerMapping_Criteo_Partner_Validaton
					// 2. valid bid
					// kgpv has regex, do not generate slotName again
					// kgpsv could be unmapped or mapped slot, generate slotName again based on bid.H and bid.W
					kgpsv = GenerateSlotName(bid.H, bid.W, kgp, impCtx.TagID, impCtx.Div, rCtx.Source)
					kgpv = kgpsv
				}
			}

			if kgpv == "" {
				kgpv = kgpsv
			}

			bidExt := models.BidExt{}
			_ = json.Unmarshal(bid.Ext, &bidExt)

			price := bid.Price
			if ao.Response.Cur != "USD" {
				price = bidExt.OriginalBidCPMUSD
			}

			if seat == "pubmatic" {
				pmMkt[bid.ImpID] = pubmaticMarketplaceMeta{
					PubmaticKGP:   kgp,
					PubmaticKGPV:  kgpv,
					PubmaticKGPSV: kgpsv,
				}
			}

			pr := PartnerRecord{
				PartnerID:  partnerID, // prebid biddercode
				BidderCode: seat,      // pubmatic biddercode: pubmatic2
				// AdapterCode: adapterCode, // prebid adapter that brought the bid
				KGPV:             kgpv,
				KGPSV:            kgpsv,
				BidID:            bid.ID,
				OrigBidID:        bid.ID,
				DefaultBidStatus: 0,
				ServerSide:       1,
				// MatchedImpression: matchedImpression,
				NetECPM: func() float64 {
					if revShare != 0.0 {
						return GetNetEcpm(price, revShare)
					}
					return price
				}(),
				GrossECPM:   GetGrossEcpm(price),
				OriginalCPM: GetGrossEcpm(bidExt.OriginalBidCPM),
				OriginalCur: bidExt.OriginalBidCur,
				PartnerSize: getSizeForPlatform(bid.W, bid.H, rCtx.Platform),
				DealID:      bid.DealID,
			}

			// don't want default banner for nobid in wl
			if len(bid.AdM) != 0 {
				pr.Adformat = models.GetAdFormat(bid.AdM)
			}

			if b, ok := rCtx.WinningBids[bid.ImpID]; ok && b.ID == bid.ID {
				pr.WinningBidStaus = 1
			}

			if len(pr.OriginalCur) == 0 {
				pr.OriginalCPM = float64(0)
				pr.OriginalCur = "USD"
			}

			if len(pr.DealID) != 0 {
				pr.DealChannel = models.DEFAULT_DEALCHANNEL
			}

			if bidExt.Prebid != nil && bidExt.Prebid.DealTierSatisfied && bidExt.Prebid.DealPriority > 0 {
				pr.DealPriority = bidExt.Prebid.DealPriority
			}

			if bidExt.Prebid != nil && bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration > 0 {
				pr.AdDuration = &bidExt.Prebid.Video.Duration
			}
			//prepare Meta Object
			if bidExt.Prebid != nil && bidExt.Prebid.Meta != nil {
				pr.setMetaDataObject(bidExt.Prebid.Meta)
			}

			if len(bid.ADomain) != 0 {
				if domain, err := ExtractDomain(bid.ADomain[0]); err == nil {
					pr.ADomain = domain
				}
			}

			ipr[bid.ImpID] = append(ipr[bid.ImpID], pr)
		}
	}

	// overwrite marketplace bid details with that of partner adatper
	if rCtx.MarketPlaceBidders != nil {
		for impID, partnerRecords := range ipr {
			for i := 0; i < len(partnerRecords); i++ {
				if _, ok := rCtx.MarketPlaceBidders[partnerRecords[i].BidderCode]; ok {
					partnerRecords[i].PartnerID = "pubmatic"
					partnerRecords[i].KGPV = pmMkt[impID].PubmaticKGPV
					partnerRecords[i].KGPSV = pmMkt[impID].PubmaticKGPSV
				}
			}
			ipr[impID] = partnerRecords
		}
	}

	// parent bidder could in one of the above and we need them by prebid's bidderCode and not seat(could be alias)
	slots := make([]SlotRecord, 0)
	for _, imp := range ao.Request.Imp {
		reward := 0
		var incomingSlots []string
		if impCtx, ok := rCtx.ImpBidCtx[imp.ID]; ok {
			if impCtx.IsRewardInventory != nil {
				reward = int(*impCtx.IsRewardInventory)
			}
			incomingSlots = impCtx.IncomingSlots
		}

		slots = append(slots, SlotRecord{
			SlotName:          getSlotName(imp.ID, imp.TagID),
			SlotSize:          incomingSlots,
			Adunit:            imp.TagID,
			PartnerData:       ipr[imp.ID],
			RewardedInventory: int(reward),
			// AdPodSlot:         getAdPodSlot(imp, responseMap.AdPodBidsExt),
		})
	}

	wlog.Slots = slots

	return PrepareLoggerURL(&wlog, rCtx.URL, GetGdprEnabledFlag(rCtx.PartnerConfigMap))
}
