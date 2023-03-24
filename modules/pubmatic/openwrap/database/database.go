package database

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

type Database interface {
	GetAdunitConfig(profileID, displayVersionID int) (*adunitconfig.AdUnitConfig, error)
	GetActivePartnerConfigurations(pubId, profileId, displayVersion int) map[int]map[string]string
	GetPubmaticSlotMappings(pubId int) map[string]models.SlotMapping
	GetPublisherSlotNameHash(pubID int) map[string]string
	GetWrapperSlotMappings(partnerConfigMap map[int]map[string]string, profileId, displayVersion int) map[int][]models.SlotMapping
	GetPublisherVASTTags(pubID int) (models.PublisherVASTTags, error)
	GetMappings(slotKey string, slotMap map[string]models.SlotMapping) (map[string]interface{}, error)
}
