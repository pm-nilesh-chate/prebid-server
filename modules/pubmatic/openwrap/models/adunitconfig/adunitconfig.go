package adunitconfig

import (
	"encoding/json"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// AdUnitConfig type definition for Ad Unit config parsed from stored config JSON
type AdUnitConfig struct {
	ConfigPattern string               `json:"configpattern"`
	Config        map[string]*AdConfig `json:"config"`
}
type Content struct {
	Mappings  map[string]openrtb_ext.TransparencyRule `json:"mappings,omitempty"`
	Dimension []string                                `json:"dimension,omitempty"`
}
type Transparency struct {
	Content Content `json:"content,omitempty"`
}

type BannerConfig struct {
	openrtb2.Banner
	ClientConfig json.RawMessage `json:"clientconfig"`
}

type Banner struct {
	Enabled *bool         `json:"enabled"`
	Config  *BannerConfig `json:"config"`
}
type Native struct {
	Enabled *bool `json:"enabled"`
}

type VideoConfig struct {
	openrtb2.Video
	ConnectionType []int           `json:"connectiontype,omitempty"`
	ClientConfig   json.RawMessage `json:"clientconfig"`
}

type Video struct {
	Enabled *bool        `json:"enabled"`
	Config  *VideoConfig `json:"config"`
}

type AdConfig struct {
	Exp          *int             `json:"exp"`
	Banner       *Banner          `json:"banner"`
	Native       *Native          `json:"native"`
	Video        *Video           `json:"video"`
	App          *openrtb2.App    `json:"app"`
	Device       *openrtb2.Device `json:"device"`
	Transparency *Transparency    `json:"transparency,omitempty"`
	Regex        *bool            `json:"regex"`
}
