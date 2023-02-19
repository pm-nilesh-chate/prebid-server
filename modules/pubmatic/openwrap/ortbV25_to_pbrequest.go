package openwrap

import (
	"fmt"

	"github.com/prebid/openrtb/v17/openrtb2"
)

func (m Module) prepareBidderParamsJSON(pubId, profileId, versionId, partnerId int, pm map[string]string, imp openrtb2.Imp) string {
	kgp := pm["kgp"]
	isRegex := kgp == "_AU_@_DIV_@_W_x_H_"

	var wh [][2]int64
	if imp.Banner != nil {
		wh = append(wh, [2]int64{*imp.Banner.H, *imp.Banner.W})

		for _, format := range imp.Banner.Format {
			wh = append(wh, [2]int64{format.H, format.W})
		}
	}

	if imp.Video != nil {
		wh = append(wh, [2]int64{0, 0})
	}

	var slots []string
	for _, format := range wh {
		slot := generateSlotName(kgp, "imp.Ext.wrapper.div", imp.TagID, "req.ext.site.domain", "req.ext.site.page", "req.ext.app.bundle", format[0], format[1])
		if slot != "" {
			slots = append(slots, slot)
			// break at i=0 for pubmatic?
		}
	}

	if len(slots) == 0 {
		return ""
	}

	var bidderParamsJSON string
	for _, slot := range slots {
		if isRegex {
			// if notTestMode{
			// slotMappingInfo = dbcache.GetCache().GetSlotToHashValueMapFromCacheV25(request, partnerConf)
			// }
			// paramsMap := RunRegexMatch(slotMap, SlotMapping)
			// if slotMappingInfo.HashValueMap != nil && matchingRegex != "" {
			// 	hashValue = slotMappingInfo.HashValueMap[matchingRegex]
			// }
		} else {
			bidderParamsJSON, _ = m.ProfileCache[profileId][versionId][partnerId][slot]
		}
	}

	// TODO
	// bidder to only unmarshal and update the bidderParamsJSON. No need of this is we offload this work while profile creation.
	// i.e. save the bidderParamsJSON in correct json (format needed by partner) instead of flat json. Update fields, etc before adding db entry
	// bidderParamsJSON = adapters.ProcessBidderParamsJSON(bidderParamsJSON)

	return bidderParamsJSON
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
func generateSlotName(kgp, div, tagid, domain, page, appbundle string, h, w int64) string {
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
		var src string
		if domain != "" {
			src = domain
		} else if page != "" {
			src = page
		} else if appbundle != "" {
			src = appbundle
		}
		return fmt.Sprintf("%s@%s@s_VASTTAG_", tagid, src) //TODO check where/how _VASTTAG_ is updated
	default: // existing generic flow (below)
	}
	return ""
}
