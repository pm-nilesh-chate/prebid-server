package openwrap

import (
	"fmt"

	"github.com/prebid/openrtb/v17/openrtb2"
)

func getIncomingSlots(imp openrtb2.Imp) []string {
	sizes := map[string]struct{}{}
	if imp.Banner != nil {
		if imp.Banner.W != nil && imp.Banner.H != nil {
			sizes[fmt.Sprintf("%dx%d", *imp.Banner.W, *imp.Banner.H)] = struct{}{}
		}

		for _, format := range imp.Banner.Format {
			sizes[fmt.Sprintf("%dx%d", format.W, format.H)] = struct{}{}
		}
	}

	if imp.Video != nil {
		sizes[fmt.Sprintf("%dx%dv", imp.Video.W, imp.Video.H)] = struct{}{}
	}

	var s []string
	for k := range sizes {
		s = append(s, k)
	}
	return s
}
