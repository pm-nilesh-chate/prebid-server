package adapters

import (
	"encoding/json"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// BidderParameters provides all properties requires for bidder to generate bidder json
type BidderParameters struct {
	//AdapterName, BidderCode should be passed in builder function
	ReqID                   string
	AdapterName, BidderCode string
	ImpExt                  *models.ImpExtension

	//bidder specific parameters
	FieldMap      JSONObject
	Width, Height *int64
	SlotKey       string
}

// JSONObject generic JSON object
type JSONObject = map[string]interface{}

// builder callback type
type builder func(params BidderParameters) (json.RawMessage, error)

// bidderBuilderFactor
var _bidderBuilderFactory map[string]builder

// initBidderBuilderFactory initialise all hard coded bidder builder
func initBidderBuilderFactory() {
	_bidderBuilderFactory = map[string]builder{
		models.BidderAdForm:          builderAdform,
		models.BidderAdf:             builderAdform,
		models.BidderAppnexus:        builderAppNexus,
		models.BidderBeachfront:      builderBeachfront,
		models.BidderCriteo:          builderCriteo,
		models.BidderGumGum:          builderGumGum,
		models.BidderImproveDigitial: builderImproveDigital,
		models.BidderIndex:           builderIndex,
		models.BidderOpenX:           builderOpenx,
		models.BidderOutbrain:        builderOutbrain,
		models.BidderPangle:          builderPangle,
		models.BidderPubMatic:        builderPubMatic, /*this only gets used incase of hybrid case*/
		models.BidderPulsePoint:      builderPulsePoint,
		models.BidderRubicon:         builderRubicon,
		models.BidderSmaato:          builderSmaato,
		models.BidderSmartAdServer:   builderSmartAdServer,
		models.BidderSonobi:          builderSonobi,
		models.BidderSovrn:           builderSovrn,
		models.BidderApacdex:         builderApacdex,
	}
}

// getBuilder will return core bidder hard coded builder, if not found then returns default builder
func getBuilder(adapterName string) builder {
	//resolve hardcoded bidder alias
	adapterName = ResolveOWBidder(adapterName)

	if callback, ok := _bidderBuilderFactory[adapterName]; ok {
		return callback
	}
	return defaultBuilder
}

// InitBidders will initialise bidder alias, default bidder parameter json and builders for each bidder
func InitBidders() {
	initBidderBuilderFactory()
	parseBidderParams()
}
