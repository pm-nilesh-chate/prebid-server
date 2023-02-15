package openwrap

import (
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
)

var accountIdSearchPath = [...]struct {
	isApp bool
	key   []string
}{
	{true, []string{"app", "publisher", "ext", "prebid", "parentAccount"}},
	{true, []string{"app", "publisher", "id"}},
	{false, []string{"site", "publisher", "ext", "prebid", "parentAccount"}},
	{false, []string{"site", "publisher", "id"}},
}

func searchAccountId(request []byte) (string, bool, error) {
	for _, path := range accountIdSearchPath {
		accountId, exists, err := getStringValueFromRequest(request, path.key)
		if err != nil {
			return "", path.isApp, err
		}
		if exists {
			return accountId, path.isApp, nil
		}
	}
	return "", false, nil
}

func getStringValueFromRequest(request []byte, key []string) (string, bool, error) {
	val, dataType, _, err := jsonparser.Get(request, key...)
	if dataType == jsonparser.NotExist {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if dataType != jsonparser.String {
		return "", true, fmt.Errorf("%s must be a string", strings.Join(key, "."))
	}
	return string(val), true, nil
}
