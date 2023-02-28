package openwrap

import (
	"context"
	"encoding/json"
	"fmt"

	"errors"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	ow_request "github.com/prebid/prebid-server/modules/pubmatic/openwrap/request"
)

func (m OpenWrap) handleEntrypointHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	result := hookstage.HookResult[hookstage.EntrypointPayload]{}
	if payload.Request.URL.Path == hookexecution.EndpointAuction {
		return result, nil
	}

	result.ChangeSet = hookstage.ChangeSet[hookstage.EntrypointPayload]{}

	requestExtWrapper, err := ow_request.GetWrapperExt(payload.Body)
	if err != nil {
		return result, err
	}

	accountID, err := ow_request.GetAccountID(payload.Body)
	if err != nil {
		return result, err
	}

	rCtx := models.RequestCtx{
		PubID:          accountID,
		ProfileID:      requestExtWrapper.ProfileId,
		DisplayID:      requestExtWrapper.VersionId,
		SSAuction:      requestExtWrapper.SSAuctionFlag,
		SummaryDisable: requestExtWrapper.SumryDisableFlag,
		LogInfoFlag:    requestExtWrapper.LogInfoFlag,
		IsCTVRequest:   models.IsCTVAPIRequest(payload.Request.URL.Path),
		UA:             payload.Request.Header.Get("User-Agent"),
		Cookies:        payload.Request.Header.Get(models.COOKIE),
		Debug:          payload.Request.Header.Get(models.Debug) == "1",
		// IsTestRequest:  payload.Request.Test == 2,
	}

	// Start------------------------------------------------------------------------------------------------------------------------
	// Move this to BeforeValidationHook where we have already unmarshaled request.
	// test, _ := ow_request.GetTest(payload.Body)
	bidRequest := &openrtb2.BidRequest{}
	err = json.Unmarshal(payload.Body, bidRequest)
	if err != nil {
		return result, fmt.Errorf("failed to decode request %v", err)
	}

	rCtx.IsTestRequest = bidRequest.Test == 2

	partnerConfigMap := m.cache.GetPartnerConfigMap(bidRequest, rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	if len(partnerConfigMap) == 0 {
		return result, errors.New("failed to get profile data")
	}
	rCtx.PartnerConfigMap = partnerConfigMap
	// End--------------------------------------------------------------------------------------------------------------------------

	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext["rctx"] = rCtx

	result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
		//NYC_TODO: convert /2.5 redirect request to auction
		rctx := result.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		ep.Body, err = m.updateORTBV25Request(rctx, payload.Body)
		return ep, err
	}, hookstage.MutationUpdate, "request-body-with-profile-data")

	return result, nil
}
