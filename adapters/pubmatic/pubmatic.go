package pubmatic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const MAX_IMPRESSIONS_PUBMATIC = 30

const (
	PUBMATIC            = "[PUBMATIC]"
	buyId               = "buyid"
	buyIdTargetingKey   = "hb_buyid_"
	skAdnetworkKey      = "skadn"
	rewardKey           = "reward"
	dctrKeywordName     = "dctr"
	urlEncodedEqualChar = "%3D"
	AdServerKey         = "adserver"
	PBAdslotKey         = "pbadslot"
)

type PubmaticAdapter struct {
	URI string
}

type pubmaticBidExt struct {
	BidType            *int                 `json:"BidType,omitempty"`
	VideoCreativeInfo  *pubmaticBidExtVideo `json:"video,omitempty"`
	Marketplace        string               `json:"marketplace,omitempty"`
	PrebidDealPriority int                  `json:"prebiddealpriority,omitempty"`
}

type pubmaticWrapperExt struct {
	ProfileID int `json:"profile,omitempty"`
	VersionID int `json:"version,omitempty"`

	WrapperImpID string `json:"wiid,omitempty"`
}

type pubmaticBidExtVideo struct {
	Duration *int `json:"duration,omitempty"`
}

type ExtImpBidderPubmatic struct {
	adapters.ExtImpBidder
	Data        json.RawMessage `json:"data,omitempty"`
	SKAdnetwork json.RawMessage `json:"skadn,omitempty"`
}

type ExtAdServer struct {
	Name   string `json:"name"`
	AdSlot string `json:"adslot"`
}

type marketplaceReqExt struct {
	AllowedBidders []string `json:"allowedbidders,omitempty"`
}

type extRequestAdServer struct {
	Wrapper     *pubmaticWrapperExt `json:"wrapper,omitempty"`
	Acat        []string            `json:"acat,omitempty"`
	Marketplace *marketplaceReqExt  `json:"marketplace,omitempty"`
	openrtb_ext.ExtRequest
}

const (
	dctrKeyName              = "key_val"
	pmZoneIDKeyName          = "pmZoneId"
	pmZoneIDRequestParamName = "pmzoneid"
	ImpExtAdUnitKey          = "dfp_ad_unit_code"
	AdServerGAM              = "gam"
)

