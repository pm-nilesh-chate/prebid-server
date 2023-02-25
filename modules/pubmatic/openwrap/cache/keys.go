package cache

import (
	"fmt"
)

const (
	PUB_SLOT_INFO  = "pslot_%d_%d_%d_%d" // publisher slot mapping at publisher, profile, display version and adapter level
	PUB_HB_PARTNER = "hbplist_%d_%d_%d"  // header bidding partner list at publishr,profile, display version level
	//HB_PARTNER_CFG = "hbpcfg_%d"         // header bidding partner configuration at partner level
	//PubAadunitConfig - this key for storing adunit config at pub, profile and version level
	PubAdunitConfig = "aucfg_%d_%d_%d"
	PubSlotHashInfo = "pshash_%d_%d_%d_%d"     // slot and its hash info at publisher, profile, display version and adapter level
	PubSlotRegex    = "psregex_%d_%d_%d_%d_%s" // slot and its matching regex info at publisher, profile, display version and adapter level
	PubSlotNameHash = "pslotnamehash_%d"       //publisher slotname hash mapping cache key
	PubVASTTags     = "pvasttags_%d"           //publisher level vasttags
)

func key(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}
