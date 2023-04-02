package openwrap

import (
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// CheckABTestEnabled checks whether a given request is AB test enabled or not
func CheckABTestEnabled(rctx models.RequestCtx) bool {

	abTestEnabled := models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.AbTestEnabled)

	//for 2.5 display platform, check the abtest config flag in request.ext
	if rctx.Platform == models.PLATFORM_DISPLAY {
		if rctx.ABTestConfig == 1 && abTestEnabled == "1" {
			return true
		}
		return false
	}

	//for pure server side platforms, check the abTestEnabled entry in DB
	if abTestEnabled == "1" {
		return true
	}

	return false
}

// ABTestProcessing function checks if test config should be applied and change the partner config accordingly
func ABTestProcessing(rctx models.RequestCtx) (map[int]map[string]string, bool) {
	//test config logic
	if CheckABTestEnabled(rctx) {
		if rctx.Platform == models.PLATFORM_DISPLAY || ApplyTestConfig(rctx) {
			return UpdateTestConfig(rctx), true
		}
	}
	return nil, false
}

// ApplyTestConfig checks if test config should be applied
func ApplyTestConfig(rctx models.RequestCtx) bool {
	testGroupSize, err := strconv.Atoi(models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, AppendTest(models.TestGroupSize)))
	if err != nil || testGroupSize == 0 {
		return false
	}

	randomNumber := GetRandomNumberIn1To100()
	return randomNumber <= testGroupSize
}

// AppendTest appends "_test" string to given key
func AppendTest(key string) string {
	return key + test
}

// UpdateTestConfig returns the updated partnerconfig according to the test type
func UpdateTestConfig(rctx models.RequestCtx) map[int]map[string]string {

	//create copy of the map
	newPartnerConfig := copyPartnerConfigMap(rctx.PartnerConfigMap)

	//read test type
	testType := models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, AppendTest(models.TestType))

	//change partnerconfig based on test type
	switch testType {
	case models.TestTypeAuctionTimeout:
		replaceControlConfig(newPartnerConfig, models.VersionLevelConfigID, models.SSTimeoutKey)
	case models.TestTypePartners:
		//check the partner config map for test partners
		for partnerID, config := range rctx.PartnerConfigMap {
			if partnerID == models.VersionLevelConfigID {
				continue
			}

			//if current partner is test enabled, update the config with test config
			//otherwise if its a control partner, then remove it from final partner config map
			if config[models.PartnerTestEnabledKey] == "1" {
				for key := range config {
					copyTestConfig(newPartnerConfig, partnerID, key)
				}

			} else {
				delete(newPartnerConfig, partnerID)
			}
		}
		break

	case models.TestTypeClientVsServerPath:
		for partnerID, _ := range rctx.PartnerConfigMap {
			if partnerID == models.VersionLevelConfigID {
				continue
			}

			//update the "serverSideEnabled" value with test config
			replaceControlConfig(newPartnerConfig, partnerID, models.SERVER_SIDE_FLAG)

		}
		break
	default:
	}

	return newPartnerConfig
}

// copyPartnerConfigMap creates a copy of given partner config map
func copyPartnerConfigMap(m map[int]map[string]string) map[int]map[string]string {
	cp := make(map[int]map[string]string)
	for pid, conf := range m {
		config := make(map[string]string)
		for key, val := range conf {
			config[key] = val
		}
		cp[pid] = config
	}
	return cp
}

// replaceControlConfig replace control config with test config for a given key
func replaceControlConfig(partnerConfig map[int]map[string]string, partnerID int, key string) {
	if testValue := partnerConfig[partnerID][AppendTest(key)]; testValue != "" {
		partnerConfig[partnerID][key] = testValue
	}

}

// copyTestConfig checks if the given key is test config, if yes it copies it in control config
func copyTestConfig(partnerConfig map[int]map[string]string, partnerID int, key string) {
	//if the current key is test config
	if strings.HasSuffix(key, test) {
		if testValue := partnerConfig[partnerID][key]; testValue != "" {
			//get control key for the given test key to copy data
			controlKey := strings.TrimSuffix(key, test)
			partnerConfig[partnerID][controlKey] = testValue
		}
	}
}