func (a *PubmaticAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	errs := make([]error, 0, len(request.Imp))

	pubID := ""
	extractWrapperExtFromImp := true
	extractPubIDFromImp := true

	newReqExt, cookies, err := extractPubmaticExtFromRequest(request)
	if err != nil {
		return nil, []error{err}
	}
	wrapperExt := newReqExt.Wrapper
	if wrapperExt != nil && wrapperExt.ProfileID != 0 && wrapperExt.VersionID != 0 {
		extractWrapperExtFromImp = false
	}

	for i := 0; i < len(request.Imp); i++ {
		wrapperExtFromImp, pubIDFromImp, err := parseImpressionObject(&request.Imp[i], extractWrapperExtFromImp, extractPubIDFromImp)

		// If the parsing is failed, remove imp and add the error.
		if err != nil {
			errs = append(errs, err)
			request.Imp = append(request.Imp[:i], request.Imp[i+1:]...)
			i--
			continue
		}

		if extractWrapperExtFromImp {
			if wrapperExtFromImp != nil {
				if wrapperExt == nil {
					wrapperExt = &pubmaticWrapperExt{}
				}
				if wrapperExt.ProfileID == 0 {
					wrapperExt.ProfileID = wrapperExtFromImp.ProfileID
				}
				if wrapperExt.VersionID == 0 {
					wrapperExt.VersionID = wrapperExtFromImp.VersionID
				}

				if wrapperExt.WrapperImpID == "" {
					wrapperExt.WrapperImpID = wrapperExtFromImp.WrapperImpID
				}

				if wrapperExt != nil && wrapperExt.ProfileID != 0 && wrapperExt.VersionID != 0 {
					extractWrapperExtFromImp = false
				}
			}
		}

		if extractPubIDFromImp && pubIDFromImp != "" {
			pubID = pubIDFromImp
			extractPubIDFromImp = false
		}
	}

	// If all the requests are invalid, Call to adaptor is skipped
	if len(request.Imp) == 0 {
		return nil, errs
	}

	newReqExt.Wrapper = wrapperExt
	rawExt, err := json.Marshal(newReqExt)
	if err != nil {
		return nil, []error{err}
	}
	request.Ext = rawExt

	if request.Site != nil {
		siteCopy := *request.Site
		if siteCopy.Publisher != nil {
			publisherCopy := *siteCopy.Publisher
			publisherCopy.ID = pubID
			siteCopy.Publisher = &publisherCopy
		} else {
			siteCopy.Publisher = &openrtb2.Publisher{ID: pubID}
		}
		request.Site = &siteCopy
	} else if request.App != nil {
		appCopy := *request.App
		if appCopy.Publisher != nil {
			publisherCopy := *appCopy.Publisher
			publisherCopy.ID = pubID
			appCopy.Publisher = &publisherCopy
		} else {
			appCopy.Publisher = &openrtb2.Publisher{ID: pubID}
		}
		request.App = &appCopy
	}

	// move user.ext.eids to user.eids
	if request.User != nil && request.User.Ext != nil {
		var userExt *openrtb_ext.ExtUser
		if err = json.Unmarshal(request.User.Ext, &userExt); err == nil {
			if userExt != nil && userExt.Eids != nil {
				var eidArr []openrtb2.EID
				for _, eid := range userExt.Eids {
					newEid := &openrtb2.EID{
						ID:     eid.ID,
						Source: eid.Source,
						Ext:    eid.Ext,
					}
					var uidArr []openrtb2.UID
					for _, uid := range eid.UIDs {
						newUID := &openrtb2.UID{
							ID:    uid.ID,
							AType: uid.AType,
							Ext:   uid.Ext,
						}
						uidArr = append(uidArr, *newUID)
					}
					newEid.UIDs = uidArr
					eidArr = append(eidArr, *newEid)
				}

				user := *request.User
				user.EIDs = eidArr
				userExt.Eids = nil
				updatedUserExt, err1 := json.Marshal(userExt)
				if err1 == nil {
					user.Ext = updatedUserExt
				}
				request.User = &user
			}
		}
	}

	//adding hack to support DNT, since hbopenbid does not support lmt
	if request.Device != nil && request.Device.Lmt != nil && *request.Device.Lmt != 0 {
		request.Device.DNT = request.Device.Lmt
	}

	reqJSON, err := json.Marshal(request)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	for _, line := range cookies {
		headers.Add("Cookie", line)
	}
	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     a.URI,
		Body:    reqJSON,
		Headers: headers,
	}}, errs
}

// validateAdslot validate the optional adslot string
// valid formats are 'adslot@WxH', 'adslot' and no adslot
func validateAdSlot(adslot string, imp *openrtb2.Imp) error {
	adSlotStr := strings.TrimSpace(adslot)

	if len(adSlotStr) == 0 {
		return nil
	}

	if !strings.Contains(adSlotStr, "@") {
		imp.TagID = adSlotStr
		return nil
	}

	adSlot := strings.Split(adSlotStr, "@")
	if len(adSlot) == 2 && adSlot[0] != "" && adSlot[1] != "" {
		imp.TagID = strings.TrimSpace(adSlot[0])

		adSize := strings.Split(strings.ToLower(adSlot[1]), "x")
		if len(adSize) != 2 {
			return errors.New(fmt.Sprintf("Invalid size provided in adSlot %v", adSlotStr))
		}

		width, err := strconv.Atoi(strings.TrimSpace(adSize[0]))
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid width provided in adSlot %v", adSlotStr))
		}

		heightStr := strings.Split(adSize[1], ":")
		height, err := strconv.Atoi(strings.TrimSpace(heightStr[0]))
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid height provided in adSlot %v", adSlotStr))
		}

		//In case of video, size could be derived from the player size
		if imp.Banner != nil && width != 0 && height != 0 {
			imp.Banner = assignBannerWidthAndHeight(imp.Banner, int64(width), int64(height))
		}
	} else {
		return errors.New(fmt.Sprintf("Invalid adSlot %v", adSlotStr))
	}

	return nil
}

