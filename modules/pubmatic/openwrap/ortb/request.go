package ortb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type BidRequest struct {
	Id      *string     `json:"id"`               //Unique ID of the bid request
	Imp     []*Imp      `json:"imp"`              // Array of Imp objects
	Site    *Site       `json:"site,omitempty"`   // Details via Site Object
	App     *App        `json:"app,omitempty"`    // Details via App object
	Device  *Device     `json:"device,omitempty"` // Details via Device object
	User    *User       `json:"user,omitempty"`   // Details via User object
	Source  *Source     `json:"source,omitempty"` // Details via Source object
	Test    *int        `json:"test,omitempty"`   //Indicator of test mode
	At      *int        `json:"at,omitempty"`     // default 2, Auction type, where 1 = First Price, 2 = Second Price Plus
	Tmax    *int        `json:"tmax,omitempty"`   // max timeout
	Wseat   []string    `json:"wseat,omitempty"`
	Wlang   []string    `json:"wlang,omitempty"`
	Bseat   []string    `json:"bseat,omitempty"`
	AllImps *int        `json:"allimps,omitempty"`
	Cur     []string    `json:"cur,omitempty"`  // Array of allowed currencies for bids on this bid request using ISO-4217 alpha codes
	Bcat    []string    `json:"bcat,omitempty"` // Blocked advertiser categories using the IAB content categories.
	Badv    []string    `json:"badv,omitempty"` // Block list of advertisers by their domains
	Bapp    []string    `json:"bapp,omitempty"` //Block list of applications by their platform-specific exchange
	Regs    *Regs       `json:"regs,omitempty"`
	Ext     interface{} `json:"ext,omitempty"`
}

type Source struct {
	FD     *int8       `json:"fd,omitempty"`     // Entity responsible for final impression sale decision, 0 = exchange, 1 = upstream source.
	TID    *string     `json:"tid,omitempty"`    //   Transaction ID that must be common across all participants in this bid request (e.g., potentially multiple exchanges).
	PChain *string     `json:"pchain,omitempty"` //   Payment ID chain string containing embedded syntax  described in the TAG Payment ID Protocol v1.0.
	Ext    interface{} `json:"ext,omitempty"`
}

// A Regs object that specifies any industry, legal,
// or governmental regulations in force for this request.
type Regs struct {
	Coppa *int        `json:"coppa,omitempty"` // Flag indicating if this request is subject to the COPPA regulations
	Ext   interface{} `json:"ext,omitempty"`
}

type Imp struct {
	Id                *string     `json:"id"`               //A unique identifier for this impression, typically, starts with 1 and increments
	Metric            []*Metric   `json:"metric,omitempty"` // Array of metric object
	Banner            *Banner     `json:"banner,omitempty"`
	Video             *Video      `json:"video,omitempty"`
	Audio             *Audio      `json:"audio,omitempty"`
	Native            *Native     `json:"native,omitempty"`
	Pmp               *Pmp        `json:"pmp,omitempty"`               //Pmp object containing any private marketplace deals in effect for this impression.
	DisplayManager    *string     `json:"displaymanager,omitempty"`    // Name of ad mediation partner
	DisplayManagerVer *string     `json:"displaymanagerver,omitempty"` // Version of ad mediation partner
	Instl             *int        `json:"instl,omitempty"`             // 1 = the ad is interstitial or full screen, 0 = not interstitial
	TagId             *string     `json:"tagid,omitempty"`             // Identifier for specific ad placement,  useful for debugging
	BidFloor          *float64    `json:"bidfloor,omitempty"`          // Minimum bid for this impression expressed in CPM
	BidFloorCur       *string     `json:"bidfloorcur,omitempty"`       // default “USD”, Currency specified using ISO-4217 alpha codes,
	ClickBrowser      *int        `json:"clickbrowser,omitempty"`      //Indicates the type of browser opened upon clicking the creative in an app, where 0 = embedded, 1 = native.
	Secure            *int        `json:"secure,omitempty"`            // Flag to indicate if the impression requires secure HTTPS URL creative assets and markup
	IframeBuster      []string    `json:"iframebuster,omitempty"`      // Array of exchange-specific names of supported iframe busters
	Exp               *int        `json:"exp,omitempty"`               //Advisory as to the number of seconds that may elapse between the auction and the actual impression
	Ext               interface{} `json:"ext,omitempty"`
}

