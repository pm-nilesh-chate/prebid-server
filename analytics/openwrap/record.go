package openwrap

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	uuid "github.com/satori/go.uuid"
)

// WloggerRecord structure for wrapper analytics logger object
type WloggerRecord struct {
	record
	NonBidRejections map[string]analytics.RejectedBid
}

type record struct {
	Timeout           int              `json:"to,omitempty"`
	PubID             int              `json:"pubid,omitempty"`
	PageURL           string           `json:"purl,omitempty"`
	Timestamp         int64            `json:"tst,omitempty"`
	IID               string           `json:"iid,omitempty"`
	ProfileID         string           `json:"pid,omitempty"`
	VersionID         string           `json:"pdvid,omitempty"`
	IP                string           `json:"-,omitempty"`
	UserAgent         string           `json:"-,omitempty"`
	UID               string           `json:"-,omitempty"`
	GDPR              int              `json:"gdpr,omitempty"`
	ConsentString     string           `json:"cns,omitempty"`
	PubmaticConsent   int              `json:"pmc,omitempty"`
	UserID            string           `json:"uid,omitempty"`
	PageValue         float64          `json:"pv,omitempty"` //sum of all winning bids
	ServerLogger      int              `json:"sl,omitempty"`
	Slots             []SlotRecord     `json:"s,omitempty"`
	CachePutMiss      int              `json:"cm,omitempty"`
	Origin            string           `json:"orig,omitempty"`
	Device            Device           `json:"dvc,omitempty"`
	AdPodPercentage   *AdPodPercentage `json:"aps,omitempty"`
	Content           *Content         `json:"ct,omitempty"`
	TestConfigApplied int              `json:"tgid,omitempty"`
	//Geo             GeoRecord    `json:"geo,omitempty"`
}

// Device struct for storing device information
type Device struct {
	Platform models.DevicePlatform `json:"plt,omitempty"`
	IFAType  *models.DeviceIFAType `json:"ifty,omitempty"` //OTT-416, adding device.ext.ifa_type
}

/*
//GeoRecord structure for storing geo information
type GeoRecord struct {
	CountryCode string `json:"cc,omitempty"`
}
*/

// AdPodPercentage will store adpod percentage value comes in request
type AdPodPercentage struct {
	CrossPodAdvertiserExclusionPercent  *int `json:"cpexap,omitempty"` //Percent Value - Across multiple impression there will be no ads from same advertiser. Note: These cross pod rule % values can not be more restrictive than per pod
	CrossPodIABCategoryExclusionPercent *int `json:"cpexip,omitempty"` //Percent Value - Across multiple impression there will be no ads from same advertiser
	IABCategoryExclusionWindow          *int `json:"exapw,omitempty"`  //Duration in minute between pods where exclusive IAB rule needs to be applied
	AdvertiserExclusionWindow           *int `json:"exipw,omitempty"`  //Duration in minute between pods where exclusive advertiser rule needs to be applied
}

// Content of openrtb request object
type Content struct {
	ID      string   `json:"id,omitempty"`  // ID uniquely identifying the content
	Episode int      `json:"eps,omitempty"` // Episode number (typically applies to video content).
	Title   string   `json:"ttl,omitempty"` // Content title.
	Series  string   `json:"srs,omitempty"` // Content series
	Season  string   `json:"ssn,omitempty"` // Content season
	Cat     []string `json:"cat,omitempty"` // Array of IAB content categories that describe the content producer
}

// AdPodSlot of adpod object logging
type AdPodSlot struct {
	MinAds                      *int `json:"mnad,omitempty"` //Default 1 if not specified
	MaxAds                      *int `json:"mxad,omitempty"` //Default 1 if not specified
	MinDuration                 *int `json:"amnd,omitempty"` // (adpod.adminduration * adpod.minads) should be greater than or equal to video.minduration
	MaxDuration                 *int `json:"amxd,omitempty"` // (adpod.admaxduration * adpod.maxads) should be less than or equal to video.maxduration + video.maxextended
	AdvertiserExclusionPercent  *int `json:"exap,omitempty"` // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int `json:"exip,omitempty"` // Percent value 0 means all ads should be of different IAB categories.
}