func assignBannerSize(banner *openrtb2.Banner) (*openrtb2.Banner, error) {
	if banner.W != nil && banner.H != nil {
		return banner, nil
	}

	if len(banner.Format) == 0 {
		return nil, errors.New(fmt.Sprintf("No sizes provided for Banner %v", banner.Format))
	}

	return assignBannerWidthAndHeight(banner, banner.Format[0].W, banner.Format[0].H), nil
}

func assignBannerWidthAndHeight(banner *openrtb2.Banner, w, h int64) *openrtb2.Banner {
	bannerCopy := *banner
	bannerCopy.W = openrtb2.Int64Ptr(w)
	bannerCopy.H = openrtb2.Int64Ptr(h)
	return &bannerCopy
}

// parseImpressionObject parse the imp to get it ready to send to pubmatic
func parseImpressionObject(imp *openrtb2.Imp, extractWrapperExtFromImp, extractPubIDFromImp bool) (*pubmaticWrapperExt, string, error) {
	var wrapExt *pubmaticWrapperExt
	var pubID string

	// PubMatic supports banner and video impressions.
	if imp.Banner == nil && imp.Video == nil && imp.Native == nil {
		return wrapExt, pubID, fmt.Errorf("invalid MediaType. PubMatic only supports Banner, Video and Native. Ignoring ImpID=%s", imp.ID)
	}

	if imp.Audio != nil {
		imp.Audio = nil
	}

	var bidderExt ExtImpBidderPubmatic
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return wrapExt, pubID, err
	}

	var pubmaticExt openrtb_ext.ExtImpPubmatic
	if err := json.Unmarshal(bidderExt.Bidder, &pubmaticExt); err != nil {
		return wrapExt, pubID, err
	}

	if extractPubIDFromImp {
		pubID = strings.TrimSpace(pubmaticExt.PublisherId)
	}

	// Parse Wrapper Extension only once per request
	if extractWrapperExtFromImp && len(pubmaticExt.WrapExt) != 0 {
		err := json.Unmarshal([]byte(pubmaticExt.WrapExt), &wrapExt)
		if err != nil {
			return wrapExt, pubID, fmt.Errorf("Error in Wrapper Parameters = %v  for ImpID = %v WrapperExt = %v", err.Error(), imp.ID, string(pubmaticExt.WrapExt))
		}
	}

	if err := validateAdSlot(strings.TrimSpace(pubmaticExt.AdSlot), imp); err != nil {
		return wrapExt, pubID, err
	}

	if imp.Banner != nil {
		bannerCopy, err := assignBannerSize(imp.Banner)
		if err != nil {
			return wrapExt, pubID, err
		}
		imp.Banner = bannerCopy
	}

	if pubmaticExt.Kadfloor != "" {
		bidfloor, err := strconv.ParseFloat(strings.TrimSpace(pubmaticExt.Kadfloor), 64)
		if err == nil {
			//do not overwrite existing value if kadfloor is invalid
			imp.BidFloor = bidfloor
		}
	}

	extMap := make(map[string]interface{}, 0)
	if pubmaticExt.Keywords != nil && len(pubmaticExt.Keywords) != 0 {
		addKeywordsToExt(pubmaticExt.Keywords, extMap)
	}
	//Give preference to direct values of 'dctr' & 'pmZoneId' params in extension
	if pubmaticExt.Dctr != "" {
		extMap[dctrKeyName] = pubmaticExt.Dctr
	}
	if pubmaticExt.PmZoneID != "" {
		extMap[pmZoneIDKeyName] = pubmaticExt.PmZoneID
	}

	if bidderExt.SKAdnetwork != nil {
		extMap[skAdnetworkKey] = bidderExt.SKAdnetwork
	}

	if bidderExt.Prebid != nil && bidderExt.Prebid.IsRewardedInventory != nil && *bidderExt.Prebid.IsRewardedInventory == 1 {
		extMap[rewardKey] = *bidderExt.Prebid.IsRewardedInventory
	}

	if len(bidderExt.Data) > 0 {
		populateFirstPartyDataImpAttributes(bidderExt.Data, extMap)
	}

	imp.Ext = nil
	if len(extMap) > 0 {
		ext, err := json.Marshal(extMap)
		if err == nil {
			imp.Ext = ext
		}
	}

	return wrapExt, pubID, nil
}

