package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

func (m OpenWrap) handleAuctionResponseHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	result := hookstage.HookResult[hookstage.AuctionResponsePayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.AuctionResponsePayload]{}
	result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
		// rctx := result.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		// ap.BidResponse, err = m.updateORTBV25Request(rctx, payload.Body)
		return ap, err
	}, hookstage.MutationUpdate, "response-body-with-sshb-format")

	return result, nil
}
