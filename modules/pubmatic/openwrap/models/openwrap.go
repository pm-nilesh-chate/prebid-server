package models

import (
	"encoding/json"
	"net/http"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

type RequestCtx struct {
	PubID, ProfileID, DisplayID, VersionID int
	SSAuction                              int
	SummaryDisable                         int
	LogInfoFlag                            int
	SSAI                                   string
	PartnerConfigMap                       map[int]map[string]string
	PreferDeals                            bool
	Platform                               string
	LoggerImpressionID                     string
	ClientConfigFlag                       int

	IP string

	//NYC_TODO: use enum?
	IsTestRequest bool
	IsCTVRequest  bool

	UA        string
	Cookies   string
	UidCookie *http.Cookie

	Debug bool
	Trace bool

	//tracker
	PageURL        string
	StartTime      int64
	DevicePlatform DevicePlatform

	//logger
	URL string

	// imp-bid ctx to avoid computing same thing for bidder params, logger and tracker
	ImpBidCtx   map[string]ImpCtx
	Aliases     map[string]string
	NewReqExt   json.RawMessage
	ResponseExt json.RawMessage

	AdapterThrottleMap map[string]struct{}

	AdUnitConfig *adunitconfig.AdUnitConfig

	Source string

	SendAllBids bool
	WinningBids map[string]OwBid
	DroppedBids map[string][]openrtb2.Bid
}

type OwBid struct {
	*openrtb2.Bid
	NetEcpm              float64
	BidDealTierSatisfied bool
}

func (r RequestCtx) GetVersionLevelKey(key string) (string, bool) {
	if len(r.PartnerConfigMap) == 0 || len(r.PartnerConfigMap[VersionLevelConfigID]) == 0 {
		return "", false
	}
	v, ok := r.PartnerConfigMap[VersionLevelConfigID][key]
	return v, ok
}

type ImpCtx struct {
	ImpID             string
	TagID             string
	Secure            int
	KGPV              string
	IsRewardInventory *int8
	Video             *openrtb2.Video
	Type              string // banner, video, native, etc
	Bidders           map[string]PartnerData
	NewExt            json.RawMessage
	BidCtx            map[string]BidCtx

	BannerAdUnitCtx AdUnitCtx
	VideoAdUnitCtx  AdUnitCtx
}

type PartnerData struct {
	PartnerID   int
	MatchedSlot string
	KGPV        string
	Params      json.RawMessage
}

type BidCtx struct {
	BidExt
}

type AdUnitCtx struct {
	MatchedSlot              string
	IsRegex                  bool
	MatchedRegex             string
	SelectedSlotAdUnitConfig *adunitconfig.AdConfig
	AppliedSlotAdUnitConfig  *adunitconfig.AdConfig
	UsingDefaultConfig       bool
	AllowedConnectionTypes   []int
}
