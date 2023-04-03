package bidderparams

import (
	"fmt"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

var ignoreKeys = map[string]bool{
	models.PARTNER_ACCOUNT_NAME: true,
	models.ADAPTER_NAME:         true,
	models.ADAPTER_ID:           true,
	models.TIMEOUT:              true,
	models.KEY_GEN_PATTERN:      true,
	models.PREBID_PARTNER_NAME:  true,
	models.PROTOCOL:             true,
	models.SERVER_SIDE_FLAG:     true,
	models.LEVEL:                true,
	models.PARTNER_ID:           true,
	models.REVSHARE:             true,
	models.THROTTLE:             true,
	models.BidderCode:           true,
	models.IsAlias:              true,
}

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

	var hw [][2]int64
	if imp.Banner != nil {
		if imp.Banner.W != nil && imp.Banner.H != nil {
			hw = append(hw, [2]int64{*imp.Banner.H, *imp.Banner.W})
		}

		for _, format := range imp.Banner.Format {
			hw = append(hw, [2]int64{format.H, format.W})
		}
	}

	if imp.Video != nil {
		hw = append(hw, [2]int64{0, 0})
	}

	kgp := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN]

	var div string
	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	var slots []string
	for _, format := range hw {
		// TODO fix the param sequence. make it consistent. HxW
		slot := GenerateSlotName(format[0], format[1], kgp, imp.TagID, div, rctx.Source)
		if slot != "" {
			slots = append(slots, slot)
			// NYC_TODO: break at i=0 for pubmatic?
		}
	}

	// NYC_TODO wh is returned temporarily
	return slots, slotMap, slotMappingInfo, hw
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
	case "_AU_": // adunitconfig
		return tagid
	case "_DIV_":
		return div
	case "_AU_@_W_x_H_":
		return fmt.Sprintf("%s@%dx%d", tagid, w, h)
	case "_DIV_@_W_x_H_":
		return fmt.Sprintf("%s@%dx%d", div, w, h)
	case "_W_x_H_@_W_x_H_":
		return fmt.Sprintf("%dx%d@%dx%d", w, h, w, h)
	case "_AU_@_DIV_@_W_x_H_":
		if div == "" {
			return fmt.Sprintf("%s@%s@s%dx%d", tagid, div, w, h)
		}
		return fmt.Sprintf("%s@%s@s%dx%d", tagid, div, w, h)
	case "_AU_@_SRC_@_VASTTAG_":
		return fmt.Sprintf("%s@%s@s_VASTTAG_", tagid, src) //TODO check where/how _VASTTAG_ is updated
	default:
		// TODO: check if we need to fallback to old generic flow (below)
		// Add this cases in a map and read it from yaml file
	}
	return ""
}

/*
formSlotForDefaultMapping: In this method, we are removing wxh from the kgp because
pubmatic adapter sets wxh that we send in imp.ext.pubmatic.adslot as primary size while calling translator.
In case of default mappings, since all sizes are unmapped, we don't want to treat any size as primary
thats why we are removing size from kgp
*/
func getDefaultMappingKGP(keyGenPattern string) string {
	if strings.Contains(keyGenPattern, "@_W_x_H_") {
		return strings.ReplaceAll(keyGenPattern, "@_W_x_H_", "")
	}
	return keyGenPattern
}
