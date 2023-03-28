package openwrap

import (
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func getMarketplaceBidders(reqABC *openrtb_ext.ExtAlternateBidderCodes, partnerConfigMap map[int]map[string]string) *openrtb_ext.ExtAlternateBidderCodes {
	if reqABC != nil {
		return reqABC
	}

	// string validations, etc will be done by api-wrapper-tag. Not need to repetitively do the typical string validations
	marketplaceBiddersDB := partnerConfigMap[models.VersionLevelConfigID][models.MarketplaceBidders]
	if len(marketplaceBiddersDB) == 0 {
		return nil
	}
	marketplaceBidders := strings.Split(marketplaceBiddersDB, ",")

	return &openrtb_ext.ExtAlternateBidderCodes{
		Enabled: true,
		Bidders: map[string]openrtb_ext.ExtAdapterAlternateBidderCodes{
			string(openrtb_ext.BidderPubmatic): { // How do we get non-pubmatic bidders?. Does core-platform even have it?
				Enabled:            true,
				AllowedBidderCodes: marketplaceBidders,
			},
		},
	}
}
