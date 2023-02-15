package ortb

import (
	"encoding/json"

	"github.com/prebid/prebid-server/openrtb_ext"
)

type ExtRegs struct {
	// GDPR should be "1" if the caller believes the user is subject to GDPR laws, "0" if not, and undefined
	// if it's unknown. For more info on this parameter, see: https://iabtechlab.com/wp-content/uploads/2018/02/OpenRTB_Advisory_GDPR_2018-02.pdf
	Gdpr *int `json:"gdpr,omitempty"`
	// USPrivacy should be a four character string, see: https://iabtechlab.com/wp-content/uploads/2019/11/OpenRTB-Extension-U.S.-Privacy-IAB-Tech-Lab.pdf
	USPrivacy *string `json:"us_privacy,omitempty"`
}

// ExtRequestAdPod holds AdPod specific extension parameters at request level
type ExtRequestAdPod struct {
	AdPod
	CrossPodAdvertiserExclusionPercent  *int `json:"crosspodexcladv,omitempty"`    //Percent Value - Across multiple impression there will be no ads from same advertiser. Note: These cross pod rule % values can not be more restrictive than per pod
	CrossPodIABCategoryExclusionPercent *int `json:"crosspodexcliabcat,omitempty"` //Percent Value - Across multiple impression there will be no ads from same advertiser
	IABCategoryExclusionWindow          *int `json:"excliabcatwindow,omitempty"`   //Duration in minute between pods where exclusive IAB rule needs to be applied
	AdvertiserExclusionWindow           *int `json:"excladvwindow,omitempty"`      //Duration in minute between pods where exclusive advertiser rule needs to be applied
}

// AdPod holds Video AdPod specific extension parameters at impression level
type AdPod struct {
	MinAds                      *int `json:"minads,omitempty"`        //Default 1 if not specified
	MaxAds                      *int `json:"maxads,omitempty"`        //Default 1 if not specified
	MinDuration                 *int `json:"adminduration,omitempty"` // (adpod.adminduration * adpod.minads) should be greater than or equal to video.minduration
	MaxDuration                 *int `json:"admaxduration,omitempty"` // (adpod.admaxduration * adpod.maxads) should be less than or equal to video.maxduration + video.maxextended
	AdvertiserExclusionPercent  *int `json:"excladv,omitempty"`       // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int `json:"excliabcat,omitempty"`    // Percent value 0 means all ads should be of different IAB categories.
}

// ImpExtension - Impression Extension
type ImpExtension struct {
	Wrapper     *ExtImpWrapper              `json:"wrapper,omitempty"`
	Bidder      map[string]*BidderExtension `json:"bidder,omitempty"`
	SKAdnetwork json.RawMessage             `json:"skadn,omitempty"`
	Reward      *int                        `json:"reward,omitempty"`
	Data        json.RawMessage             `json:"data,omitempty"`
	Prebid      *openrtb_ext.ExtImpPrebid   `json:"prebid,omitempty"`
}

// BidderExtension - Bidder specific items
type BidderExtension struct {
	KeyWords []KeyVal  `json:"keywords,omitempty"`
	DealTier *DealTier `json:"dealtier,omitempty"`
}

// DealTier - Deal information for individual bidders
type DealTier struct {
	Prefix      string `json:"prefix,omitempty"`
	MinDealTier int    `json:"mindealtier,omitempty"`
}

// ExtImpWrapper - Impression wrapper Extension
type ExtImpWrapper struct {
	Div *string `json:"div,omitempty"`
}

// ExtVideo structure to accept video specific more parameters like adpod
type ExtVideo struct {
	Offset *int   `json:"offset,omitempty"` // Minutes from start where this ad is intended to show
	AdPod  *AdPod `json:"adpod,omitempty"`
}

// ExtRequest Request Extension
type ExtRequest struct {
	Wrapper *ExtRequestWrapper                `json:"wrapper,omitempty"`
	Bidder  map[string]map[string]interface{} `json:"bidder,omitempty"`
	AdPod   *ExtRequestAdPod                  `json:"adpod,omitempty"`
	Prebid  *ExtRequestPrebid                 `json:"prebid"`
}

// ExtRequestPrebid defines the contract for bidrequest.ext.prebid
type ExtRequestPrebid struct {
	Aliases              interface{} `json:"aliases,omitempty"`
	BidAdjustmentFactors interface{} `json:"bidadjustmentfactors,omitempty"`
	Cache                interface{} `json:"cache,omitempty"`
	Data                 interface{} `json:"data,omitempty"`
	Debug                bool        `json:"debug,omitempty"`
	Events               interface{} `json:"events,omitempty"`
	SChains              interface{} `json:"schains,omitempty"`
	StoredRequest        interface{} `json:"storedrequest,omitempty"`
	SupportDeals         bool        `json:"supportdeals,omitempty"`
	Targeting            interface{} `json:"targeting,omitempty"`

	// NoSale specifies bidders with whom the publisher has a legal relationship where the
	// passing of personally identifiable information doesn't constitute a sale per CCPA law.
	// The array may contain a single sstar ('*') entry to represent all bidders.
	NoSale       []string         `json:"nosale,omitempty"`
	Transparency *ExtTransparency `json:"transparency,omitempty"`
	// Floors       *PriceFloorRules `json:"floors,omitempty"`
}

// pbopenrtb_ext alias for prebid server openrtb_ext
// type PriceFloorRules = openrtb_ext.PriceFloorRules

// TransparencyRule contains transperancy rule for a single bidder
type TransparencyRule struct {
	Include bool     `json:"include,omitempty"`
	Keys    []string `json:"keys,omitempty"`
}

// ExtTransparency holds bidder level content transparency rules
type ExtTransparency struct {
	Content map[string]TransparencyRule `json:"content,omitempty"`
}

// KeyVal structure to store bidder related custom key-values
type KeyVal struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"value,omitempty"`
}

// ExtRequestWrapper holds wrapper specific extension parameters
type ExtRequestWrapper struct {
	ProfileId            *int    `json:"profileid,omitempty"`
	VersionId            *int    `json:"versionid,omitempty"`
	SSAuctionFlag        *int    `json:"ssauction,omitempty"`
	SumryDisableFlag     *int    `json:"sumry_disable,omitempty"`
	ClientConfigFlag     *int    `json:"clientconfig,omitempty"`
	LogInfoFlag          *int    `json:"loginfo,omitempty"`
	SupportDeals         bool    `json:"supportdeals,omitempty"`
	IncludeBrandCategory *int    `json:"includebrandcategory,omitempty"`
	ABTestConfig         *int    `json:"abtest,omitempty"`
	LoggerImpressionID   *string `json:"wiid,omitempty"`
	SSAI                 *string `json:"ssai,omitempty"`
}
