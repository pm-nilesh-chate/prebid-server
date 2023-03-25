package openwrap

import (
	"encoding/json"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type RequestType string

const (
	COOKIE_SYNC        RequestType = "/cookie_sync"
	AUCTION            RequestType = "/openrtb2/auction"
	VIDEO              RequestType = "/openrtb2/video"
	SETUID             RequestType = "/set_uid"
	AMP                RequestType = "/openrtb2/amp"
	NOTIFICATION_EVENT RequestType = "/event"
)

// Module that can perform transactional logging
type HTTPLogger struct {
	URL    string
	client *http.Client
}

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
			IID:          rCtx.LoggerImpressionID,
			Timestamp:    rCtx.StartTime,
			ServerLogger: 1,
		},
	}

	extWrapper := models.RequestExtWrapper{}
	err := json.Unmarshal(ao.Request.Ext, &extWrapper)
	if err != nil {
		return ""
	}

	var publisherID int
	var pageURL, origin string
	if ao.Request.App != nil {
		if ao.Request.App.Publisher != nil {
			publisherID, _ = strconv.Atoi(ao.Request.App.Publisher.ID)
		}
		pageURL = ao.Request.App.StoreURL
		origin = ao.Request.App.Bundle
	} else if ao.Request.Site != nil {
		if ao.Request.Site.Publisher != nil {
			publisherID, _ = strconv.Atoi(ao.Request.Site.Publisher.ID)
		}
		pageURL = ao.Request.Site.Page
		if len(ao.Request.Site.Domain) != 0 {
			origin = ao.Request.Site.Domain
		} else {
			pageURL, err := url.Parse(ao.Request.Site.Page)
			if err == nil && pageURL != nil {
				origin = pageURL.Host
			}
		}
	}
	wlog.SetPubID(publisherID)
	wlog.SetOrigin(origin)
	wlog.SetPageURL(pageURL)

	var consent string
	if ao.Request.User != nil {
		extUser := openrtb_ext.UserExt{}
		err := json.Unmarshal(ao.Request.User.Ext, &extWrapper)
		if err != nil {

		}
		if c := extUser.GetConsent(); c != nil {
			consent = *c
		}
	}
	wlog.SetConsentString(consent)

	var gdpr int8
	if ao.Request.Regs != nil {
		extReg := openrtb_ext.RegExt{}
		err := json.Unmarshal(ao.Request.Regs.Ext, &extWrapper)
		if err != nil {

		}
		gdpr = *extReg.GetGDPR()
	}
	wlog.SetGDPR(int(gdpr))

	if ao.Request.Device != nil {
		wlog.SetIP(ao.Request.Device.IP)
		wlog.SetUserAgent(ao.Request.Device.UA)
	}

	//log device object
	wlog.logDeviceObject(*rCtx, rCtx.UA, ao.Request, rCtx.Platform)

	//log content object
	if nil != ao.Request.Site {
		wlog.logContentObject(ao.Request.Site.Content)
	} else if nil != ao.Request.App {
		wlog.logContentObject(ao.Request.App.Content)
	}

	// //log adpod percentage object
	// if nil != ao.Request.Ext {
	// 	ext, ok := ao.Request.Ext.(*openrtb.ExtRequest)
	// 	if ok {
	// 		wlog.logAdPodPercentage(ext.AdPod)
	// 	}
	// }

	wlog.SetTimeout(int(ao.Request.TMax))

	// if uidCookie, _ := rCtx.Cookies(models.KADUSERCOOKIE); uidCookie != nil {
	// 	controller.LoggerRecord.SetUID(uidCookie.Value)
	// }
	// if testConfigApplied {
	// 	controller.LoggerRecord.SetTestConfigApplied(1)
	// }

	// wl := CreateCommonLogger(ao)

	// partnerMap := make([]map[string]PartnerRecord, len(ao.Request.Imp))
	// // for imp, impCtx := range rCtx.ImpBidCtx {
	// for i, imp := range ao.Request.Imp {
	// 	impCtx, ok := rCtx.ImpBidCtx[imp.ID]
	// 	if !ok {
	// 		continue
	// 	}

	// 	partnerMap[i] = make(map[string]PartnerRecord)
	// 	for bidder, bidderData := range impCtx.Bidders { // for partnerID, partnerConfig := range partnerConfigMap {
	// 		partnerMap[i][bidder] = PartnerRecord{
	// 			PartnerID:        bidder,
	// 			BidderCode:       bidder,
	// 			PartnerSize:      "0x0",
	// 			KGPV:             impCtx.MatchedSlot,
	// 			KGPSV:            impCtx.KGPV,
	// 			BidID:            imp.ID,
	// 			OrigBidID:        imp.ID,
	// 			DefaultBidStatus: 1,
	// 			ServerSide:       1,
	// 			RevShare: func() float64 {
	// 				r, _ := strconv.ParseFloat(rCtx.PartnerConfigMap[bidderData.PartnerID][models.REVSHARE], 64)
	// 				return r
	// 			}(),
	// 			KGP: impCtx.MatchedSlot,
	// 			// MatchedImpression: matchedImpression,
	// 		}
	// 	}
	// }

	// imp-partner-record
	// ipr := make(map[string]map[string]PartnerRecord)
	ipr := make(map[string][]PartnerRecord)
	for _, seatBid := range ao.Response.SeatBid {
		if seatBid.Seat == string(openrtb_ext.BidderOWPrebidCTV) {
			continue
		}

		for _, bid := range seatBid.Bid {
			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			// if _, ok := ipr[bid.ImpID]; !ok {
			// 	ipr[bid.ImpID] = make(map[string]PartnerRecord)
			// }

			revShare := 0.0
			if pd, ok := impCtx.Bidders[seatBid.Seat]; ok {
				revShare, _ = strconv.ParseFloat(rCtx.PartnerConfigMap[pd.PartnerID][models.REVSHARE], 64)
			}

			bidExt := models.BidExt{}
			_ = json.Unmarshal(bid.Ext, &bidExt)

			pr := PartnerRecord{
				PartnerID:        seatBid.Seat,
				BidderCode:       seatBid.Seat,
				KGPV:             impCtx.MatchedSlot,
				KGPSV:            impCtx.MatchedSlot,
				BidID:            bid.ID,
				OrigBidID:        bid.ID,
				DefaultBidStatus: 0,
				ServerSide:       1,
				// MatchedImpression: matchedImpression,
				NetECPM: func() float64 {
					if revShare != 0.0 {
						return GetNetEcpm(bid.Price, revShare)
					}
					return bid.Price
				}(),
				GrossECPM:       GetGrossEcpm(bid.Price),
				OriginalCPM:     GetGrossEcpm(bidExt.OriginalBidCPM),
				OriginalCur:     bidExt.OriginalBidCur,
				PartnerSize:     getSizeForPlatform(bid.W, bid.H, rCtx.Platform),
				DealID:          bid.DealID,
				Adformat:        GetAdFormat(bid.AdM),
				WinningBidStaus: bidExt.Winner,
			}

			if len(pr.OriginalCur) == 0 {
				pr.OriginalCPM = float64(0)
				pr.OriginalCur = "USD"
			}

			if len(pr.DealID) != 0 {
				pr.DealChannel = models.DEFAULT_DEALCHANNEL
			}

			if bidExt.Prebid.DealTierSatisfied && bidExt.Prebid.DealPriority > 0 {
				pr.DealPriority = bidExt.Prebid.DealPriority
			}

			if bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration > 0 {
				pr.AdDuration = &bidExt.Prebid.Video.Duration
			}
			//prepare Meta Object
			if bidExt.Prebid.Meta != nil {
				pr.setMetaDataObject(bidExt.Prebid.Meta)
			}

			if len(bid.ADomain) != 0 {
				if domain, err := ExtractDomain(bid.ADomain[0]); err == nil {
					pr.ADomain = domain
				}
			}

			// ipr[bid.ImpID][seatBid.Seat] = pr
			ipr[bid.ImpID] = append(ipr[bid.ImpID], pr)
		}
	}

	slots := make([]SlotRecord, 0)
	for _, imp := range ao.Request.Imp {
		reward := 0
		if v, ok := rCtx.ImpBidCtx[imp.ID]; ok && v.IsRewardInventory != nil {
			reward = int(*v.IsRewardInventory)
		}

		slots = append(slots, SlotRecord{
			SlotName:          getSlotName(imp.ID, imp.TagID),
			SlotSize:          getSizesFromImp(imp, rCtx.Platform),
			Adunit:            imp.TagID,
			PartnerData:       ipr[imp.ID],
			RewardedInventory: int(reward),
			// AdPodSlot:         getAdPodSlot(imp, responseMap.AdPodBidsExt),
		})
	}

	wlog.SetProfileID(strconv.Itoa(rCtx.ProfileID))
	wlog.SetVersionID(strconv.Itoa(rCtx.DisplayID))

	wlog.SetSlots(slots)

	return PrepareLoggerURL(&wlog, rCtx.URL, GetGdprEnabledFlag(rCtx.PartnerConfigMap))
}

