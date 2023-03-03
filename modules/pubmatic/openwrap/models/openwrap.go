package models

type RequestCtx struct {
	PubID, ProfileID, DisplayID, VersionID int
	SSAuction                              int
	SummaryDisable                         int
	LogInfoFlag                            int
	PartnerConfigMap                       map[int]map[string]string
	PreferDeals                            bool
	Platform                               string

	//NYC_TODO: use enum?
	IsTestRequest bool
	IsCTVRequest  bool

	UA      string
	Cookies string

	Debug bool
}
