package adapters

import (
	"encoding/json"
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"strings"

	"github.com/prebid/prebid-server/openrtb_ext"
)

// BidderParamJSON defines type as per JSON schema files in static/bidder-param
type BidderParamJSON struct {
	Title        string                     `json:"title"`
	Properties   map[string]BidderParameter `json:"properties"`
	Required     []string                   `json:"required"`
	OneOf        interface{}                `json:"oneOf"`
	Not          interface{}                `json:"not"`
	AnyOf        interface{}                `json:"anyOf"`
	Dependencies interface{}                `json:"dependencies"`
}

// BidderParameter defines properties type as per JSON schema files in static/bidder-param
type BidderParameter struct {
	Type  interface{}    `json:"type"`
	Items ArrayItemsType `json:"items"`
}

// ArrayItemsType defines items type as per JSON schema files in static/bidder-param
type ArrayItemsType struct {
	Type string `json:"type"`
}

// ParameterMapping holds mapping information for bidder parameter
type ParameterMapping struct {
	BidderParamName string      `json:"bidderParameterName"`
	KeyName         string      `json:"keyName"`
	Datatype        string      `json:"type"`
	Required        bool        `json:"required"`
	DefaultValue    interface{} `json:"defaultValue"`
}

func (pm *ParameterMapping) String() string {
	return fmt.Sprintf("[bidderParameterName:%s, keyName:%s, type:%s, required:%v, defaultValue:%v]",
		pm.BidderParamName, pm.KeyName, pm.Datatype, pm.Required, pm.DefaultValue)
}

func parseBidderParams() error {
	schemas, err := parseBidderSchemaDefinitions()
	if err != nil {
		return err
	}
	owParameterMappings := parseOpenWrapParameterMappings()

	adapterParams = make(map[string]map[string]*ParameterMapping)

	for bidderName, jsonSchema := range schemas {

		if jsonSchema.OneOf != nil || jsonSchema.AnyOf != nil || jsonSchema.Not != nil || jsonSchema.Dependencies != nil {
			//JSON schema definition is complex and we rely on case block for this bidder
			continue
		}

		parameters := make(map[string]*ParameterMapping)
		for propertyName, propertyDef := range jsonSchema.Properties {
			bidderParam := ParameterMapping{}
			bidderParam.BidderParamName = propertyName
			bidderParam.KeyName = propertyName
			bidderParam.Datatype = getType(propertyDef)
			bidderParam.Required = false

			parameters[propertyName] = &bidderParam
		}

		owParameterOverrides := owParameterMappings[bidderName]
		for propertyName, propertyDef := range owParameterOverrides {
			if parameters[propertyName] != nil {
				parameter := parameters[propertyName]
				if propertyDef.BidderParamName != "" {
					parameter.BidderParamName = propertyDef.BidderParamName
				}
				if propertyDef.KeyName != "" {
					parameter.KeyName = propertyDef.KeyName
				}
				if propertyDef.Datatype != "" {
					parameter.Datatype = propertyDef.Datatype
				}
				if propertyDef.DefaultValue != nil {
					parameter.DefaultValue = propertyDef.DefaultValue
				}
				parameter.Required = propertyDef.Required
			} else {
			}
		}

		for _, propertyName := range jsonSchema.Required {
			if parameters[propertyName] != nil {
				parameters[propertyName].Required = true
			} else {
			}
		}

		adapterParams[bidderName] = parameters
	}

	return nil
}

func getType(param BidderParameter) string {
	tp := ""
	switch param.Type.(type) {
	case string:
		tp = param.Type.(string)
	case []string:
		v := param.Type.([]string)
		tp = v[0]
		for _, typ := range v {
			if typ == "string" {
				tp = "string"
			}
		}
	}
	if tp == "array" {
		tp = fmt.Sprintf("[]%s", param.Items.Type)
	}
	return tp
}

func parseBidderSchemaDefinitions() (map[string]*BidderParamJSON, error) {
	schemas := make(map[string]*BidderParamJSON)

	schemaDirectory := getBidderParamsDirectory()
	if schemaDirectory == "" {
		return schemas, errors.New("Error failed to parse bidder params files")
	}

	fileInfos, err := os.ReadDir(schemaDirectory)
	if err != nil {
		return schemas, errors.New("Error failed to parse bidder params files" + err.Error())
	}

	bidderMap := openrtb_ext.BuildBidderMap()

	for _, fileInfo := range fileInfos {
		bidderName := strings.TrimSuffix(fileInfo.Name(), ".json")
		if _, isValid := bidderMap[bidderName]; !isValid {
			continue
		}
		_, err := filepath.Abs(filepath.Join(schemaDirectory, fileInfo.Name()))
		if err != nil {
			continue
		}
		fileBytes, err := os.ReadFile(fmt.Sprintf("%s/%s", schemaDirectory, fileInfo.Name()))
		if err != nil {
			continue
		}

		var bidderParamJSON BidderParamJSON
		err = json.Unmarshal(fileBytes, &bidderParamJSON)
		if err != nil {
			continue
		}

		schemas[bidderName] = &bidderParamJSON
	}

	if len(schemas) == 0 {
		return schemas, errors.New("Error failed to parse bidder params files")
	}

	return schemas, nil
}

func getBidderParamsDirectory() string {
	schemaDirectory := "./static/bidder-params"
	if isDirectoryExists(schemaDirectory) {
		return schemaDirectory
	}

	return ""
}

func parseOpenWrapParameterMappings() map[string]map[string]*ParameterMapping {
	owParameterMappings := make(map[string]map[string]*ParameterMapping)

	schemaDirectory := getBidderParamsDirectory()
	if schemaDirectory == "" {
		return owParameterMappings
	}

	fileBytes, err := os.ReadFile("parameterMappings.json")
	if err != nil {
		return owParameterMappings
	}

	err = json.Unmarshal(fileBytes, &owParameterMappings)
	if err != nil {
		return owParameterMappings
	}
	return owParameterMappings
}
