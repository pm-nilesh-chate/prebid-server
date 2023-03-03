package models

import (
	"encoding/json"

	"github.com/prebid/prebid-server/openrtb_ext"
)

type BidExt struct {
	ErrorCode       int    `json:"errorCode,omitempty"`
	ErrorMsg        string `json:"errorMessage,omitempty"`
	Partner         string `json:"partner,omitempty"`
	RefreshInterval int    `json:"refreshInterval,omitempty"`
	CreativeType    string `json:"crtype,omitempty"`
	// AdPod           ExtBidPrebidAdPod `json:"adpod,omitempty"`
	Summary     []Summary       `json:"summary,omitempty"`
	Prebid      ExtBidPrebid    `json:"prebid,omitempty"`
	SKAdnetwork json.RawMessage `json:"skadn,omitempty"`
	Video       ExtBidVideo     `json:"video,omitempty"`
	Banner      ExtBidBanner    `json:"banner,omitempty"`
	DspId       int             `json:"dspid,omitempty"`
	Winner      int             `json:"winner,omitempty"`
	NetECPM     float64         `json:"netecpm,omitempty"`

	OriginalBidCPM    float64 `json:"origbidcpm,omitempty"`
	OriginalBidCur    string  `json:"origbidcur,omitempty"`
	OriginalBidCPMUSD float64 `json:"origbidcpmusd,omitempty"`
}

// ExtBidPrebid defines the contract for bidresponse.seatbid.bid[i].ext.prebid
type ExtBidPrebid struct {
	Cache             ExtBidPrebidCache            `json:"cache,omitempty"`
	Targeting         map[string]string            `json:"targeting"`
	Type              BidType                      `json:"type,omitempty"`
	DealTierSatisfied bool                         `json:"dealtiersatisfied,omitempty"`
	DealPriority      int                          `json:"dealpriority,omitempty"`
	Video             ExtBidPrebidVideo            `json:"video,omitempty"`
	BidID             string                       `json:"bidid,omitempty"`
	Meta              openrtb_ext.ExtBidPrebidMeta `json:"meta,omitempty"`
}

// ExtBidPrebidVideo defines the contract for bidresponse.seatbid.bid[i].ext.prebid.video
type ExtBidPrebidVideo struct {
	Duration  int    `json:"duration,omitempty"`
	VASTTagID string `json:"vasttagid,omitempty"`
	//PrimaryCategory string `json:"primary_category"`
}

// ExtBidVideo defines the contract for bidresponse.seatbid.bid[i].ext.video
type ExtBidVideo struct {
	MinDuration    int         `json:"minduration,omitempty"`    // Minimum video ad duration in seconds.
	MaxDuration    int         `json:"maxduration,omitempty"`    // Maximum video ad duration in seconds.
	Skip           int         `json:"skip,omitempty"`           // Indicates if the player will allow the video to be skipped,where 0 = no, 1 = yes.
	SkipMin        int         `json:"skipmin,omitempty"`        // Videos of total duration greater than this number of seconds can be skippable; only applicable if the ad is skippable.
	SkipAfter      int         `json:"skipafter,omitempty"`      // Number of seconds a video must play before skipping is enabled; only applicable if the ad is skippable.
	BAttr          []int       `json:"battr,omitempty"`          // Blocked creative attributes
	PlaybackMethod []int       `json:"playbackmethod,omitempty"` // Allowed playback methods
	ClientConfig   interface{} `json:"clientconfig,omitempty"`
}

// ExtBidBanner defines the contract for bidresponse.seatbid.bid[i].ext.banner
type ExtBidBanner struct {
	ClientConfig interface{} `json:"clientconfig,omitempty"`
}

// ExtBidPrebidCache defines the contract for  bidresponse.seatbid.bid[i].ext.prebid.cache
type ExtBidPrebidCache struct {
	Key string `json:"key"`
	Url string `json:"url"`
}

// Prebid Response Ext with DspId
type OWExt struct {
	openrtb_ext.ExtOWBid
	DspId int `json:"dspid,omitempty"`

	OriginalBidCPM    float64 `json:"origbidcpm,omitempty"`
	OriginalBidCur    string  `json:"origbidcur,omitempty"`
	OriginalBidCPMUSD float64 `json:"origbidcpmusd,omitempty"`
}

// ExtBidderMessage defines an error object to be returned, consiting of a machine readable error code, and a human readable error message string.
type ExtBidderMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BidType describes the allowed values for bidresponse.seatbid.bid[i].ext.prebid.type
type BidType string

// Targeting map of string of strings
type Targeting map[string]string

type Summary struct {
	VastTagID    string  `json:"vastTagID,omitempty"`
	Bidder       string  `json:"bidder,omitempty"`
	Bid          float64 `json:"bid,omitempty"`
	ErrorCode    int     `json:"errorCode,omitempty"`
	ErrorMessage string  `json:"errorMessage,omitempty"`
	Width        int     `json:"width,omitempty"`
	Height       int     `json:"height,omitempty"`
	Regex        string  `json:"regex,omitempty"`
}