// Metric  object is associated with an impression as an array of metrics
type Metric struct {
	Type   *string     `json:"type"`             // Type of metric being presented using exchange
	Value  *float32    `json:"value"`            // Number representing the value of the metric, 0.0 – 1.0.
	Vendor *string     `json:"vendor,omitempty"` // Source of the value using exchange curated string names which should be published to bidders a priori.
	Ext    interface{} `json:"ext,omitempty"`
}

// A Banner object; required if this impression is
// offered as a banner ad opportunity.
type Banner struct {
	Format   []*Format   `json:"format,omitempty"`   //Array of format objects representing the banner sizes permitted.
	W        *int        `json:"w,omitempty"`        // Width of the impression in pixels
	H        *int        `json:"h,omitempty"`        // Height of the impression in pixels
	WMax     *int        `json:"wmax,omitempty"`     // Maximum Width of the impression in pixels
	HMax     *int        `json:"hmax,omitempty"`     // Maximum Height of the impression in pixels
	WMin     *int        `json:"wmin,omitempty"`     // Minimum width of the impression in pixels
	HMin     *int        `json:"hmin,omitempty"`     // Minimum height of the impression in pixels
	BType    []int       `json:"btype,omitempty"`    // Blocked banner ad types
	BAttr    []int       `json:"battr,omitempty"`    // Blocked creative attributes
	Pos      *int        `json:"pos,omitempty"`      // Ad position on screen
	Mimes    []string    `json:"mimes,omitempty"`    // Content MIME types supported
	TopFrame *int        `json:"topframe,omitempty"` // Indicates if the banner is in the top frame as opposed to an iframe
	Expdir   []int       `json:"expdir,omitempty"`   // Directions in which the banner may expand
	API      []int       `json:"api,omitempty"`      // List of supported API frameworks for this impression
	ID       *string     `json:"id,omitempty"`       // Unique identifier for this banner object
	Vcm      *int        `json:"vcm,omitempty"`      //Relevant only for Banner objects used with a Video object
	Ext      interface{} `json:"ext,omitempty"`
}

