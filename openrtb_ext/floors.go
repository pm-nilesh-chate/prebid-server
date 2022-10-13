package openrtb_ext

// PriceFloorRules defines the contract for bidrequest.ext.prebid.floors
type PriceFloorRules struct {
	FloorMin    float64                `json:"floormin,omitempty"`
	FloorMinCur string                 `json:"floormincur,omitempty"`
	SkipRate    int                    `json:"skiprate,omitempty"`
	Location    *PriceFloorEndpoint    `json:"location,omitempty"`
	Data        *PriceFloorData        `json:"data,omitempty"`
	Enforcement *PriceFloorEnforcement `json:"enforcement,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Skipped     *bool                  `json:"skipped,omitempty"`
}

type PriceFloorEndpoint struct {
	URL string `json:"url,omitempty"`
}

type PriceFloorData struct {
	Currency            string                 `json:"currency,omitempty"`
	SkipRate            int                    `json:"skiprate,omitempty"`
	FloorsSchemaVersion string                 `json:"floorsschemaversion,omitempty"`
	ModelTimestamp      int                    `json:"modeltimestamp,omitempty"`
	ModelGroups         []PriceFloorModelGroup `json:"modelgroups,omitempty"`
}

type PriceFloorModelGroup struct {
	Currency     string             `json:"currency,omitempty"`
	ModelWeight  int                `json:"modelweight,omitempty"`
	DebugWeight  int                `json:"debugweight,omitempty"` // Added for Debug purpose, shall be removed
	ModelVersion string             `json:"modelversion,omitempty"`
	SkipRate     int                `json:"skiprate,omitempty"`
	Schema       PriceFloorSchema   `json:"schema,omitempty"`
	Values       map[string]float64 `json:"values,omitempty"`
	Default      float64            `json:"default,omitempty"`
}
type PriceFloorSchema struct {
	Fields    []string `json:"fields,omitempty"`
	Delimiter string   `json:"delimiter,omitempty"`
}

type PriceFloorEnforcement struct {
	EnforcePBS    *bool `json:"enforcepbs,omitempty"`
	FloorDeals    *bool `json:"floordeals,omitempty"`
	BidAdjustment bool  `json:"bidadjustment,omitempty"`
	EnforceRate   int   `json:"enforcerate,omitempty"`
}

// GetEnabled will check if floors is enabled in request
func (Floors *PriceFloorRules) GetEnabled() bool {
	if Floors != nil && Floors.Enabled != nil && !*Floors.Enabled {
		return *Floors.Enabled
	}
	return true
}