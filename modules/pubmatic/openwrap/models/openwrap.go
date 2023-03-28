package models

import (
	"encoding/json"

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

	//NYC_TODO: use enum?
	IsTestRequest bool
	IsCTVRequest  bool

	UA      string
	Cookies string

	Debug bool

	//tracker
	PageURL        string
	StartTime      int64
	DevicePlatform DevicePlatform

	//logger
	URL string

	// imp-bid ctx to avoid computing same thing for bidder params, logger and tracker
	ImpBidCtx map[string]ImpCtx
	Aliases   map[string]string

	AdapterThrottleMap map[string]struct{}

	AdUnitConfig            *adunitconfig.AdUnitConfig
	AdUnitConfigMatchedSlot string

	Source string
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
	MatchedSlot       string
	IsRewardInventory *int8
	Type              string // banner, video, native, etc
	Bidders           map[string]PartnerData
	BidCtx            map[string]BidCtx
}

type PartnerData struct {
	PartnerID int
	Params    json.RawMessage
}

type BidCtx struct {
	BidID      string
	OrigBidID  string
	PartnerID  string
	BidderCode string
	GrossECPM  float64
	NetECPM    float64
}