// SlotRecord structure for storing slot level information
type SlotRecord struct {
	SlotName          string          `json:"sn,omitempty"`
	SlotSize          []string        `json:"sz,omitempty"`
	Adunit            string          `json:"au,omitempty"`
	AdPodSlot         *AdPodSlot      `json:"aps,omitempty"`
	PartnerData       []PartnerRecord `json:"ps"`
	RewardedInventory int             `json:"rwrd,omitempty"` // Indicates if the ad slot was enabled (rwrd=1) for rewarded or disabled (rwrd=0)
}

// PartnerRecord structure for storing partner information
type PartnerRecord struct {
	PartnerID            string  `json:"pn"`
	BidderCode           string  `json:"bc"`
	KGPV                 string  `json:"kgpv"`  // In case of Regex mapping, this will contain the regex string.
	KGPSV                string  `json:"kgpsv"` // In case of Regex mapping, this will contain the actual slot name that matched the regex.
	PartnerSize          string  `json:"psz"`   //wxh
	Adformat             string  `json:"af"`
	GrossECPM            float64 `json:"eg"`
	NetECPM              float64 `json:"en"`
	Latency1             int     `json:"l1"` //response time
	Latency2             int     `json:"l2"`
	PostTimeoutBidStatus int     `json:"t"`
	WinningBidStaus      int     `json:"wb"`
	BidID                string  `json:"bidid"`
	OrigBidID            string  `json:"origbidid"`
	DealID               string  `json:"di"`
	DealChannel          string  `json:"dc"`
	DealPriority         int     `json:"dp,omitempty"`
	DefaultBidStatus     int     `json:"db"`
	ServerSide           int     `json:"ss"`
	MatchedImpression    int     `json:"mi"`

	//AdPod Specific
	AdPodSequenceNumber *int     `json:"adsq,omitempty"`
	AdDuration          *int     `json:"dur,omitempty"`
	ADomain             string   `json:"adv,omitempty"`
	Cat                 []string `json:"cat,omitempty"`
	NoBidReason         *int     `json:"aprc,omitempty"`

	OriginalCPM float64 `json:"ocpm"`
	OriginalCur string  `json:"ocry"`

	MetaData *MetaData `json:"md,omitempty"`
}

type MetaData struct {
	NetworkID            int             `json:"nwid,omitempty"`
	AdvertiserID         int             `json:"adid,omitempty"`
	NetworkName          string          `json:"nwnm,omitempty"`
	PrimaryCategoryID    string          `json:"pcid,omitempty"`
	AdvertiserName       string          `json:"adnm,omitempty"`
	AgencyID             int             `json:"agid,omitempty"`
	AgencyName           string          `json:"agnm,omitempty"`
	BrandID              int             `json:"brid,omitempty"`
	BrandName            string          `json:"brnm,omitempty"`
	DChain               json.RawMessage `json:"dc,omitempty"`
	DemandSource         string          `json:"ds,omitempty"`
	SecondaryCategoryIDs []string        `json:"secondaryCatIds,omitempty"`
}

// NewRecord returns a new wlogger record
func NewRecord() *WloggerRecord {
	wlog := new(WloggerRecord)
	wlog.SetIID(uuid.NewV4().String())
	wlog.SetTimestamp(int64(time.Now().Unix()))
	wlog.SetServerLogger(1)
	wlog.initNonBidRejections()
	return wlog
}

func (wlog *WloggerRecord) initNonBidRejections() {
	wlog.NonBidRejections = make(map[string]analytics.RejectedBid)
}

// String returns string object
func (wlog *WloggerRecord) String() string {
	byts, _ := json.Marshal(wlog)
	return string(byts)
}

// SetTimeout sets timeout in WloggerRecord
func (wlog *WloggerRecord) SetTimeout(timeout int) {
	wlog.Timeout = timeout
}

// SetUID sets uid in WloggerRecord
func (wlog *WloggerRecord) SetUID(uid string) {
	wlog.UID = uid
}

// SetProfileID sets timeout in WloggerRecord
func (wlog *WloggerRecord) SetProfileID(profileID string) {
	wlog.ProfileID = profileID
}

// SetVersionID sets versionId in WloggerRecord
func (wlog *WloggerRecord) SetVersionID(versionID string) {
	wlog.VersionID = versionID
}

// SetPubID sets pubid in WloggerRecord
func (wlog *WloggerRecord) SetPubID(pubID int) {
	wlog.PubID = pubID
}

