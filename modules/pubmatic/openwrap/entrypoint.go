package openwrap

import (
	"context"
	"encoding/json"
	"fmt"

	"errors"

	"github.com/prebid/openrtb/v17/openrtb2"
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

// // Han rejects bids for a specific bidder if they fail the attribute check.
// func (m OpenWrap) HandleBeforeValidationHook(
// 	_ context.Context,
// 	miCtx hookstage.ModuleInvocationContext,
// 	payload hookstage.BeforeValidationRequestPayload,
// ) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
// 	result := hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{}
// 	result.ChangeSet = hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{}

// 	profileId := miCtx.ModuleContext["profileid"].(int)
// 	pubId := miCtx.ModuleContext["pubid"].(int)

// 	var versionID, displayVersionIDFromDB int
// 	row := m.DB.QueryRow(ow_db.DisplayVersionInnerQuery, profileId, miCtx.ModuleContext["displayversionid"].(int), pubId)
// 	err := row.Scan(&versionID, &displayVersionIDFromDB)
// 	if err != nil {
// 		return result, fmt.Errorf("failed to get profile version id: %v", err)
// 	}

// 	miCtx.ModuleContext["versionid"] = versionID

// 	mapping, err := m.getActivePartnerConfigurations(pubId, profileId, versionID)
// 	if err != nil {
// 		return result, fmt.Errorf("failed to get profile details: %v", err)
// 	}

// 	if len(mapping) != 0 && mapping[-1] != nil {
// 		mapping[-1]["displayversionid"] = strconv.Itoa(displayVersionIDFromDB)
// 	}

// 	if m.ProfileCache[profileId] == nil {
// 		m.ProfileCache[profileId] = make(map[int]ProfileMapping)
// 	}
// 	m.ProfileCache[profileId][versionID] = mapping

// 	m.PublisherCache[pubId] = m.getPublisherSlotNameHash(pubId)

// 	slotmapping := m.getWrapperSlotMappings(mapping, profileId, displayVersionIDFromDB)

// 	for partnerId, slotMappingList := range slotmapping {
// 		sort.Slice(slotMappingList, func(i, j int) bool {
// 			return slotMappingList[i].OrderID < slotMappingList[j].OrderID
// 		})
// 		for _, slotMapping := range slotMappingList {
// 			m.ProfileCache[profileId][versionID][partnerId][slotMapping.SlotName] = slotMapping.MappingJson
// 		}
// 	}

// 	miCtx.ModuleContext["profileMeta"] = m.ProfileCache[profileId][versionID]
// 	miCtx.ModuleContext["publisherMeta"] = m.PublisherCache[pubId]

// 	result.ChangeSet.AddMutation(func(parp hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
// 		// TODO: mov all declartion here to avoid race condition.
// 		// ex. pubId

// 		// parp.BidRequest.Site.Page = "dummy.updated.by.pubmatic.module"
// 		// platform := m.ProfileCache[profileId][versionID][-1]["platform"]

// 		pm, ok := miCtx.ModuleContext["profileMeta"].(ProfileMapping)
// 		if !ok {
// 			return parp, errors.New("invalid profile details in cache")
// 		}

// 		if cur, ok := pm[-1]["adServerCurrency"]; ok {
// 			parp.BidRequest.Cur = []string{cur}
// 		}
// 		// prebid timeout, etc

// 		for i := 0; i < len(parp.BidRequest.Imp); i++ {
// 			bidderParams := make(map[string]json.RawMessage)
// 			for _, p := range pm {
// 				if p["serverSideEnabled"] == "1" {
// 					partnerId, err := strconv.Atoi(p["partnerid"])
// 					if err != nil && partnerId > 0 {
// 						continue
// 					}

// 					bidderCode := p["bidder"]

// 					slotMappingJSON := m.prepareBidderParamsJSON(pubId, profileId, versionID, partnerId, p, parp.BidRequest.Imp[i])

// 					if bidderCode == string(openrtb_ext.BidderPubmatic) {
// 						slotMappingJSON = slotMappingJSON[:len(slotMappingJSON)-1] + `,"publisherId":"` + fmt.Sprintf("%d", pubId) + `"}`
// 					}

// 					bidderParams[bidderCode] = json.RawMessage(slotMappingJSON)
// 				}
// 			}

// 			if len(bidderParams) != 0 {
// 				impExt := make(map[string]json.RawMessage)
// 				_ = json.Unmarshal(parp.BidRequest.Imp[i].Ext, &impExt)

// 				var prebid openrtb_ext.ExtImpPrebid
// 				if _, ok := impExt["prebid"]; ok {
// 					_ = json.Unmarshal(impExt["prebid"], &prebid)
// 				}
// 				prebid.Bidder = bidderParams
// 				impExt["prebid"], _ = json.Marshal(prebid)

// 				newImpExt, err := json.Marshal(impExt)
// 				if err != nil {
// 					fmt.Println("error creating impExt", bidderParams)
// 				}
// 				parp.BidRequest.Imp[i].Ext = newImpExt
// 			}
// 		}

// 		return parp, nil
// 	}, hookstage.MutationUpdate, "request.site.page")

// 	return result, nil
// }
