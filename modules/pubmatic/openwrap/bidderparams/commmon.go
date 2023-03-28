package bidderparams

import (
	"fmt"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/errorcodes"
)

func getSlotMeta(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) ([]string, map[string]models.SlotMapping, models.SlotMappingInfo, [][2]int64) {
	var slotMap map[string]models.SlotMapping
	var slotMappingInfo models.SlotMappingInfo

	//don't read mappings from cache in case of test=2
	if !rctx.IsTestRequest {
		slotMap = cache.GetMappingsFromCacheV25(rctx, partnerID)
		if slotMap == nil {
			return nil, nil, models.SlotMappingInfo{}, nil
		}
		slotMappingInfo = cache.GetSlotToHashValueMapFromCacheV25(rctx, partnerID)
		if len(slotMappingInfo.OrderedSlotList) == 0 {
			return nil, nil, models.SlotMappingInfo{}, nil
		}
	}

	var wh [][2]int64
	if imp.Banner != nil {
		if imp.Banner.W != nil && imp.Banner.H != nil {
			wh = append(wh, [2]int64{*imp.Banner.H, *imp.Banner.W})
		}

		for _, format := range imp.Banner.Format {
			wh = append(wh, [2]int64{format.H, format.W})
		}
	}

	if imp.Video != nil {
		wh = append(wh, [2]int64{0, 0})
	}

	kgp := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN]

	var div string
	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	var slots []string
	for _, format := range wh {
		slot := GenerateSlotName(format[0], format[1], kgp, imp.TagID, div, rctx.Source)
		if slot != "" {
			slots = append(slots, slot)
			// NYC_TODO: break at i=0 for pubmatic?
		}
	}

	// NYC_TODO wh is returned temporarily
	return slots, slotMap, slotMappingInfo, wh
}

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
func GenerateSlotName(h, w int64, kgp, tagid, div, src string) string {
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
		// fieldMap, matchingRegex = RunRegexMatch(*models.Id, slotMap, slotMappingInfo, slotKey, pubIDInt, partnerID, profileID, versionID, partnerConf[models.BidderCode])
	}

	slotMappingObj, ok := slotMap[strings.ToLower(slotName)]
	if !ok {
		return nil, errorcodes.ErrGADSMissingConfig
	}
	return slotMappingObj.SlotMappings, nil
}
