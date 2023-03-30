package openwrap

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// Tracker tracker url creation parameters
type Tracker struct {
	PubID             int
	PageURL           string
	Timestamp         int64
	IID               string
	ProfileID         string
	VersionID         string
	SlotID            string
	Adunit            string
	PartnerInfo       *Partner
	RewardedInventory int
	SURL              string // contains either req.site.domain or req.app.bundle value
	Platform          int
	Advertiser        string
	// SSAI identifies the name of the SSAI vendor
	// Applicable only in case of incase of video/json endpoint.
	SSAI string
}

// Partner partner information to be logged in tracker object
type Partner struct {
	PartnerID  string
	BidderCode string
	KGPV       string
	GrossECPM  float64
	NetECPM    float64
	BidID      string
	OrigBidID  string
}

func (m *OpenWrap) injectTrackers(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	tracker := Tracker{
		PubID:     rctx.PubID,
		ProfileID: fmt.Sprintf("%d", rctx.ProfileID),
		VersionID: fmt.Sprintf("%d", rctx.DisplayID),
		PageURL:   rctx.PageURL,
		Timestamp: rctx.StartTime,
		IID:       rctx.LoggerImpressionID,
		Platform:  int(rctx.DevicePlatform),
		SSAI:      rctx.SSAI,
	}

	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			tagid := ""
			secure := 0
			netECPM := 0
			matchedSlot := ""
			price := bid.Price
			isRewardInventory := 0

			if impCtx, ok := rctx.ImpBidCtx[bid.ImpID]; ok {
				if bidCtx, ok := impCtx.BidCtx[bid.ID]; ok {
					if bidResponse.Cur != "USD" {
						price = bidCtx.OriginalBidCPMUSD
					}
					netECPM = int(bidCtx.NetECPM)
				}

				if bidderMeta, ok := impCtx.Bidders[seatBid.Seat]; ok {
					matchedSlot = bidderMeta.MatchedSlot
				}

				tagid = impCtx.TagID
				secure = impCtx.Secure
				isRewardInventory = getRewardedInventoryFlag(rctx.ImpBidCtx[bid.ImpID].IsRewardInventory)
			}

			tracker.Adunit = tagid
			tracker.SlotID = fmt.Sprintf("%s_%s", bid.ImpID, tagid)
			tracker.RewardedInventory = isRewardInventory
			tracker.PartnerInfo = &Partner{
				PartnerID:  seatBid.Seat,
				BidderCode: seatBid.Seat,
				BidID:      bid.ID,
				OrigBidID:  bid.ID,
				KGPV:       matchedSlot,
				NetECPM:    float64(netECPM),
				GrossECPM:  models.GetGrossEcpm(price),
			}

			if len(bid.ADomain) != 0 {
				if domain, err := models.ExtractDomain(bid.ADomain[0]); err == nil {
					tracker.Advertiser = domain
				}
			}

			// construct tracker URL
			trackerURL := ConstructTrackerURL(tracker, m.cfg.OpenWrap.Tracker.Endpoint, secure, rctx.Platform)
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
func ConstructTrackerURL(tracker Tracker, trackerURLString string, secure int, platform string) string {
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
	if platform == models.PLATFORM_DISPLAY {
		if secure == 1 {
			trackerURL.Scheme = "https"
		} else {
			trackerURL.Scheme = "http"
		}

	}
	trackerQueryStr := trackerURL.String() + models.TRKQMARK + queryString
	return trackerQueryStr
}
