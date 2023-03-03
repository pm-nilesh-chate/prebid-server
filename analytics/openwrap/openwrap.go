package openwrap

import (
	"net/http"

	"github.com/prebid/prebid-server/analytics"
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

// Writes AuctionObject to file
func (ow *HTTPLogger) LogAuctionObject(ao *analytics.AuctionObject) {
	wl := CreateCommonLogger(ao)
	Send(*ow.client, ow.URL, wl, 1) // NYC_TODO: pass gdpr enabled in ao.Context
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
