package openwrap

import (
	"context"
	"fmt"
	"strconv"

	pbsOpenrtb2 "github.com/prebid/prebid-server/endpoints/openrtb2"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
)

func (m OpenWrap) handleRawAuctionHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawAuctionRequestPayload,
) (hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	result := hookstage.HookResult[hookstage.RawAuctionRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.RawAuctionRequestPayload]{}

	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	rCtx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	defer func() {
		moduleCtx.ModuleContext["rctx"] = rCtx
	}()

	// accountID/PublisherID validation not needed. Already done by PBS-Core
	accountID, _, _ := pbsOpenrtb2.SearchAccountId(payload)

	pubID, err := strconv.Atoi(accountID)
	if err != nil {
		result.Reject = true
		result.NbrCode = nbr.InvalidPublisherID
		result.Errors = append(result.Errors, "ErrInvalidPublisherID")
		return result, fmt.Errorf("invalid publisher id : %v", err)
	}
	rCtx.PubID = pubID

	return result, nil
}
