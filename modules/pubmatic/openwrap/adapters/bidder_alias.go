package adapters

import "github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

// Alias will return copy of exisiting alias
var Alias map[string]string

func init() {
	Alias = map[string]string{
		models.BidderAdGenerationAlias:      models.BidderAdGeneration,
		models.BidderDistrictmDMXAlias:      models.BidderDistrictmDMX,
		models.BidderPubMaticSecondaryAlias: models.BidderPubMatic,
		models.BidderDistrictmAlias:         models.BidderAppnexus,
		models.BidderAndBeyondAlias:         models.BidderAdKernel,
	}
}

//ResolveOWBidder it resolves hardcoded bidder alias names

func ResolveOWBidder(bidderName string) string {

	var coreBidderName string

	switch bidderName {

	case models.BidderAdGenerationAlias:
		coreBidderName = models.BidderAdGeneration
	case models.BidderDistrictmDMXAlias:
		coreBidderName = models.BidderDistrictmDMX
	case models.BidderPubMaticSecondaryAlias:
		coreBidderName = models.BidderPubMatic
	case models.BidderDistrictmAlias:
		coreBidderName = models.BidderAppnexus
	case models.BidderAndBeyondAlias:
		coreBidderName = models.BidderAdKernel
	default:
		coreBidderName = bidderName

	}

	return coreBidderName
}
