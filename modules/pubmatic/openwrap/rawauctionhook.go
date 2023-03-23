package openwrap

import (
	"context"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

func (m OpenWrap) handleRawAuctionHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawAuctionRequestPayload,
) (hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	result := hookstage.HookResult[hookstage.RawAuctionRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.RawAuctionRequestPayload]{}

	return result, nil
}