// SetGDPR sets GDPR in WloggerRecord
func (wlog *WloggerRecord) SetGDPR(gdpr int) {
	wlog.GDPR = gdpr
}

// SetConsentString sets ConsentString in WloggerRecord
func (wlog *WloggerRecord) SetConsentString(cns string) {
	wlog.ConsentString = cns
}

// SetUserAgent sets user-agent in WloggerRecord
func (wlog *WloggerRecord) SetUserAgent(ua string) {
	wlog.UserAgent = ua
}

// SetIP sets IP in WloggerRecord
func (wlog *WloggerRecord) SetIP(ip string) {
	wlog.IP = ip
}

// SetPageURL sets PageURL in WloggerRecord
func (wlog *WloggerRecord) SetPageURL(pageURL string) {
	wlog.PageURL = pageURL
}

// SetOrigin sets Origin in WloggerRecord
func (wlog *WloggerRecord) SetOrigin(origin string) {
	wlog.Origin = origin
}

// SetIID sets iid in WloggerRecord
func (wlog *WloggerRecord) SetIID(IID string) {
	wlog.IID = IID
}

// SetTimestamp sets Timestamp in WloggerRecord
func (wlog *WloggerRecord) SetTimestamp(timestamp int64) {
	wlog.Timestamp = timestamp
}

// SetSlots sets slots in WloggerRecord
func (wlog *WloggerRecord) SetSlots(slots []SlotRecord) {
	wlog.Slots = slots
}

// SetServerLogger sets server logger enabled/disabled in WloggerRecord
func (wlog *WloggerRecord) SetServerLogger(ss int) {
	wlog.ServerLogger = ss
}

// SetCachePutMiss sets Cache Put miss flag in WloggerRecord
func (wlog *WloggerRecord) SetCachePutMiss(cachePutMiss int) {
	wlog.CachePutMiss = cachePutMiss
}

// SetTestConfigApplied sets tgid in WloggerRecord
func (wlog *WloggerRecord) SetTestConfigApplied(testFlag int) {
	wlog.TestConfigApplied = testFlag
}

// logDeviceObject will be used to log device specific parameters like platform and ifa_type
func (wlog *WloggerRecord) logDeviceObject(rctx models.RequestCtx, uaFromHTTPReq string, ortbBidRequest *openrtb2.BidRequest, platform string) {
	dvc := Device{
		Platform: rctx.DevicePlatform,
	}

	if ortbBidRequest != nil && ortbBidRequest.Device != nil && ortbBidRequest.Device.Ext != nil {
		ext := make(map[string]interface{})
		err := json.Unmarshal(ortbBidRequest.Device.Ext, &ext)
		if err != nil {
			return

		}
		// if ext, ok := ortbBidRequest.Device.Ext.(map[string]interface{}); ok {
		//use ext object for logging any other extension parameters

		//log device.ext.ifa_type parameter to ifty in logger record
		if value, ok := ext["ifa_type"].(string); ok {

			//ifa_type checkking is valid parameter and log its respective id
			ifaType := models.DeviceIFATypeID[strings.ToLower(value)]
			dvc.IFAType = &ifaType
		}
		// }
	}

	//settind device object
	wlog.Device = dvc
}

// ParseRequestCookies parse request cookies
// func ParseRequestCookies(HTTPRequest *http.Request, partnerConfigMap map[int]map[string]string) map[string]int {
// 	cookieFlagMap := make(map[string]int)
// 	pc := usersync.ParseCookieFromRequest(HTTPRequest, &config.HostCookie{})
// 	for _, partnerConfig := range partnerConfigMap {
// 		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
// 			continue
// 		}

// 		syncerMap := router.SyncerMap()
// 		partnerName := partnerConfig[models.PREBID_PARTNER_NAME]

// 		syncerCode := adapters.ResolveOWBidder(partnerName)

// 		matchedImpression := 0

// 		syncer := syncerMap[syncerCode]
// 		if syncer == nil {
// 		} else {
// 			uid, _, _ := pc.GetUID(syncer.Key())

// 			// Added flag in map for Cookie is present
// 			// we are not considering if the cookie is active
// 			if uid != "" {
// 				matchedImpression = 1
// 			}
// 		}

// 		cookieFlagMap[partnerConfig[models.BidderCode]] = matchedImpression
// 	}

// 	return cookieFlagMap
// }