// A Video object; required if this impression is
// offered as a video ad opportunity.
type Video struct {
	Mimes          []string    `json:"mimes,omitempty"`          // Content MIME types supported.
	MinDuration    *int        `json:"minduration,omitempty"`    // Minimum video ad duration in seconds.
	MaxDuration    *int        `json:"maxduration,omitempty"`    // Maximum video ad duration in seconds.
	Protocols      []int       `json:"protocols,omitempty"`      // Array of supported video bid response protocols
	Protocol       *int        `json:"protocol,omitempty"`       // Supported Video Protocol, At least one supported protocol must be specified in either the protocol or protocols attribute.
	W              *int        `json:"w,omitempty"`              // Width of the video player in pixels
	H              *int        `json:"h,omitempty"`              // Height of the video player in pixels
	StartDelay     *int        `json:"startdelay,omitempty"`     // Indicates the start delay in seconds
	Placement      *int        `json:"placement,omitempty"`      // Placement type for the impression.
	Linearity      *int        `json:"linearity,omitempty"`      // Indicates if the impression must be linear, nonlinear, etc
	Skip           *int        `json:"skip,omitempty"`           //Indicates if the player will allow the video to be skipped,where 0 = no, 1 = yes.
	SkipMin        *int        `json:"skipmin,omitempty"`        //Videos of total duration greater than this number of seconds can be skippable; only applicable if the ad is skippable.
	SkipAfter      *int        `json:"skipafter,omitempty"`      //Number of seconds a video must play before skipping is enabled; only applicable if the ad is skippable.
	Sequence       *int        `json:"sequence,omitempty"`       // If multiple ad impressions are offered in the same bid request, the sequence number will allow for the coordinated delivery of multiple creatives
	BAttr          []int       `json:"battr,omitempty"`          // Blocked creative attributes
	MaxExtended    *int        `json:"maxextended,omitempty"`    // Maximum extended video ad duration if extension is allowed
	MinBitrate     *int        `json:"minbitrate,omitempty"`     // Minimum bit rate in Kbps
	MaxBitrate     *int        `json:"maxbitrate,omitempty"`     // Maximum bit rate in Kbps
	BoxingAllowed  *int        `json:"boxingallowed,omitempty"`  // Indicates if letter-boxing of 4:3 content into a 16:9 window is allowed,
	PlaybackMethod []int       `json:"playbackmethod,omitempty"` // Allowed playback methods
	PlaybackEnd    *int        `json:"playbackend,omitempty"`    // The event that causes playback to end.
	Delivery       []int       `json:"delivery,omitempty"`       // Supported delivery methods (e.g., streaming, progressive)
	Pos            *int        `json:"pos,omitempty"`            // Ad position on screen
	Companionad    []*Banner   `json:"companionad,omitempty"`    // Array of Banner objects
	API            []int       `json:"api,omitempty"`            // List of supported API frameworks for this impression
	CompanionType  []int       `json:"companiontype,omitempty"`  // Supported VAST companion ad types
	Ext            interface{} `json:"ext,omitempty"`
}

// A Audio object; required if this impression is
// offered as a Audio ad opportunity.
type Audio struct {
	Mimes         []string    `json:"mimes,omitempty"`         //Content MIME types supported
	MinDuration   *int        `json:"minduration,omitempty"`   //Minimum audio ad duration in seconds.
	MaxDuration   *int        `json:"maxduration,omitempty"`   //Maximum audio ad duration in seconds.
	Protocols     []int       `json:"protocols,omitempty"`     //Array of supported audio protocols.
	StartDelay    *int        `json:"startdelay,omitempty"`    //Indicates the start delay in seconds for pre-roll, mid-roll, or post-roll ad placements.
	Sequence      *int        `json:"sequence,omitempty"`      //If multiple ad impressions are offered in the same bid request, the sequence number will allow for the coordinated delivery of multiple creatives.
	BAttr         []int       `json:"battr,omitempty"`         //Blocked creative attributes
	MaxExtended   *int        `json:"maxextended,omitempty"`   //Maximum extended ad duration if extension is allowed
	MinBitrate    *int        `json:"minbitrate,omitempty"`    //Minimum bit rate in Kbps.
	MaxBitrate    *int        `json:"maxbitrate,omitempty"`    //Maximum bit rate in Kbps.
	Delivery      []int       `json:"delivery,omitempty"`      //Supported delivery methods
	Companionad   []*Banner   `json:"companionad,omitempty"`   //Array of Banner objects if companion ads are available
	API           []int       `json:"api,omitempty"`           //List of supported API frameworks for this impression.
	CompanionType []int       `json:"companiontype,omitempty"` //Supported DAAST companion ad types.
	MaxSeq        *int        `json:"maxseq,omitempty"`        //maximum number of ads that can be played in an ad pod.
	Feed          *int        `json:"feed,omitempty"`          //Type of audio feed.
	Stitched      *int        `json:"stitched,omitempty"`      //Indicates if the ad is stitched with audio content or delivered independently
	NVol          *int        `json:"nvol,omitempty"`          //Volume normalization mode.
	Ext           interface{} `json:"ext,omitempty"`           //Placeholder for exchange-specific extensions to OpenRTB.
}

