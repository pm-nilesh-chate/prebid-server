package openwrap

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
)

// init openwrap module and its dependecies like config, cache, db connection, bidder cfg, etc.
func Builder(rawCfg json.RawMessage, deps moduledeps.ModuleDeps) (interface{}, error) {
	return initOpenWrap(rawCfg, deps)
}

// temporary openwrap changes to support non-pbs apis like openrtb/2.5, openrtb/amp, etc
// temporary openwrap changes to support non-ortb fields like request.ext.wrapper
func (m OpenWrap) HandleEntrypointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("body:" + string(payload.Body) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleEntrypointHook(ctx, miCtx, payload)
}

// changes to init the request ctx with profile and request details
func (m OpenWrap) HandleRawAuctionHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawAuctionRequestPayload,
) (hookstage.HookResult[hookstage.RawAuctionRequestPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("body:" + string(payload) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleRawAuctionHook(ctx, miCtx, payload)
}

// changes to init the request ctx with profile and request details
func (m OpenWrap) HandleBeforeValidationHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	defer func() {
		if r := recover(); r != nil {
			request, err := json.Marshal(payload)
			if err != nil {
				glog.Error("request:" + string(request) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
				return
			}
			glog.Error("request:" + string(request) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleBeforeValidationHook(ctx, miCtx, payload)
}

func (m OpenWrap) HandleAuctionResponseHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	defer func() {
		if r := recover(); r != nil {
			response, err := json.Marshal(payload)
			if err != nil {
				glog.Error("response:" + string(response) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
				return
			}
			glog.Error("response:" + string(response) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	return m.handleAuctionResponseHook(ctx, miCtx, payload)
}
