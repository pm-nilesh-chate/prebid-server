package openwrap

import "github.com/prebid/openrtb/v17/openrtb2"

func getIncomingSlots(imp openrtb2.Imp) [][2]int64 {
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
		hw = append(hw, [2]int64{imp.Video.H, imp.Video.W})
	}

	return hw
}