// A Native object; required if this impression is
// offered as a native ad opportunity.
type Native struct {
	Request *string     `json:"request"`         // Request payload complying with the Native Ad Specification
	Ver     *string     `json:"ver,omitempty"`   // Version of the Native Ad Specification to which request complies; highly recommended for efficient parsing
	API     []int       `json:"api,omitempty"`   // List of supported API frameworks for this impression
	BAttr   []int       `json:"battr,omitempty"` // Blocked creative attributes
	Ext     interface{} `json:"ext,omitempty"`
}

type Format struct {
	W      *int        `json:"w,omitempty"`      //Width in device independent pixels (DIPS).
	H      *int        `json:"h,omitempty"`      //Height in device independent pixels (DIPS).
	Wratio *int        `json:"wratio,omitempty"` // Relative width when expressing size as a ratio.
	Hratio *int        `json:"hratio,omitempty"` //  Relative height when expressing size as a ratio.
	Wmin   *int        `json:"wmin,omitempty"`   // The minimum width in device independent pixels (DIPS) at 	//which the ad will be displayed the size is expressed as a ratio.
	Ext    interface{} `json:"ext,omitempty"`    //Placeholder for exchange-specific extensions to OpenRTB.
}

// Pmp object private marketplace container for direct deals between buyers and sellers
type Pmp struct {
	PrivateAuction *int        `json:"private_auction,omitempty"` // Indicator of auction eligibility to seats named in the Direct Deals object, where 0 = all bids are accepted
	Deals          []*Deal     `json:"deals,omitempty"`
	Ext            interface{} `json:"ext,omitempty"`
}

type Deal struct {
	ID          *string     `json:"id"`                    // A unique identifier for the direct deal
	BidFloor    *float64    `json:"bidfloor,omitempty"`    // Minimum bid for this impression expressed in CPM
	BidFloorCur *string     `json:"bidfloorcur,omitempty"` // Currency specified using ISO-4217 alpha codes
	At          *int        `json:"at,omitempty"`          // Optional override of the overall auction type of the bid request, where 1 = First Price, 2 = Second Price Plus
	WSeat       []string    `json:"wseat,omitempty"`       // Whitelist of buyer seats allowed to bid on this deal
	WaDomain    []string    `json:"wadomain,omitempty"`    // Array of advertiser domains
	Ext         interface{} `json:"ext,omitempty"`
}

// Details of publisher’s website
// Only applicable and recommended for websites
type Site struct {
	Id            *string     `json:"id,omitempty"`            // Exchange-specific site ID
	Name          *string     `json:"name,omitempty"`          // Site name (may be aliased at the publisher’s request).
	Domain        *string     `json:"domain,omitempty"`        // Domain of the site
	Cat           []string    `json:"cat,omitempty"`           // Array of IAB content categories of the site
	SectionCat    []string    `json:"sectioncat,omitempty"`    // Array of IAB content categories that describe the current section of the site
	PageCat       []string    `json:"pagecat,omitempty"`       // Array of IAB content categories that describe the current page or view of the site
	Page          *string     `json:"page,omitempty"`          // URL of the page where the impression will be shown
	Ref           *string     `json:"ref,omitempty"`           // Referrer URL that caused navigation to the current page
	Search        *string     `json:"search,omitempty"`        // Search *string that caused navigation to the current page
	Mobile        *int        `json:"mobile,omitempty"`        // Mobile-optimized signal, where 0 = no, 1 = yes
	PrivacyPolicy *int        `json:"privacypolicy,omitempty"` // Indicates if the site has a privacy policy, where 0 = no, 1 = yes
	Publisher     *Publisher  `json:"publisher,omitempty"`     // Details about the Publisher of the site
	Content       *Content    `json:"content,omitempty"`       // Details about the Content within the site
	Keywords      *string     `json:"keywords,omitempty"`      // Comma separated list of keywords about the site
	Ext           interface{} `json:"ext,omitempty"`           // Placeholder for exchange-specific extensions to OpenRTB
}

