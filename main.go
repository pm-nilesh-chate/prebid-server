package prebidServer

import (
	"flag"
	"math/rand"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/usersync"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/router"
	"github.com/prebid/prebid-server/util/task"

	"github.com/spf13/viper"
)

func InitPrebidServer(configFile string) {
	rand.Seed(time.Now().UnixNano())

	flag.Parse() // required for glog flags and testing package flags

	bidderInfoPath, err := filepath.Abs(infoDirectory)
	if err != nil {
		glog.Exitf("Unable to build configuration directory path: %v", err)
	}

	bidderInfos, err := config.LoadBidderInfo(bidderInfoPath)
	if err != nil {
		glog.Exitf("Unable to load bidder configurations: %v", err)
	}
	cfg, err := loadConfig(bidderInfos)
	if err != nil {
		glog.Exitf("Configuration could not be loaded or did not pass validation: %v", err)
	}

	// Create a soft memory limit on the total amount of memory that PBS uses to tune the behavior
	// of the Go garbage collector. In summary, `cfg.GarbageCollectorThreshold` serves as a fixed cost
	// of memory that is going to be held garbage before a garbage collection cycle is triggered.
	// This amount of virtual memory won’t translate into physical memory allocation unless we attempt
	// to read or write to the slice below, which PBS will not do.
	garbageCollectionThreshold := make([]byte, cfg.GarbageCollectorThreshold)
	defer runtime.KeepAlive(garbageCollectionThreshold)

	err = serve(cfg)
	if err != nil {
		glog.Exitf("prebid-server failed: %v", err)
	}
}

const configFileName = "pbs"
const infoDirectory = "./static/bidder-info"

func loadConfig(bidderInfos config.BidderInfos) (*config.Configuration, error) {
	v := viper.New()
	config.SetupViper(v, configFileName, bidderInfos)
	return config.New(v, bidderInfos)
}

func serve(cfg *config.Configuration) error {
	fetchingInterval := time.Duration(cfg.CurrencyConverter.FetchIntervalSeconds) * time.Second
	staleRatesThreshold := time.Duration(cfg.CurrencyConverter.StaleRatesSeconds) * time.Second
	currencyConverter := currency.NewRateConverter(&http.Client{}, cfg.CurrencyConverter.FetchURL, staleRatesThreshold)

	currencyConverterTickerTask := task.NewTickerTask(fetchingInterval, currencyConverter)
	currencyConverterTickerTask.Start()

	_, err := router.New(cfg, currencyConverter)
	if err != nil {
		return err
	}

	return nil
}

func OrtbAuction(w http.ResponseWriter, r *http.Request) error {
	return router.OrtbAuctionEndpointWrapper(w, r)
}

var VideoAuction = func(w http.ResponseWriter, r *http.Request) error {
	return router.VideoAuctionEndpointWrapper(w, r)
}

func GetUIDS(w http.ResponseWriter, r *http.Request) {
	router.GetUIDSWrapper(w, r)
}

func SetUIDS(w http.ResponseWriter, r *http.Request) {
	router.SetUIDSWrapper(w, r)
}

func CookieSync(w http.ResponseWriter, r *http.Request) {
	router.CookieSync(w, r)
}

func SyncerMap() map[string]usersync.Syncer {
	return router.SyncerMap()
}
