package models

import (
	"encoding/json"
	"fmt"
)

func GetRequestExtWrapper(ext []byte) (ExtRequestWrapper, error) {
	extRequest := ExtRequestWrapper{}

	err := json.Unmarshal(ext, &extRequest)
	if err != nil {
		return extRequest, fmt.Errorf("failed to decode request.ext : %v", err)
	}

	return extRequest, nil
}
