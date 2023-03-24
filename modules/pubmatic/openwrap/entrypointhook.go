package openwrap

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pbsOpenrtb2 "github.com/prebid/prebid-server/endpoints/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	uuid "github.com/satori/go.uuid"
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

	requestExtWrapper, err := models.GetWrapperExt(payload.Body)
	if err != nil {
		return result, err
	}

	pubid := 0
	accountID, _, err := pbsOpenrtb2.SearchAccountId(payload.Body)
	if err != nil {
		return result, fmt.Errorf("failed to get publisher id : %v", err)
	}

	pubid, err = strconv.Atoi(accountID)
	if err != nil {
		return result, fmt.Errorf("invalid publisher id : %v", err)
	}

	queryParams := payload.Request.URL.Query()

	rCtx := models.RequestCtx{
		PubID:              pubid,
		ProfileID:          requestExtWrapper.ProfileId,
		DisplayID:          requestExtWrapper.VersionId,
		LogInfoFlag:        requestExtWrapper.LogInfoFlag,
		PreferDeals:        requestExtWrapper.SupportDeals,
		SSAuction:          requestExtWrapper.SSAuctionFlag,
		SummaryDisable:     requestExtWrapper.SumryDisableFlag,
		LoggerImpressionID: requestExtWrapper.LoggerImpressionID,
		ClientConfigFlag:   requestExtWrapper.ClientConfigFlag,
		SSAI:               requestExtWrapper.SSAI,
		Aliases:            make(map[string]string),
		IsCTVRequest:       models.IsCTVAPIRequest(payload.Request.URL.Path),
		UA:                 payload.Request.Header.Get("User-Agent"),
		Cookies:            payload.Request.Header.Get(models.COOKIE),
		Debug:              queryParams.Get(models.Debug) == "1",
		StartTime:          time.Now().Unix(),
		ImpBidCtx:          make(map[string]models.ImpCtx),
	}

	if rCtx.LoggerImpressionID == "" {
		rCtx.LoggerImpressionID = uuid.NewV4().String()
	}

	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext["rctx"] = rCtx

	return result, nil
}
