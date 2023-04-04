package models

// Tracker tracker url creation parameters
type Tracker struct {
	PubID             int
	PageURL           string
	Timestamp         int64
	IID               string
	ProfileID         string
	VersionID         string
	SlotID            string
	Adunit            string
	PartnerInfo       *Partner
	RewardedInventory int
	SURL              string // contains either req.site.domain or req.app.bundle value
	Platform          int
	Advertiser        string
	// SSAI identifies the name of the SSAI vendor
	// Applicable only in case of incase of video/json endpoint.
	SSAI string

	ImpID  string `json:"-"`
	Secure int    `json:"-"`
}

// Partner partner information to be logged in tracker object
type Partner struct {
	PartnerID  string
	BidderCode string
	KGPV       string
	GrossECPM  float64
	NetECPM    float64
	BidID      string
	OrigBidID  string
}