// Details of publisher’s app (i.e., non-browser applications)
// Only applicable and recommended for apps.
type App struct {
	Id            *string     `json:"id,omitempty"`            // Exchange-specific app ID.
	Name          *string     `json:"name,omitempty"`          // App name
	Bundle        *string     `json:"bundle,omitempty"`        // Application bundle or package name
	Domain        *string     `json:"domain,omitempty"`        // Domain of the app
	StoreURL      *string     `json:"storeurl,omitempty"`      // App store URL for an installed app, for QAG 1.5 compliance
	Cat           []string    `json:"cat,omitempty"`           // Array of IAB content categories of the app
	SectionCat    []string    `json:"sectioncat,omitempty"`    // Array of IAB content categories that describe the current section of the app
	PageCat       []string    `json:"pagecat,omitempty"`       // Array of IAB content categories that describe the current page or view of the app
	Ver           *string     `json:"ver,omitempty"`           // Application version
	PrivacyPolicy *int        `json:"privacypolicy,omitempty"` // Indicates if the app has a privacy policy, where 0 = no, 1 = yes.
	Paid          *int        `json:"paid,omitempty"`          // 0 = app is free, 1 = the app is a paid version
	Publisher     *Publisher  `json:"publisher,omitempty"`     // Details about the Publisher
	Content       *Content    `json:"content,omitempty"`       // Details about the Content
	Keywords      *string     `json:"keywords,omitempty"`      // Comma separated list of keywords about the app
	Ext           interface{} `json:"ext,omitempty"`
}

// Details about the Publisher of the site
type Publisher struct {
	Id     *string     `json:"id,omitempty"`     // Exchange-specific publisher ID
	Name   *string     `json:"name,omitempty"`   // Publisher name
	Cat    []string    `json:"cat,omitempty"`    // Array of IAB content categories that describe the publisher
	Domain *string     `json:"domain,omitempty"` // Highest level domain of the publisher
	Ext    interface{} `json:"ext,omitempty"`
}

// Details about the Content within the site
type Content struct {
	ID                 *string     `json:"id,omitempty"`      // ID uniquely identifying the content
	Episode            *int        `json:"episode,omitempty"` // Episode number (typically applies to video content).
	Title              *string     `json:"title,omitempty"`   // Content title.
	Series             *string     `json:"series,omitempty"`  // Content series
	Season             *string     `json:"season,omitempty"`  // Content season
	Artist             *string     `json:"artist,omitempty"`  //Artist credited with the content.
	Genre              *string     `json:"genre,omitempty"`   //Genre that best describes the content (e.g., rock, pop, etc).
	Album              *string     `json:"album,omitempty"`   //Album to which the content belongs; typically for audio.
	IsRc               *string     `json:"isrc,omitempty"`
	Producer           *Producer   `json:"producer,omitempty"`           // Details about the content Producer
	URL                *string     `json:"url,omitempty"`                // URL of the content, for buy-side contextualization or review
	Cat                []string    `json:"cat,omitempty"`                // Array of IAB content categories that describe the content producer
	ProdQ              *int        `json:"prodq,omitempty"`              // Production quality.
	VideoQuality       *int        `json:"videoquality,omitempty"`       // Video quality per IAB’s classification
	Context            *int        `json:"context,omitempty"`            // Type of content (game, video, text, etc.)
	ContentRating      *string     `json:"contentrating,omitempty"`      // Content rating (e.g., MPAA)
	UserRating         *string     `json:"userrating,omitempty"`         // User rating of the content (e.g., number of stars, likes, etc.)
	QaGmeDiarating     *int        `json:"qagmediarating,omitempty"`     // Media rating per QAG guidelines
	Keywords           *string     `json:"keywords,omitempty"`           // Comma separated list of keywords describing the content
	LiveStream         *int        `json:"livestream,omitempty"`         // 0 = not live, 1 = content is live (e.g., stream, live blog)
	SourceRelationship *int        `json:"sourcerelationship,omitempty"` // 0 = indirect, 1 = direct
	Len                *int        `json:"len,omitempty"`                // Length of content in seconds; appropriate for video or audio
	Language           *string     `json:"language,omitempty"`           // Content language using ISO-639-1-alpha-2.
	Embeddable         *int        `json:"embeddable,omitempty"`         // Indicator of whether or not the content is embeddable
	Data               []Data      `json:"data,omitempty"`
	Ext                interface{} `json:"ext,omitempty"`
}

