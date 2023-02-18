package openwrap

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	ow_db "github.com/prebid/prebid-server/modules/pubmatic/openwrap/db"
)

/*SlotMapping object contains information for a given slot*/
type SlotMapping struct {
	PartnerId   int64
	AdapterId   int64
	VersionId   int64
	SlotName    string
	MappingJson string
	Hash        string
	OrderID     int64
}

func (m Module) getActivePartnerConfigurations(pubId, profileId int, versionID int) (map[int]map[string]string, error) {
	getActivePartnersQuery := strings.Replace(ow_db.GetParterConfig, ow_db.VersionIdKey, strconv.Itoa(versionID), -1)
	getActivePartnersQuery = fmt.Sprintf(getActivePartnersQuery, m.Config.OpenWrap.Timeout.MaxQueryTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(m.Config.OpenWrap.Database.MaxDbContextTimeout)))
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, getActivePartnersQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner mapping details: %v", err)
	}
	defer rows.Close()

	partnerConfigMap := make(map[int]map[string]string, 0)
	for rows.Next() {
		var partnerID int
		var keyName string
		var value string
		var prebidPartnerName, bidderCode string
		var entityTypeID, testConfig, isAlias int
		if err := rows.Scan(&partnerID, &prebidPartnerName, &bidderCode, &isAlias, &entityTypeID, &testConfig, &keyName, &value); err != nil {
			return nil, fmt.Errorf("failed to read partner mapping details: %v", err)
		}

		_, ok := partnerConfigMap[partnerID]
		//below logic will take care of overriding account level partner keys with version level partner keys
		//if key name is same for a given partnerID (Ref ticket: UOE-5647)
		if !ok {
			partnerConfigMap[partnerID] = map[string]string{"partnerid": strconv.Itoa(partnerID)}
		}

		if testConfig == 1 {
			keyName = keyName + "_test"
			partnerConfigMap[partnerID]["testenabled"] = "1"
		}

		partnerConfigMap[partnerID][keyName] = value

		if _, ok := partnerConfigMap[partnerID]["prebidname"]; !ok && prebidPartnerName != "-" {
			partnerConfigMap[partnerID]["partnername"] = prebidPartnerName
			partnerConfigMap[partnerID]["bidder"] = bidderCode
			partnerConfigMap[partnerID]["isalias"] = strconv.Itoa(isAlias)
		}
	}

	return partnerConfigMap, nil
}

func (m Module) getWrapperSlotMappings(partnerConfigMap map[int]map[string]string, profileId, displayVersion int) map[int][]SlotMapping {
	partnerSlotMappingMap := make(map[int][]SlotMapping)

	var query string
	query = formWrapperSlotMappingQuery(profileId, displayVersion, partnerConfigMap)

	rows, err := m.DB.Query(query)
	if err != nil {
		return partnerSlotMappingMap
	}
	defer rows.Close()

	for rows.Next() {
		var slotMapping = SlotMapping{}
		err := rows.Scan(&slotMapping.PartnerId, &slotMapping.AdapterId, &slotMapping.VersionId, &slotMapping.SlotName, &slotMapping.MappingJson, &slotMapping.OrderID)
		if nil != err {
			continue
		}

		slotMappingList, found := partnerSlotMappingMap[int(slotMapping.PartnerId)]
		if found {
			slotMappingList = append(slotMappingList, slotMapping)
			partnerSlotMappingMap[int(slotMapping.PartnerId)] = slotMappingList
		} else {
			newSlotMappingList := make([]SlotMapping, 0)
			newSlotMappingList = append(newSlotMappingList, slotMapping)
			partnerSlotMappingMap[int(slotMapping.PartnerId)] = newSlotMappingList
		}

	}
	//vastTagHookPartnerSlotMapping(partnerSlotMappingMap, profileId, displayVersion)
	return partnerSlotMappingMap
}

func formWrapperSlotMappingQuery(profileID, displayVersion int, partnerConfigMap map[int]map[string]string) string {
	var query string
	var partnerIDStr string
	for partnerID := range partnerConfigMap {
		partnerIDStr = partnerIDStr + strconv.Itoa(partnerID) + ","
	}
	partnerIDStr = strings.TrimSuffix(partnerIDStr, ",")

	if displayVersion != 0 {
		query = strings.Replace(ow_db.GetWrapperSlotMappingsQuery, ow_db.ProfileIdKey, strconv.Itoa(profileID), -1)
		query = strings.Replace(query, ow_db.DisplayVersionKey, strconv.Itoa(displayVersion), -1)
		query = strings.Replace(query, ow_db.PartnerIdKey, partnerIDStr, -1)
	} else {
		query = strings.Replace(ow_db.GetWrapperLiveVersionSlotMappings, ow_db.ProfileIdKey, strconv.Itoa(profileID), -1)
		query = strings.Replace(query, ow_db.PartnerIdKey, partnerIDStr, -1)
	}
	return query
}

func (m Module) getPublisherSlotNameHash(pubID int) map[string]string {
	nameHashMap := make(map[string]string)

	query := strings.Replace(ow_db.GetSlotNameHash, ow_db.PubIdKey, strconv.Itoa(pubID), -1)
	rows, err := m.DB.Query(query)
	if err != nil {
		return nameHashMap
	}
	defer rows.Close()

	for rows.Next() {
		var name, hash string
		if err = rows.Scan(&name, &hash); err != nil {
			continue
		}
		nameHashMap[name] = hash
	}

	return nameHashMap
}
