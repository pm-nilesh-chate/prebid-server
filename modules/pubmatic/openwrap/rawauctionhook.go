package openwrap

import (
	"context"
	"fmt"
	"strconv"

	pbsOpenrtb2 "github.com/prebid/prebid-server/endpoints/openrtb2"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
)

func (m OpenWrap) handleRawAuctionHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawAuctionRequestPayload,
) (hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	result := hookstage.HookResult[hookstage.RawAuctionRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.RawAuctionRequestPayload]{}

	// accountID/PublisherID validation not needed. Already done by PBS-Core
	accountID, _, _ := pbsOpenrtb2.SearchAccountId(payload)

	pubID, err := strconv.Atoi(accountID)
	if err != nil {
		result.Reject = true
		result.NbrCode = errorcodes.ErrInvalidPublisherID.Code()
		result.Errors = append(result.Errors, errorcodes.ErrInvalidPublisherID.Error())
		return result, fmt.Errorf("invalid publisher id : %v", err)
	}

	// key presence validationnot needed as failure of EntryPointHook sets result.Reject = true
	rCtx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	rCtx.PubID = pubID
	moduleCtx.ModuleContext["rctx"] = rCtx

	return result, nil
}