type Producer struct {
	ID     *string     `json:"id,omitempty"`     // Content producer or originator ID. Useful if content is, syndicated and may be posted on a site using embed tags
	Name   *string     `json:"name,omitempty"`   // Content producer or originator name
	Cat    []string    `json:"cat,omitempty"`    // Array of IAB content categories that describe the content producer
	Domain *string     `json:"domain,omitempty"` // Highest level domain of the content producer
	Ext    interface{} `json:"ext,omitempty"`
}

// Details of Device object about the user’s
// device to which the impression will be delivered.
type Device struct {
	Ua             *string     `json:"ua,omitempty"`             // Browser user agent *string
	Geo            *Geo        `json:"geo,omitempty"`            // Location of the device assumed to be the user’s current location
	Dnt            *int        `json:"dnt,omitempty"`            // Standard “Do Not Track” flag as set in the header by the browser
	Lmt            *int        `json:"lmt,omitempty"`            // “Limit Ad Tracking” signal commercially endorsed
	IP             *string     `json:"ip,omitempty"`             // IPv4 address closest to device
	Ipv6           *string     `json:"ipv6,omitempty"`           // IP address closest to device as IPv6
	DeviceType     *int        `json:"devicetype,omitempty"`     // The general type of device
	Make           *string     `json:"make,omitempty"`           // Device make (e.g., “Apple”).
	Model          *string     `json:"model,omitempty"`          // Device model (e.g., “iPhone”)
	Os             *string     `json:"os,omitempty"`             // Device operating system (e.g., “iOS”)
	Osv            *string     `json:"osv,omitempty"`            // Device operating system version
	Hwv            *string     `json:"hwv,omitempty"`            // Hardware version of the device (e.g., “5S” for iPhone 5S)
	H              *int        `json:"h,omitempty"`              // Physical height of the screen in pixels
	W              *int        `json:"w,omitempty"`              // Physical width of the screen in pixels
	Ppi            *int        `json:"ppi,omitempty"`            // Screen size as pixels per linear inch
	PxRatio        *float64    `json:"pxratio,omitempty"`        // The ratio of physical pixels to device independent pixels
	JS             *int        `json:"js,omitempty"`             // Support for JavaScript, where 0 = no, 1 = yes
	GeoFetch       *int        `json:"geofetch,omitempty"`       //Indicates if the geolocation API will be available to JavaScript code running in the banner, where 0 = no, 1 = yes.
	FlashVer       *string     `json:"flashver,omitempty"`       // Version of Flash supported by the browser
	Language       *string     `json:"language,omitempty"`       // Browser language using ISO-639-1-alpha-2
	Carrier        *string     `json:"carrier,omitempty"`        // Carrier or ISP (e.g., “VERIZON”). “WIFI” is often used in mobile to indicate high bandwidth
	Mccmnc         *string     `json:"mccmnc,omitempty"`         // Mobile carrier as the concatenated MCC-MNC code
	ConnectionType *int        `json:"connectiontype,omitempty"` // Network connection type
	Ifa            *string     `json:"ifa,omitempty"`            // ID sanctioned for advertiser use in the clear
	DidSha1        *string     `json:"didsha1,omitempty"`        // Hardware device ID (e.g., IMEI); hashed via SHA1
	DidMd5         *string     `json:"didmd5,omitempty"`         // Hardware device ID (e.g., IMEI); hashed via MD5
	DpidSha1       *string     `json:"dpidsha1,omitempty"`       // Platform device ID (e.g., Android ID); hashed via SHA1
	DpidMd5        *string     `json:"dpidmd5,omitempty"`        // Platform device ID (e.g., Android ID); hashed via MD5
	MacSha1        *string     `json:"macsha1,omitempty"`        // MAC address of the device; hashed via SHA1
	MacMd5         *string     `json:"macmd5,omitempty"`         // MAC address of the device; hashed via MD5
	Ext            interface{} `json:"ext,omitempty"`
}

