package openwrap

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	ow_config "github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	ow_db "github.com/prebid/prebid-server/modules/pubmatic/openwrap/db"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func Builder(rawCfg json.RawMessage, _ moduledeps.ModuleDeps) (interface{}, error) {
	cfg := ow_config.SSHB{}

	err := json.Unmarshal(rawCfg, &cfg)
	if err != nil {
		return Module{}, fmt.Errorf("invalid openwrap config: %v", err)
	}

	db, err := ow_db.Open("mysql", cfg.OpenWrap.Database)
	if err != nil {
		return Module{}, fmt.Errorf("failed to open db connection: %v", err)
	}

	return Module{
		Config:         cfg,
		DB:             db,
		ProfileCache:   make(map[int]map[int]ProfileMapping),
		PublisherCache: make(map[int]map[string]string),
	}, nil
}

// partnerid-key-value
type ProfileMapping map[int]map[string]string

type Module struct {
	Config         ow_config.SSHB
	DB             *sql.DB
	ProfileCache   map[int]map[int]ProfileMapping // profile-version-mapping
	PublisherCache map[int]map[string]string
}

func (m Module) HandleEntrypointHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.EntrypointPayload,
) (hookstage.HookResult[hookstage.EntrypointPayload], error) {
	result := hookstage.HookResult[hookstage.EntrypointPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.EntrypointPayload]{}

	extWrapperBytes, _, _, err := jsonparser.Get(payload.Body, "ext", "wrapper")
	if err != nil {
		return result, fmt.Errorf("request.ext.wrapper not found: %v", err)
	}

	extWrapperMap := make(map[string]interface{})
	err = json.Unmarshal(extWrapperBytes, &extWrapperMap)
	if err != nil {
		return result, fmt.Errorf("failed to decode request.ext.wrapper : %v", err)
	}

	profileId, _ := extWrapperMap["profileid"].(float64) //update unmarshler to get int

	versionId := 1
	version, ok := extWrapperMap["version"].(float64)
	if ok && version != 0 {
		versionId = int(version)
	}

	pubIdStr, _, err := searchAccountId(payload.Body)
	if err != nil {
		return result, fmt.Errorf("failed to get publisher id : %v", err)
	}

	pubId, err := strconv.Atoi(pubIdStr)
	if err != nil {
		return result, fmt.Errorf("invalid publisher id : %v", err)
	}

	result.ModuleContext = make(hookstage.ModuleContext)
	result.ModuleContext["profileid"] = int(profileId)
	result.ModuleContext["displayversionid"] = int(versionId)
	result.ModuleContext["pubid"] = int(pubId)

	result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
		//TODO

		return ep, nil
	}, hookstage.MutationUpdate, "-")

	return result, nil

	// result := hookstage.HookResult[hookstage.EntrypointPayload]{}
	// result.ChangeSet = hookstage.ChangeSet[hookstage.EntrypointPayload]{}
	// result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
	// 	bidRequest := &ow_ortb.BidRequest{}
	// 	err := json.Unmarshal(ep.Body, bidRequest)
	// 	if err != nil {
	// 		return ep, fmt.Errorf("failed to unmarsha request: %v", err)
	// 	}

	// 	wtExt, ok := bidRequest.Ext.(*ow_ortb.ExtRequest)
	// 	if !ok {
	// 		return ep, fmt.Errorf("invalid ow request.ext: %v", err)
	// 	}

	// 	var displayVersion, profileId int
	// 	SSAuctionFlag := -1
	// 	SumryDisableFlag := 0
	// 	logInfoFlag := 0
	// 	unMappedSlotCnt := 0

	// 	displayVersion = *wtExt.Wrapper.VersionId
	// 	profileId = *wtExt.Wrapper.ProfileId
	// 	profileIdStr := strconv.Itoa(profileId)
	// 	pubID, err := strconv.Atoi(ow_ortb.GetPublisherID(bidRequest))
	// 	if err != nil {
	// 		return ep, fmt.Errorf("invalid publisherId: %v", err)
	// 	}

	// 	if wtExt.Wrapper.SSAuctionFlag != nil {
	// 		SSAuctionFlag = *wtExt.Wrapper.SSAuctionFlag
	// 	}
	// 	SumryDisableFlag = *wtExt.Wrapper.SumryDisableFlag
	// 	if wtExt.Wrapper.LogInfoFlag != nil {
	// 		logInfoFlag = *wtExt.Wrapper.LogInfoFlag
	// 	}

	// 	for i := 0; i < len(bidRequest.Imp); i++ {

	// 	}

	// 	return ep, err
	// }, hookstage.MutationUpdate, "requestbody")

	// result.ChangeSet = hookstage.ChangeSet[hookstage.EntrypointPayload]{}
	// result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {

	// 	extBytes, _, _, err := jsonparser.Get(ep.Body, "ext")
	// 	if err != nil {
	// 		return ep, fmt.Errorf("request.ext not found: %v", err)
	// 	}

	// 	extMap := make(map[string]interface{})
	// 	err = json.Unmarshal(extBytes, &extMap)
	// 	if err != nil {
	// 		return ep, fmt.Errorf("failed to decode request.ext : %v", err)
	// 	}

	// 	wrapperI, ok := extMap["wrapper"]
	// 	if !ok {
	// 		return ep, errors.New("request.ext.wrapper not found")
	// 	}

	// 	wrapperMap, ok := wrapperI.(map[string]interface{})
	// 	if !ok {
	// 		return ep, errors.New("request.ext.wrapper not valid")
	// 	}

	// 	profileId, _ := wrapperMap["profileid"].(float64)

	// 	versionId := 1
	// 	version, ok := wrapperMap["version"].(float64)
	// 	if ok && version != 0 {
	// 		versionId = int(version)
	// 	}

	// 	storedProcedureId := fmt.Sprintf(`"%d-%d"`, int(profileId), int(versionId))
	// 	// newPayload, err := jsonparser.Set(ep.Body, []byte(storedProcedureId), "imp", "[0]", "ext", openrtb_ext.PrebidExtKey, "storedrequest", "id")

	// 	storedProcedureId = `{"id":` + storedProcedureId + `}`
	// 	newPayload, err := jsonparser.Set(ep.Body, []byte(storedProcedureId), "imp", "[0]", "ext", "prebid", "storedrequest")
	// 	if err == nil {
	// 		ep.Body = newPayload
	// 	}

	// 	return ep, err
	// }, hookstage.MutationUpdate, "requestbody")

	// return result, nil
}

