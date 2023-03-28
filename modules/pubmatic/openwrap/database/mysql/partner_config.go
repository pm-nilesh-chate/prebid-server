package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// return the list of active server side header bidding partners
// with their configurations at publisher-profile-version level
func (db *mySqlDB) GetActivePartnerConfigurations(pubId, profileId int, displayVersion int) (map[int]map[string]string, error) {
	versionID, displayVersionID, err := db.getVersionID(profileId, displayVersion, pubId)
	if err != nil {
		return nil, err
	}

	partnerConfigMap, err := db.getActivePartnerConfigurations(pubId, profileId, versionID)
	if err != nil && partnerConfigMap[-1] != nil {
		partnerConfigMap[-1][models.DisplayVersionID] = strconv.Itoa(displayVersionID)
	}
	return partnerConfigMap, err
}

func (db *mySqlDB) getActivePartnerConfigurations(pubId, profileId int, versionID int) (map[int]map[string]string, error) {
	getActivePartnersQuery := strings.Replace(getParterConfig, versionIdKey, strconv.Itoa(versionID), -1)
	getActivePartnersQuery = fmt.Sprintf(getActivePartnersQuery, db.cfg.MaxDbContextTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()
	rows, err := db.conn.QueryContext(ctx, getActivePartnersQuery)
	if err != nil {
		return nil, err
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
			continue
		}

		_, ok := partnerConfigMap[partnerID]
		//below logic will take care of overriding account level partner keys with version level partner keys
		//if key name is same for a given partnerID (Ref ticket: UOE-5647)
		if !ok {
			partnerConfigMap[partnerID] = map[string]string{models.PARTNER_ID: strconv.Itoa(partnerID)}
		}

		if testConfig == 1 {
			keyName = keyName + "_test"
			partnerConfigMap[partnerID][models.PartnerTestEnabledKey] = "1"
		}

		partnerConfigMap[partnerID][keyName] = value

		if _, ok := partnerConfigMap[partnerID][models.PREBID_PARTNER_NAME]; !ok && prebidPartnerName != "-" {
			partnerConfigMap[partnerID][models.PREBID_PARTNER_NAME] = prebidPartnerName
			partnerConfigMap[partnerID][models.BidderCode] = bidderCode
			partnerConfigMap[partnerID][models.IsAlias] = strconv.Itoa(isAlias)
		}
	}

	// NYC_TODO: ignore close error
	if err = rows.Err(); err != nil {

	}
	return partnerConfigMap, nil
}

func (db *mySqlDB) getVersionID(profileID, displayVersionID, pubID int) (int, int, error) {
	var versionID, displayVersionIDFromDB int
	var row *sql.Row

	if displayVersionID == 0 {
		row = db.conn.QueryRow(liveVersionInnerQuery, profileID, pubID)
	} else {
		row = db.conn.QueryRow(displayVersionInnerQuery, profileID, displayVersionID, pubID)
	}

	err := row.Scan(&versionID, &displayVersionIDFromDB)
	if err != nil {
		return versionID, displayVersionIDFromDB, err
	}
	return versionID, displayVersionIDFromDB, nil
}