// Geo Object
type Geo struct {
	Lat           *float64    `json:"lat,omitempty"`           // Latitude from -90.0 to +90.0, where negative is south
	Lon           *float64    `json:"lon,omitempty"`           // Longitude from -180.0 to +180.0, where negative is west
	Type          *int        `json:"type,omitempty"`          // Source of location data; recommended when passing lat/lon
	Accuracy      *int        `json:"accuracy,omitempty"`      // Estimated location accuracy in meters; recommended when lat/lon are specified and derived from a device’s location services (i.e., type = 1).
	LastFix       *int        `json:"lastfix,omitempty"`       // Number of seconds since this geolocation fix was established.
	IPService     *int        `json:"ipservice,omitempty"`     // Service or provider used to determine geolocation from IP address if applicable (i.e., type = 2).
	Country       *string     `json:"country,omitempty"`       // Country code using ISO-3166-1-alpha-3
	Region        *string     `json:"region,omitempty"`        // Region code using ISO-3166-2; 2-letter state code if USA
	RegionFips104 *string     `json:"regionfips104,omitempty"` // Region of a country using FIPS 10-4 notation
	Metro         *string     `json:"metro,omitempty"`         // Google metro code; similar to but not exactly Nielsen DMAs
	City          *string     `json:"city,omitempty"`          // City using United Nations Code for Trade & Transport Location
	Zip           *string     `json:"zip,omitempty"`           // Zip or postal code
	UtcOffset     *int        `json:"utcoffset,omitempty"`     // Local time as the number +/- of minutes from UTC
	Ext           interface{} `json:"ext,omitempty"`
}

// User object  about the human  user of the device
type User struct {
	ID         *string     `json:"id,omitempty"`         // Exchange-specific ID for the user
	BuyerUID   *string     `json:"buyeruid,omitempty"`   // Buyer-specific ID for the user as mapped by the exchange for the buyer
	Yob        *int        `json:"yob,omitempty"`        // Year of birth as a 4-digit integer
	Gender     *string     `json:"gender,omitempty"`     // Gender, where “M” = male, “F” = female, “O” = known to be other
	Keywords   *string     `json:"keywords,omitempty"`   // Comma separated list of keywords, interests, or intent
	CustomData *string     `json:"customdata,omitempty"` // Optional feature to pass bidder data that was set in the exchange’s cookie
	Geo        *Geo        `json:"geo,omitempty"`        // Location of the user’s home base defined by a Geo object
	Data       []*Data     `json:"data,omitempty"`       // Additional user data
	Ext        interface{} `json:"ext,omitempty"`
}

// Data Object ; Additional User data
type Data struct {
	ID      *string     `json:"id,omitempty"`      // Exchange-specific ID for the data provider
	Name    *string     `json:"name,omitempty"`    // Exchange-specific name for the data provider
	Segment []*Segment  `json:"segment,omitempty"` // Array of Segment  objects that contain the actual data values
	Ext     interface{} `json:"ext,omitempty"`
}

type Segment struct {
	ID    *string     `json:"id,omitempty"`    // ID of the data segment specific to the data provider
	Name  *string     `json:"name,omitempty"`  // Name of the data segment specific to the data provider
	Value *string     `json:"value,omitempty"` // String representation of the data segment value
	Ext   interface{} `json:"ext,omitempty"`
}

// ValidateRequest : This method checks for mandatory parameters for a request
func (req *BidRequest) ValidateRequest() error {

	if nil == req.Id || len(*req.Id) == 0 {
		return errors.New("Invalid Request: Request ID not present")
	}
	if nil == req.Imp || len(req.Imp) <= 0 {
		return errors.New("Invalid Request: Request.Imp not present")
	}

	if nil == req.Site && nil == req.App {
		return errors.New("Invalid request: Site/App Object not present")
	}

	if req.Site != nil && req.App != nil {
		return errors.New("Invalid request: Both Site and App object present")
	}

	if req.Site != nil {
		if err := validateSite(req.Site); err != nil {
			return fmt.Errorf("Invalid Site Object: %v", err)
		}
	} else {
		if err := validateApp(req.App); err != nil {
			return fmt.Errorf("Invalid App Object: %v", err)
		}
	}

	return nil
}