// extractPubmaticExtFromRequest parse the req.ext to fetch wrapper and acat params
func extractPubmaticExtFromRequest(request *openrtb2.BidRequest) (extRequestAdServer, []string, error) {
	var cookies []string

	// req.ext.prebid would always be there and Less nil cases to handle, more safe!
	var pmReqExt extRequestAdServer

	if request == nil || len(request.Ext) == 0 {
		return pmReqExt, cookies, nil
	}

	reqExt := &openrtb_ext.ExtRequest{}
	err := json.Unmarshal(request.Ext, &reqExt)
	if err != nil {
		return pmReqExt, fmt.Errorf("error decoding Request.ext : %s", err.Error())
	}
	pmReqExt.ExtRequest = *reqExt

	reqExtBidderParams := make(map[string]json.RawMessage)
	if reqExt.Prebid.BidderParams != nil {
		err = json.Unmarshal(reqExt.Prebid.BidderParams, &reqExtBidderParams)
		if err != nil {
			return pmReqExt, cookies, nil
		}
	}

	//get request ext bidder params
	if wrapperObj, present := reqExtBidderParams["wrapper"]; present && len(wrapperObj) != 0 {
		wrpExt := &pubmaticWrapperExt{}
		err = json.Unmarshal(wrapperObj, wrpExt)
		if err != nil {
			return pmReqExt, cookies, nil
		}
		pmReqExt.Wrapper = wrpExt
	}

	if acatBytes, ok := reqExtBidderParams["acat"]; ok {
		var acat []string
		err = json.Unmarshal(acatBytes, &acat)
		if err != nil {
			return pmReqExt, cookies, nil
		}
		for i := 0; i < len(acat); i++ {
			acat[i] = strings.TrimSpace(acat[i])
		}
		pmReqExt.Acat = acat
	}

	if allowedBidders := getAlternateBidderCodesFromRequestExt(reqExt); allowedBidders != nil {
		pmReqExt.Marketplace = &marketplaceReqExt{AllowedBidders: allowedBidders}
	}

	// OW patch -start-
	if wiid, ok := reqExtBidderParams["wiid"]; ok {
		if pmReqExt.Wrapper == nil {
			pmReqExt.Wrapper = &pubmaticWrapperExt{}
		}
		pmReqExt.Wrapper.WrapperImpID, _ = strconv.Unquote(string(wiid))
	}
	if wrapperObj, present := reqExtBidderParams["Cookie"]; present && len(wrapperObj) != 0 {
		err = json.Unmarshal(wrapperObj, &cookies)
	}
	// OW patch -end-

	return pmReqExt, cookies, nil
}

func getAlternateBidderCodesFromRequestExt(reqExt *openrtb_ext.ExtRequest) []string {
	if reqExt == nil || reqExt.Prebid.AlternateBidderCodes == nil {
		return nil
	}

	allowedBidders := []string{"pubmatic"}
	if reqExt.Prebid.AlternateBidderCodes.Enabled {
		if pmABC, ok := reqExt.Prebid.AlternateBidderCodes.Bidders["pubmatic"]; ok && pmABC.Enabled {
			if pmABC.AllowedBidderCodes == nil || (len(pmABC.AllowedBidderCodes) == 1 && pmABC.AllowedBidderCodes[0] == "*") {
				return []string{"all"}
			}
			return append(allowedBidders, pmABC.AllowedBidderCodes...)
		}
	}

	return allowedBidders
}

