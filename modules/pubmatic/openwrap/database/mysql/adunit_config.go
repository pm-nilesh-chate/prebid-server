package mysql

import (
	"fmt"
	"strconv"
	"strings"
)

// GetAdunitConfig - Method to get adunit config for a given profile and display version from giym DB
func (db *mySqlDB) GetAdunitConfig(profileID, displayVersionID int) (string, error) {
	var adunitConfigJSON string
	adunitConfigQuery := ""
	if displayVersionID == 0 {
		adunitConfigQuery = getAdunitConfigForLiveVersion
	} else {
		adunitConfigQuery = getAdunitConfigQuery
	}
	adunitConfigQuery = strings.Replace(adunitConfigQuery, profileIdKey, strconv.Itoa(profileID), -1)
	adunitConfigQuery = strings.Replace(adunitConfigQuery, displayVersionKey, strconv.Itoa(displayVersionID), -1)
	err := db.conn.QueryRow(adunitConfigQuery).Scan(&adunitConfigJSON)
	if err != nil {
		err = fmt.Errorf("[QUERY_FAILED] Name:[%v] Error:[%v]", "GetAdunitConfig", err.Error())
		return "", err
	}
	return adunitConfigJSON, nil
}
