package openwrap

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// populateAndLogRegex populates compiled regex object in container with respective needed regex patterns and logs Invalid Regular expression
func populateAndLogRegex(adUnitCfgs *adunitconfig.AdUnitConfig) []error {
	// for expression := range adUnitCfgs {
	// 	//excluding keys for "regex","configPattern" , "default" to compile regexs
	// 	if expression != models.AdunitConfigRegex && expression != models.AdunitConfigConfigPatternKey && expression != models.AdunitConfigDefaultKey {
	// 		//Populating and Validating
	// 		_, err := Compile(expression)
	// 		if err != nil {
	// 			// return err
	// 		}
	// 	}
	// }
	return nil
}