func addKeywordsToExt(keywords []*openrtb_ext.ExtImpPubmaticKeyVal, extMap map[string]interface{}) {
	for _, keyVal := range keywords {
		if len(keyVal.Values) == 0 {
			logf("No values present for key = %s", keyVal.Key)
			continue
		} else {
			key := keyVal.Key
			val := strings.Join(keyVal.Values[:], ",")
			if strings.EqualFold(key, pmZoneIDRequestParamName) {
				key = pmZoneIDKeyName
			} else if key == dctrKeywordName {
				key = dctrKeyName
				// URL-decode dctr value if it is url-encoded
				if strings.Contains(val, urlEncodedEqualChar) {
					urlDecodedVal, err := url.QueryUnescape(val)
					if err == nil {
						val = urlDecodedVal
					}
				}
			}
			extMap[key] = val
		}
	}
}

func (a *PubmaticAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode)}
	}

	var bidResp openrtb2.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(5)

	var errs []error
	for _, sb := range bidResp.SeatBid {
		targets := getTargetingKeys(sb.Ext, string(externalRequest.BidderName))
		for i := 0; i < len(sb.Bid); i++ {
			bid := sb.Bid[i]

			// Copy SeatBid Ext to Bid.Ext
			bid.Ext = copySBExtToBidExt(sb.Ext, bid.Ext)

			if len(bid.Cat) > 1 {
				bid.Cat = bid.Cat[0:1]
			}

			typedBid := &adapters.TypedBid{
				Bid:      &bid,
				BidType:  openrtb_ext.BidTypeBanner,
				BidVideo: &openrtb_ext.ExtBidPrebidVideo{},

				BidTargets: targets,
			}

			var bidExt *pubmaticBidExt
			err := json.Unmarshal(bid.Ext, &bidExt)
			if err != nil {
				errs = append(errs, err)
			} else if bidExt != nil {
				typedBid.Seat = openrtb_ext.BidderName(bidExt.Marketplace)
				typedBid.BidType = getBidType(bidExt)
				if bidExt.PrebidDealPriority > 0 {
					typedBid.DealPriority = bidExt.PrebidDealPriority
				}

				if bidExt.VideoCreativeInfo != nil && bidExt.VideoCreativeInfo.Duration != nil {
					typedBid.BidVideo.Duration = *bidExt.VideoCreativeInfo.Duration
				}
			}

			if typedBid.BidType == openrtb_ext.BidTypeNative {
				bid.AdM, err = getNativeAdm(bid.AdM)
				if err != nil {
					errs = append(errs, err)
				}
			}

			bidResponse.Bids = append(bidResponse.Bids, typedBid)
		}
	}
	return bidResponse, errs
}

func getNativeAdm(adm string) (string, error) {
	var err error
	nativeAdm := make(map[string]interface{})
	err = json.Unmarshal([]byte(adm), &nativeAdm)
	if err != nil {
		return adm, errors.New("unable to unmarshal native adm")
	}

	// move bid.adm.native to bid.adm
	if _, ok := nativeAdm["native"]; ok {
		//using jsonparser to avoid marshaling, encode escape, etc.
		value, _, _, err := jsonparser.Get([]byte(adm), string(openrtb_ext.BidTypeNative))
		if err != nil {
			return adm, errors.New("unable to get native adm")
		}
		adm = string(value)
	}

	return adm, nil
}

//getMapFromJSON converts JSON to map
func getMapFromJSON(source json.RawMessage) map[string]interface{} {
	if source != nil {
		dataMap := make(map[string]interface{})
		err := json.Unmarshal(source, &dataMap)
		if err == nil {
			return dataMap
		}
	}
	return nil
}