// Han rejects bids for a specific bidder if they fail the attribute check.
func (m Module) HandleBeforeValidationHook(
	_ context.Context,
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.BeforeValidationRequestPayload,
) (hookstage.HookResult[hookstage.BeforeValidationRequestPayload], error) {
	result := hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{}

	profileId := miCtx.ModuleContext["profileid"].(int)
	pubId := miCtx.ModuleContext["pubid"].(int)

	var versionID, displayVersionIDFromDB int
	row := m.DB.QueryRow(ow_db.DisplayVersionInnerQuery, profileId, miCtx.ModuleContext["displayversionid"].(int), pubId)
	err := row.Scan(&versionID, &displayVersionIDFromDB)
	if err != nil {
		return result, fmt.Errorf("failed to get profile version id: %v", err)
	}

	miCtx.ModuleContext["versionid"] = versionID

	mapping, err := m.getActivePartnerConfigurations(pubId, profileId, versionID)
	if err != nil {
		return result, fmt.Errorf("failed to get profile details: %v", err)
	}

	if len(mapping) != 0 && mapping[-1] != nil {
		mapping[-1]["displayversionid"] = strconv.Itoa(displayVersionIDFromDB)
	}

	if m.ProfileCache[profileId] == nil {
		m.ProfileCache[profileId] = make(map[int]ProfileMapping)
	}
	m.ProfileCache[profileId][versionID] = mapping

	m.PublisherCache[pubId] = m.getPublisherSlotNameHash(pubId)

	slotmapping := m.getWrapperSlotMappings(mapping, profileId, displayVersionIDFromDB)

	for partnerId, slotMappingList := range slotmapping {
		sort.Slice(slotMappingList, func(i, j int) bool {
			return slotMappingList[i].OrderID < slotMappingList[j].OrderID
		})
		for _, slotMapping := range slotMappingList {
			m.ProfileCache[profileId][versionID][partnerId][slotMapping.SlotName] = slotMapping.MappingJson
		}
	}

	miCtx.ModuleContext["profileMeta"] = m.ProfileCache[profileId][versionID]
	miCtx.ModuleContext["publisherMeta"] = m.PublisherCache[pubId]

	result.ChangeSet.AddMutation(func(parp hookstage.BeforeValidationRequestPayload) (hookstage.BeforeValidationRequestPayload, error) {
		// TODO: mov all declartion here to avoid race condition.
		// ex. pubId

		// parp.BidRequest.Site.Page = "dummy.updated.by.pubmatic.module"
		// platform := m.ProfileCache[profileId][versionID][-1]["platform"]

		pm, ok := miCtx.ModuleContext["profileMeta"].(ProfileMapping)
		if !ok {
			return parp, errors.New("invalid profile details in cache")
		}

		if cur, ok := pm[-1]["adServerCurrency"]; ok {
			parp.BidRequest.Cur = []string{cur}
		}
		// prebid timeout, etc

		for i := 0; i < len(parp.BidRequest.Imp); i++ {
			bidderParams := make(map[string]json.RawMessage)
			for _, p := range pm {
				if p["serverSideEnabled"] == "1" {
					partnerId, err := strconv.Atoi(p["partnerid"])
					if err != nil && partnerId > 0 {
						continue
					}

					bidderCode := p["bidder"]

					slotMappingJSON := m.prepareBidderParamsJSON(pubId, profileId, versionID, partnerId, p, parp.BidRequest.Imp[i])

					if bidderCode == string(openrtb_ext.BidderPubmatic) {
						slotMappingJSON = slotMappingJSON[:len(slotMappingJSON)-1] + `,"publisherId":"` + fmt.Sprintf("%d", pubId) + `"}`
					}

					bidderParams[bidderCode] = json.RawMessage(slotMappingJSON)
				}
			}

			if len(bidderParams) != 0 {
				impExt := make(map[string]json.RawMessage)
				_ = json.Unmarshal(parp.BidRequest.Imp[i].Ext, &impExt)

				var prebid openrtb_ext.ExtImpPrebid
				if _, ok := impExt["prebid"]; ok {
					_ = json.Unmarshal(impExt["prebid"], &prebid)
				}
				prebid.Bidder = bidderParams
				impExt["prebid"], _ = json.Marshal(prebid)

				newImpExt, err := json.Marshal(impExt)
				if err != nil {
					fmt.Println("error creating impExt", bidderParams)
				}
				parp.BidRequest.Imp[i].Ext = newImpExt
			}
		}

		return parp, nil
	}, hookstage.MutationUpdate, "request.site.page")

	return result, nil
}