func (imp *Imp) ValidateImp(impIndex int) error {

	if imp.Id == nil || *imp.Id == "" {
		return errors.New("Invalid request! Mandatory field imp.Id missing")
	}

	if imp.TagId == nil || *imp.TagId == "" {
		return fmt.Errorf("Invalid imp:%s ! Mandatory field imp.tagid missing", *imp.Id)
	}

	if imp.Banner == nil && imp.Video == nil {
		return fmt.Errorf("Invalid imp:%s ! Mandatory object Banner/Video missing", *imp.Id)
	}
	var err error
	if err = validateBanner(imp.Banner, impIndex); err != nil {
		return fmt.Errorf("Invalid imp:%s ! Error: %v", *imp.Id, err)
	}
	if err = validateVideo(imp.Video); err != nil {
		return fmt.Errorf("Invalid imp:%s ! Error: %v", *imp.Id, err)
	}

	return nil
}

func validateBanner(banner *Banner, impIndex int) error {
	if nil == banner {
		return nil
	}

	if banner.W != nil && *banner.W <= 0 {
		return fmt.Errorf("request.imp[%d].banner.w must be a positive number", impIndex)
	}
	if banner.H != nil && *banner.H <= 0 {
		return fmt.Errorf("request.imp[%d].banner.h must be a positive number", impIndex)
	}

	hasRootSize := banner.H != nil && banner.W != nil && *banner.H > 0 && *banner.W > 0

	if !hasRootSize && len(banner.Format) == 0 {
		return fmt.Errorf("request.imp[%d].banner has no sizes. Define \"w\" and \"h\", or include \"format\" elements.", impIndex)
	}

	for i, format := range banner.Format {
		if err := validateFormat(format, impIndex, i); err != nil {
			return err
		}
	}

	return nil
}

func validateFormat(format *Format, impIndex, formatIndex int) error {

	if format.W == nil || *format.W <= 0 {
		return fmt.Errorf("request.imp[%d].banner.format[%d].w must be a positive number", impIndex, formatIndex)
	}
	if format.H == nil || *format.H <= 0 {
		return fmt.Errorf("request.imp[%d].banner.format[%d].h must be a positive number", impIndex, formatIndex)
	}

	return nil
}

func validateVideo(video *Video) error {
	if nil == video {
		return nil
	}

	if nil == video.Mimes || len(video.Mimes) == 0 {
		return errors.New("Invalid Request! Mandatory field video.mimes is missing")
	}

	return nil
}

func validateSite(site *Site) error {
	if site != nil && (site.Publisher == nil || site.Publisher.Id == nil) {
		return errors.New("Invalid Request! Mandatory field site.publisher.id missing")
	}

	return validatePubID(*site.Publisher.Id)
}

func validateApp(app *App) error {
	if app != nil && (app.Publisher == nil || app.Publisher.Id == nil) {
		return errors.New("Invalid Request! Mandatory field app.publisher.id missing")
	}

	return validatePubID(*app.Publisher.Id)
}

func validatePubID(pubID string) error {
	if _, err := strconv.Atoi(pubID); err != nil {
		return errors.New("Invalid publisher ID")
	}

	return nil
}

func (req *BidRequest) String() string {
	byts, _ := json.Marshal(req)
	return string(byts)
}

func (user *User) IsUserExtPresent() bool {
	if user != nil && user.Ext != nil {
		return true
	}
	return false
}

func (regs *Regs) IsRegsExtPresent() bool {
	if regs != nil && regs.Ext != nil {
		return true
	}
	return false
}