// Writes AuctionObject to file
func (ow *HTTPLogger) LogAuctionObject(ao *analytics.AuctionObject) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	_ = GetLogAuctionObjectAsURL(ao)
	// Send(*ow.client, ow.URL, wl, 1) // NYC_TODO: pass gdpr enabled in ao.Context
}

// Writes VideoObject to file
func (ow *HTTPLogger) LogVideoObject(vo *analytics.VideoObject) {
}

// Logs SetUIDObject to file
func (ow *HTTPLogger) LogSetUIDObject(so *analytics.SetUIDObject) {
}

// Logs CookieSyncObject to file
func (ow *HTTPLogger) LogCookieSyncObject(cso *analytics.CookieSyncObject) {
}

// Logs AmpObject to file
func (ow *HTTPLogger) LogAmpObject(ao *analytics.AmpObject) {
}

// Logs NotificationEvent to file
func (ow *HTTPLogger) LogNotificationEventObject(ne *analytics.NotificationEvent) {
}

// Method to initialize the analytic module
func NewHTTPLogger(url string) (analytics.PBSAnalyticsModule, error) {
	return &HTTPLogger{
		URL: url,
	}, nil
}

// GetGdprEnabledFlag returns gdpr flag set in the partner config
func GetGdprEnabledFlag(partnerConfigMap map[int]map[string]string) int {
	gdpr := 0
	if val := partnerConfigMap[models.VersionLevelConfigID][models.GDPR_ENABLED]; val != "" {
		gdpr, _ = strconv.Atoi(val)
	}
	return gdpr
}