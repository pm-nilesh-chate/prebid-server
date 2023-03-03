package exchange

import (
	"fmt"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (targData *targetData) setTargetingOW(auc *auction, isApp bool, categoryMapping map[string]string, truncateTargetAttr *int) {
	for impId, topBidsPerImp := range auc.winningBidsByBidder {
		overallWinner := auc.winningBids[impId]
		for bidderName, topBidPerBidder := range topBidsPerImp {
			isOverallWinner := overallWinner == topBidPerBidder

			targets := make(map[string]string, 10)
			targets[CreatePartnerKey(bidderName.String(), models.PWT_SLOTID)] = impId
			targets[CreatePartnerKey(bidderName.String(), models.PWT_SZ)] = GetSize(int(topBidPerBidder.Bid.W), int(topBidPerBidder.Bid.H))
			targets[CreatePartnerKey(bidderName.String(), models.PWT_PARTNERID)] = bidderName.String()
			// revShare := GetRevenueShare(partnerNameMap[*winningSeat])
			// netEcpm := GetNetEcpm(*winningvBidv25.Price, revShare)
			// targets[CreatePartnerKey(bidderName.String(), models.PWT_ECPM)] = fmt.Sprintf("%.2f", netEcpm)
			targets[CreatePartnerKey(bidderName.String(), models.PWT_PLATFORM)] = getPlatformName(platform)
			targets[CreatePartnerKey(bidderName.String(), models.PWT_BIDSTATUS)] = "1"
			if bid.DealId != nil {
				targets[CreatePartnerKey(bidderName.String(), models.PWT_DEALID)] = *bid.DealId
			}

			if isOverallWinner {
				targets[models.PWT_SLOTID] = impId
				targets[models.PWT_BIDSTATUS] = "1"
				targets[models.PWT_SZ] = GetSize(*bid.W, *bid.H)
				targets[models.PWT_PARTNERID] = bidderName.String()
				targets[models.PWT_ECPM] = fmt.Sprintf("%.2f", netEcpm)
				targets[models.PWT_PLATFORM] = getPlatformName(platform)
				if bid.DealId != nil {
					targets[models.PWT_DEALID] = *bid.DealId
				}
			}

			topBidPerBidder.BidTargets = targets
		}
	}
}

// CreatePartnerKey returns key with bidderName.String() appended
func CreatePartnerKey(bidderName, key string) string {
	if bidderName == "" {
		return key
	}
	return key + "_" + bidderName
}

func GetSize(width int, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}
