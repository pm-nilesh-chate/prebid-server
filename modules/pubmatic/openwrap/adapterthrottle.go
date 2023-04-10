package openwrap

import (
	"errors"
	"math/rand"
	"strconv"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// GetAdapterThrottleMap creates map of adapter and bool value which tells whether the adapter should be throtled or not
func GetAdapterThrottleMap(partnerConfigMap map[int]map[string]string) (map[string]struct{}, error) {
	adapterThrottleMap := make(map[string]struct{})
	allPartnersThrottledFlag := true
	for _, partnerConfig := range partnerConfigMap {
		if "" == partnerConfig[models.SERVER_SIDE_FLAG] || partnerConfig[models.SERVER_SIDE_FLAG] == "0" {
			continue
		}
		if ThrottleAdapter(partnerConfig) {
			adapterThrottleMap[partnerConfig[models.BidderCode]] = struct{}{}
		} else if allPartnersThrottledFlag {
			allPartnersThrottledFlag = false
		}
	}
	if allPartnersThrottledFlag {
		return adapterThrottleMap, errors.New("ErrAllPartnerThrottled")
	}
	return adapterThrottleMap, nil
}

// ThrottleAdapter this function returns bool value for whether a adapter should be throttled or not
func ThrottleAdapter(partnerConfig map[string]string) bool {
	if partnerConfig[models.THROTTLE] == "100" || partnerConfig[models.THROTTLE] == "" {
		return false
	}

	if partnerConfig[models.THROTTLE] == "0" {
		return true
	}

	//else check throttle value based on random no
	throttle, _ := strconv.ParseFloat(partnerConfig[models.THROTTLE], 64)
	throttle = 100 - throttle

	randomNumberBelow100 := GetRandomNumberBelow100()
	return !(float64(randomNumberBelow100) >= throttle)
}

var GetRandomNumberBelow100 = func() int {
	return rand.Intn(99)
}

var GetRandomNumberIn1To100 = func() int {
	return rand.Intn(100) + 1
}
