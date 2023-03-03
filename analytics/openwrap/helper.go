package openwrap

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func CreateCommonLogger(ao *analytics.AuctionObject) *WloggerRecord {
	wl := NewRecord()
	// uaFromHTTPReq := ao.Request.Header.Get("User-Agent")
	wl.CreateLoggerRecordFromRequest("uaFromHTTPReq", ao)
	wl.SetTimeout(int(ao.Request.TMax))
	// if uidCookie, _ := controller.HTTPRequest.Cookie(models.KADUSERCOOKIE); uidCookie != nil {
	// 	controller.LoggerRecord.SetUID(uidCookie.Value)
	// }
	// if testConfigApplied {
	// 	controller.LoggerRecord.SetTestConfigApplied(1)
	// }
	return wl
}

// CreateLoggerRecordFromRequest creates logger and tracker records from bidRequest data
func (wlog *WloggerRecord) CreateLoggerRecordFromRequest(uaFromHTTPReq string, ao *analytics.AuctionObject) {
	extWrapper := models.RequestExtWrapper{}
	err := json.Unmarshal(ao.Request.Ext, &extWrapper)
	if err != nil {
		return
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

	wlog.SetProfileID(strconv.Itoa(extWrapper.ProfileId))
	wlog.SetVersionID(strconv.Itoa(extWrapper.VersionId))

	var consent string
	if ao.Request.User != nil {
		extUser := openrtb_ext.UserExt{}
		err := json.Unmarshal(ao.Request.User.Ext, &extWrapper)
		if err != nil {
			return
		}
		consent = *extUser.GetConsent()
	}
	wlog.SetConsentString(consent)

	var gdpr int8
	if ao.Request.Regs != nil {
		extReg := openrtb_ext.RegExt{}
		err := json.Unmarshal(ao.Request.Regs.Ext, &extWrapper)
		if err != nil {
			return
		}
		gdpr = *extReg.GetGDPR()
	}
	wlog.SetGDPR(int(gdpr))

	if ao.Request.Device != nil {
		wlog.SetIP(ao.Request.Device.IP)
		wlog.SetUserAgent(ao.Request.Device.UA)
	}

	//log device object
	// wlog.logDeviceObject(uaFromHTTPReq, ao.Request, platform)

	//log content object
	// if nil != ao.Request.Site {
	// 	wlog.logContentObject(ao.Request.Site.Content)
	// } else if nil != ao.Request.App {
	// 	wlog.logContentObject(ao.Request.App.Content)
	// }

	// //log adpod percentage object
	// if nil != ao.Request.Ext {
	// 	ext, ok := ao.Request.Ext.(*openrtb.ExtRequest)
	// 	if ok {
	// 		wlog.logAdPodPercentage(ext.AdPod)
	// 	}
	// }
}

// Send method
func Send(client http.Client, loggerURL string, wl *WloggerRecord, gdprEnabled int) error {
	loggerURL = PrepareLoggerURL(wl, loggerURL, gdprEnabled)
	hc, err := http.NewRequest(http.MethodGet, loggerURL, nil)
	if err != nil {
		return err
	}

	hc.Header.Add(models.USER_AGENT_HEADER, wl.UserAgent)
	hc.Header.Add(models.IP_HEADER, wl.IP)
	if wl.UID != "" {
		hc.Header.Add(models.KADUSERCOOKIE, wl.UID)
	}

	_, err = client.Do(hc)
	if err != nil {
		return errors.New("error in sending logger pixel")
	}
	return nil
}

// PrepareLoggerURL returns the url for OW logger call
func PrepareLoggerURL(wlog *WloggerRecord, loggerURL string, gdprEnabled int) string {
	v := url.Values{}

	jsonString, err := json.Marshal(wlog.record)
	if err != nil {
		return ""
	}

	v.Set(models.WLJSON, string(jsonString))
	v.Set(models.WLPUBID, strconv.Itoa(wlog.PubID))
	if gdprEnabled == 1 {
		v.Set(models.WLGDPR, strconv.Itoa(gdprEnabled))
	}
	queryString := v.Encode()

	finalLoggerURL := loggerURL + "?" + queryString
	return finalLoggerURL
}

// GetString converts interface to string if it is compatible
func GetString(val interface{}) string {
	var result string
	if val != nil {
		result, _ = val.(string)
	}
	return result
}

// GetInt converts interface to int if it is compatible
func GetInt(val interface{}) int {
	var result int
	if val != nil {
		switch val.(type) {
		case int:
			result = val.(int)
		case float64:
			val := val.(float64)
			result = int(val)
		case float32:
			val := val.(float32)
			result = int(val)
		}
	}
	return result
}
