package models

const (
	//Device.DeviceType values as per OpenRTB-API-Specification-Version-2-5
	DeviceTypeMobile           = 1
	DeviceTypePersonalComputer = 2
	DeviceTypeConnectedTv      = 3
	DeviceTypePhone            = 4
	DeviceTypeTablet           = 5
	DeviceTypeConnectedDevice  = 6
	DeviceTypeSetTopBox        = 7
)

// DevicePlatform defines enums as per int values from KomliAdServer.platform table
type DevicePlatform int

const (
	DevicePlatformUnknown          DevicePlatform = -1
	DevicePlatformDesktop          DevicePlatform = 1 //Desktop Web
	DevicePlatformMobileWeb        DevicePlatform = 2 //Mobile Web
	DevicePlatformNotDefined       DevicePlatform = 3
	DevicePlatformMobileAppIos     DevicePlatform = 4 //In-App iOS
	DevicePlatformMobileAppAndroid DevicePlatform = 5 //In-App Android
	DevicePlatformMobileAppWindows DevicePlatform = 6
	DevicePlatformConnectedTv      DevicePlatform = 7 //Connected TV
)

// DeviceIFAType defines respective logger int id for device type
type DeviceIFAType = int

// DeviceIFATypeID
var DeviceIFATypeID = map[string]DeviceIFAType{
	DeviceIFATypeDPID:      1,
	DeviceIFATypeRIDA:      2,
	DeviceIFATypeAAID:      3,
	DeviceIFATypeIDFA:      4,
	DeviceIFATypeAFAI:      5,
	DeviceIFATypeMSAI:      6,
	DeviceIFATypePPID:      7,
	DeviceIFATypeSSPID:     8,
	DeviceIFATypeSESSIONID: 9,
}

// Device Ifa type constants
const (
	DeviceIFATypeDPID      = "dpid"
	DeviceIFATypeRIDA      = "rida"
	DeviceIFATypeAAID      = "aaid"
	DeviceIFATypeIDFA      = "idfa"
	DeviceIFATypeAFAI      = "afai"
	DeviceIFATypeMSAI      = "msai"
	DeviceIFATypePPID      = "ppid"
	DeviceIFATypeSSPID     = "sspid"
	DeviceIFATypeSESSIONID = "sessionid"
)

// OpenRTB Constants
const (
	//constant for adformat
	Banner = "banner"
	Video  = "video"
	Native = "native"

	//constants for reading video keys from adunit Config
	VideoMinDuration    = "minduration"
	VideoMaxDuration    = "maxduration"
	VideoSkip           = "skip"
	VideoSkipMin        = "skipmin"
	VideoSkipAfter      = "skipafter"
	VideoBattr          = "battr"
	VideoConnectionType = "connectiontype"
	VideoMinBitRate     = "minbitrate"
	VideoMaxBitRate     = "maxbitrate"
	VideoMaxExtended    = "maxextended"
	VideoStartDelay     = "startdelay"
	VideoPlacement      = "placement"
	VideoLinearity      = "linearity"
	VideoMimes          = "mimes"
	VideoProtocol       = "protocol"
	VideoProtocols      = "protocols"
	VideoW              = "w"
	VideoH              = "h"
	VideoSequence       = "sequence"
	VideoBoxingAllowed  = "boxingallowed"
	VideoPlaybackMethod = "playbackmethod"
	VidepPlaybackEnd    = "playbackend"
	VideoDelivery       = "delivery"
	VideoPos            = "pos"
	VideoAPI            = "api"
	VideoCompanionType  = "companiontype"
	VideoComapanionAd   = "companionad"

	//banner obj
	BannerFormat   = "format"
	BannerW        = "w"
	BannerH        = "h"
	BannerWMax     = "wmax"
	BannerHMax     = "hmax"
	BannerWMin     = "wmin"
	BannerHMin     = "hmin"
	BannerBType    = "btype"
	BannerBAttr    = "battr"
	BannerPos      = "pos"
	BannerMimes    = "mimes"
	BannerTopFrame = "topframe"
	BannerExpdir   = "expdir"
	BannerAPI      = "api"
	BannerID       = "id"
	BannerVcm      = "vcm"

	//format object
	FormatW      = "w"
	FormatH      = "h"
	FormatWRatio = "wratio"
	FormatHRatio = "hratio"
	FormatWmin   = "wmin"
)

type ConsentType int

const (
	Unknown ConsentType = iota
	TCF_V1
	TCF_V2
	CCPA
)