//populateFirstPartyDataImpAttributes will parse imp.ext.data and populate imp extMap
func populateFirstPartyDataImpAttributes(data json.RawMessage, extMap map[string]interface{}) {

	dataMap := getMapFromJSON(data)

	if dataMap == nil {
		return
	}

	populateAdUnitKey(data, dataMap, extMap)
	populateDctrKey(dataMap, extMap)
}

//populateAdUnitKey parses data object to read and populate DFP adunit key
func populateAdUnitKey(data json.RawMessage, dataMap, extMap map[string]interface{}) {

	if name, err := jsonparser.GetString(data, "adserver", "name"); err == nil && name == AdServerGAM {
		if adslot, err := jsonparser.GetString(data, "adserver", "adslot"); err == nil && adslot != "" {
			extMap[ImpExtAdUnitKey] = adslot
		}
	}

	//imp.ext.dfp_ad_unit_code is not set, then check pbadslot in imp.ext.data
	if extMap[ImpExtAdUnitKey] == nil && dataMap[PBAdslotKey] != nil {
		extMap[ImpExtAdUnitKey] = dataMap[PBAdslotKey].(string)
	}
}

//populateDctrKey reads key-val pairs from imp.ext.data and add it in imp.ext.key_val
func populateDctrKey(dataMap, extMap map[string]interface{}) {
	var dctr strings.Builder

	//append dctr key if already present in extMap
	if extMap[dctrKeyName] != nil {
		dctr.WriteString(extMap[dctrKeyName].(string))
	}

	for key, val := range dataMap {

		//ignore 'pbaslot' and 'adserver' key as they are not targeting keys
		if key == PBAdslotKey || key == AdServerKey {
			continue
		}

		//separate key-val pairs in dctr string by pipe(|)
		if dctr.String() != "" {
			dctr.WriteString("|")
		}

		//trimming spaces from key
		key = strings.TrimSpace(key)

		switch typedValue := val.(type) {
		case string:
			fmt.Fprintf(&dctr, "%s=%s", key, strings.TrimSpace(typedValue))

		case float64, bool:
			fmt.Fprintf(&dctr, "%s=%v", key, typedValue)

		case []interface{}:
			if isStringArray(typedValue) {
				if valStrArr := getStringArray(typedValue); valStrArr != nil && len(valStrArr) > 0 {
					valStr := strings.Join(valStrArr[:], ",")
					fmt.Fprintf(&dctr, "%s=%s", key, valStr)
				}
			}
		}
	}

	if dctrStr := dctr.String(); dctrStr != "" {
		extMap[dctrKeyName] = strings.TrimSuffix(dctrStr, "|")
	}
}

//isStringArray check if []interface is a valid string array
func isStringArray(array []interface{}) bool {
	for _, val := range array {
		if _, ok := val.(string); !ok {
			return false
		}
	}
	return true
}

//getStringArray converts interface of type string array to string array
func getStringArray(val interface{}) []string {
	aInterface, ok := val.([]interface{})
	if !ok {
		return nil
	}
	aString := make([]string, len(aInterface))
	for i, v := range aInterface {
		if str, ok := v.(string); ok {
			aString[i] = strings.TrimSpace(str)
		}
	}

	return aString
}

// getBidType returns the bid type specified in the response bid.ext
func getBidType(bidExt *pubmaticBidExt) openrtb_ext.BidType {
	// setting "banner" as the default bid type
	bidType := openrtb_ext.BidTypeBanner
	if bidExt != nil && bidExt.BidType != nil {
		switch *bidExt.BidType {
		case 0:
			bidType = openrtb_ext.BidTypeBanner
		case 1:
			bidType = openrtb_ext.BidTypeVideo
		case 2:
			bidType = openrtb_ext.BidTypeNative
		default:
			// default value is banner
			bidType = openrtb_ext.BidTypeBanner
		}
	}
	return bidType
}

func logf(msg string, args ...interface{}) {
	if glog.V(2) {
		glog.Infof(msg, args...)
	}
}

// Builder builds a new instance of the Pubmatic adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	bidder := &PubmaticAdapter{
		URI: config.Endpoint,
	}
	return bidder, nil
}
