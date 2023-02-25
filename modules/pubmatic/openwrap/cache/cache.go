package cache

import (
	"github.com/prebid/openrtb/v17/openrtb2"
)

type Cache interface {
	GetPartnerConfigMap(bidRequest *openrtb2.BidRequest, pubid, profileid, displayversion int) map[int]map[string]string
}

// type internalCache interface {
// }
