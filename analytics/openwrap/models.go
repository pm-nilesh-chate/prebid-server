package openwrap

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

const (
	//constant for adformat
	Banner = "banner"
	Video  = "video"
	Native = "native"

	REVSHARE      = "rev_share"
	BID_PRECISION = 2
)

func GetSize(width, height int64) string {
	return fmt.Sprintf("%dx%d", width, height)
}

// CreatePartnerKey returns key with partner appended
func CreatePartnerKey(partner, key string) string {
	if partner == "" {
		return key
	}
	return key + "_" + partner
}

// GetAdFormat gets adformat from creative(adm) of the bid
func GetAdFormat(adm string) string {
	adFormat := Banner
	videoRegex, _ := regexp.Compile("<VAST\\s+")

	if videoRegex.MatchString(adm) {
		adFormat = Video
	} else {
		var admJSON map[string]interface{}
		err := json.Unmarshal([]byte(strings.Replace(adm, "/\\/g", "", -1)), &admJSON)
		if err == nil && admJSON != nil && admJSON["native"] != nil {
			adFormat = Native
		}
	}
	return adFormat
}

func GetRevenueShare(partnerConfig map[string]string) float64 {
	var revShare float64

	if val, ok := partnerConfig[REVSHARE]; ok {
		revShare, _ = strconv.ParseFloat(val, 64)
	}
	return revShare
}

func GetNetEcpm(price float64, revShare float64) float64 {
	if revShare == 0 {
		return toFixed(price, BID_PRECISION)
	}
	price = price * (1 - revShare/100)
	return toFixed(price, BID_PRECISION)
}

func GetGrossEcpm(price float64) float64 {
	return toFixed(price, BID_PRECISION)
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ExtractDomain(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return u.Host, nil
}

func getSlotName(impID string, tagID string) string {
	return fmt.Sprintf("%s_%s", impID, tagID)
}

func getSizesFromImp(imp openrtb2.Imp, platform string) []string {
	//get unique sizes from banner.format and banner.w and banner.h
	sizes := make(map[string]bool)
	var sizeArr []string
	// TODO: handle video
	if imp.Banner != nil && imp.Banner.W != nil && imp.Banner.H != nil {
		size := getSizeForPlatform(*imp.Banner.W, *imp.Banner.H, platform)
		if _, ok := sizes[size]; !ok {
			sizeArr = append(sizeArr, size)
			sizes[size] = true
		}
	}

	if imp.Banner != nil && imp.Banner.Format != nil && len(imp.Banner.Format) != 0 {
		for _, eachFormat := range imp.Banner.Format {
			size := GetSize(eachFormat.W, eachFormat.H)
			if _, ok := sizes[size]; !ok {
				sizeArr = append(sizeArr, size)
				sizes[size] = true
			}
		}
	}

	if imp.Video != nil {
		size := getSizeForPlatform(imp.Video.W, imp.Video.H, models.PLATFORM_VIDEO)
		if _, ok := sizes[size]; !ok {
			sizeArr = append(sizeArr, size)
			sizes[size] = true
		}
	}
	return sizeArr
}

func getSizeByPlatform(incomingSlots [][2]int64, platform string) []string {
	sizes := map[string]struct{}{}

	for _, slot := range incomingSlots {
		if platform == models.PLATFORM_VIDEO {
			sizes[fmt.Sprintf("%dx%dv", slot[1], slot[0])] = struct{}{}
		}
		sizes[fmt.Sprintf("%dx%d", slot[1], slot[0])] = struct{}{}
	}

	var s []string
	for k := range sizes {
		s = append(s, k)
	}
	return s
}
