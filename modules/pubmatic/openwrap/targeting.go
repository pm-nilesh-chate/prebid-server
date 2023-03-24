package openwrap

import (
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// whitelist of prebid targeting keys
var prebidTargetingKeysWhitelist = map[string]struct{}{
	string(openrtb_ext.HbpbConstantKey): {},
	models.HbBuyIdPubmaticConstantKey:   {},
	// OTT - 18 Deal priortization support
	// this key required to send deal prefix and priority
	string(openrtb_ext.HbCategoryDurationKey): {},
}

// check if prebid targeting keys are whitelisted
func allowTargetingKey(key string) bool {
	if _, ok := prebidTargetingKeysWhitelist[key]; ok {
		return true
	}
	return strings.HasPrefix(key, models.HbBuyIdPrefix)
}
