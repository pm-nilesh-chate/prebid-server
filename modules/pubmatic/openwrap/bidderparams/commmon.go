package bidderparams

import (
	"fmt"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
)

// Harcode would be the optimal. We could make it configurable like _AU_@_W_x_H_:%s@%dx%d entries in pbs.yaml
// mysql> SELECT DISTINCT key_gen_pattern FROM wrapper_mapping_template;
// +----------------------+
// | key_gen_pattern      |
// +----------------------+
// | _AU_@_W_x_H_         |
// | _DIV_@_W_x_H_        |
// | _W_x_H_@_W_x_H_      |
// | _DIV_                |
// | _AU_@_DIV_@_W_x_H_   |
// | _AU_@_SRC_@_VASTTAG_ |
// +----------------------+
// 6 rows in set (0.21 sec)
func generateSlotName(h, w int64, kgp, tagid, div, src string) string {
	// func (H, W, Div), no need to validate, will always be non-nil
	switch kgp {
	case "_AU_@_W_x_H_":
		return fmt.Sprintf("%s@%dx%d", tagid, w, h)
	case "_DIV_@_W_x_H_":
		return fmt.Sprintf("%s@%dx%d", div, w, h)
	case "_W_x_H_@_W_x_H_":
		return fmt.Sprintf("%dx%d@%dx%d", w, h, w, h)
	case "_DIV_":
		return div
	case "_AU_@_DIV_@_W_x_H_":
		if div == "" {
			return fmt.Sprintf("%s@%s@s%dx%d", tagid, div, w, h)
		}
		return fmt.Sprintf("%s@%s@s%dx%d", tagid, div, w, h)
	case "_AU_@_SRC_@_VASTTAG_":
		return fmt.Sprintf("%s@%s@s_VASTTAG_", tagid, src) //TODO check where/how _VASTTAG_ is updated
	default: // existing generic flow (below)
	}
	return ""
}

func CheckSlotName(slotName string, isRegex bool, slotMap map[string]models.SlotMapping) (map[string]interface{}, error) {
	if isRegex {
		// fieldMap, matchingRegex = RunRegexMatch(*request.Id, slotMap, slotMappingInfo, slotKey, pubIDInt, partnerID, profileID, versionID, partnerConf[models.BidderCode])
	}

	slotMappingObj, ok := slotMap[strings.ToLower(slotName)]
	if !ok {
		return nil, errorcodes.ErrGADSMissingConfig
	}
	return slotMappingObj.SlotMappings, nil
}
