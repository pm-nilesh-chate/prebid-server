package ortb2blocking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/prebid/openrtb/v17/adcom1"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
)

func Builder(_ json.RawMessage, _ moduledeps.ModuleDeps) (interface{}, error) {
	return Module{}, nil
}

type Module struct {
	// implement a module level cache accross requests here.
	// Use PBS-Core (hookExecutor) for request level content of a module
	moduleCache map[string]interface{}
}

func (m Module) HandleEntrypointHook(
	ctx context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	result := hookstage.HookResult[hookstage.EntrypointPayload]{}

	how, ok := miCtx.ModuleContext["abc"].(int)
	if ok {
		panic(fmt.Sprintf("miCtx.ModuleContext is shared across requests!!! %v %v", how, ok))
	}

	dummyctx := make(map[string]interface{})
	dummyctx["abc"] = 123
	result.ModuleContext = dummyctx

	return result, nil
}

// HandleBidderRequestHook updates blocking fields on the openrtb2.BidRequest.
// Fields are updated only if request satisfies conditions provided by the module config.
func (m Module) HandleBidderRequestHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BidderRequestPayload,
) (hookstage.HookResult[hookstage.BidderRequestPayload], error) {
	result := hookstage.HookResult[hookstage.BidderRequestPayload]{}
	if len(miCtx.AccountConfig) == 0 {
		return result, nil
	}

	cfg, err := newConfig(miCtx.AccountConfig)
	if err != nil {
		return result, err
	}

	return handleBidderRequestHook(cfg, payload)
}

// HandleRawBidderResponseHook rejects bids for a specific bidder if they fail the attribute check.
func (m Module) HandleRawBidderResponseHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (hookstage.HookResult[hookstage.RawBidderResponsePayload], error) {
	result := hookstage.HookResult[hookstage.RawBidderResponsePayload]{}
	var cfg config
	if len(miCtx.AccountConfig) != 0 {
		ncfg, err := newConfig(miCtx.AccountConfig)
		if err != nil {
			return result, err
		}
		cfg = ncfg
	}

	return handleRawBidderResponseHook(cfg, payload, miCtx.ModuleContext)
}

type blockingAttributes struct {
	bAdv   []string
	bApp   []string
	bCat   []string
	bType  map[string][]int
	bAttr  map[string][]int
	catTax adcom1.CategoryTaxonomy
}
