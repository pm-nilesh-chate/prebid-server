package mysql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// GetPublisherVASTTags - Method to get vast tags associated with publisher id from giym DB
func (db *mySqlDB) GetPublisherVASTTags(pubID int) (models.PublisherVASTTags, error) {

	/*
		//TOOD:VIRAL Remove Hook once UI/API changes are in place
		if out := vastTagHookPublisherVASTTags(rtbReqId, pubID); nil != out {
			return out, nil
		}
	*/

	getActiveVASTTagsQuery := strings.Replace(getPublisherVASTTagsQuery, pubIdKey, strconv.Itoa(pubID), -1)

	rows, err := db.conn.Query(getActiveVASTTagsQuery)
	if err != nil {
		err = fmt.Errorf("[QUERY_FAILED] Name:[%v] Error:[%v]", "GetPublisherVASTTags", err.Error())
		return nil, err
	}
	defer rows.Close()

	vasttags := models.PublisherVASTTags{}
	for rows.Next() {
		var vastTag models.VASTTag
		if err := rows.Scan(&vastTag.ID, &vastTag.PartnerID, &vastTag.URL, &vastTag.Duration, &vastTag.Price); err != nil {
			continue
		}
		vasttags[vastTag.ID] = &vastTag
	}
	return vasttags, nil
}
