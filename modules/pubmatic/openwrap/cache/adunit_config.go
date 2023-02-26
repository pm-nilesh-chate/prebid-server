package cache

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func (c *cache) populateCacheWithAdunitConfig(pubID int, profileID, displayVersion int) {

	adunitConfigJSON, err := c.db.GetAdunitConfig(profileID, displayVersion)
	if err != nil {
		//Error returned during DB access, Do not populate cache with Ad-Unit-Config
		return
	}

	adUnitConfig := make(models.AdUnitConfig, 0)

	if adunitConfigJSON != "" {

		configPatternVal, _, _, _ := jsonparser.Get([]byte(adunitConfigJSON), models.AdunitConfigConfigPatternKey)
		if configPatternVal != nil {
			adUnitConfig[models.AdunitConfigConfigPatternKey] = string(configPatternVal)
		} else {
			//Default configPattern value is "_AU_" if not present in db config
			adUnitConfig[models.AdunitConfigConfigPatternKey] = models.MACRO_AD_UNIT_ID
		}
		//checking for regex attribute strict boolean check
		isRegex, err := jsonparser.GetBoolean([]byte(adunitConfigJSON), models.AdunitConfigRegex)
		if isRegex && err == nil {
			adUnitConfig[models.AdunitConfigRegex] = isRegex
		}
		configVal, _, _, _ := jsonparser.Get([]byte(adunitConfigJSON), models.AdunitConfigConfigKey)

		if configVal != nil {
			jsonparser.ObjectEach(configVal, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
				data := make(map[string]interface{})
				if err := json.Unmarshal(value, &data); err == nil {
					if floorVal, ok := data[models.AdunitConfigFloorJSON]; ok {
						floorBytes, err := json.Marshal(floorVal)
						if err == nil {
							floor := openrtb_ext.PriceFloorRules{}
							if err := json.Unmarshal(floorBytes, &floor); err == nil {
								data[models.AdunitConfigFloorJSON] = floor
							}
						}
					}
					adUnitConfig[strings.ToLower(string(key))] = data
				}
				return nil
			})
		}
	}

	cacheKey := key(PubAdunitConfig, pubID, profileID, displayVersion)
	c.cache.Set(cacheKey, adUnitConfig, time.Duration(c.cfg.CacheDefaultExpiry))
}

// GetAdunitConfigFromCache this function gets adunit config from cache for a given request
func (c *cache) GetAdunitConfigFromCache(request *openrtb2.BidRequest, pubID int, profileID, displayVersion int) models.AdUnitConfig {
	if request.Test == 2 {
		return nil
	}

	cacheKey := key(PubAdunitConfig, pubID, profileID, displayVersion)
	if obj, ok := c.cache.Get(cacheKey); ok {
		return obj.(models.AdUnitConfig)
	}

	return nil
}
