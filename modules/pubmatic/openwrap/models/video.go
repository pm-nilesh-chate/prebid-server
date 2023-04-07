package models

const (
	//VideoVASTTag video VAST parameter constant
	VideoVASTTag = "./VAST"
	//VideoVASTVersion video version parameter constant
	VideoVASTVersion = "version"
	//VideoVASTVersion2_0 video version 2.0 parameter constant
	VideoVASTVersion2_0 = "2.0"
	//VideoVASTAdWrapperTag video ad/wrapper element constant
	VideoVASTAdWrapperTag = "./Ad/Wrapper"
	//VideoVASTAdInLineTag video ad/inline element constant
	VideoVASTAdInLineTag = "./Ad/InLine"
	//VideoExtensionsTag video extensions element constant
	VideoExtensionsTag = "Extensions"
	//VideoExtensionTag video extension element constant
	VideoExtensionTag = "Extension"
	//VideoPricingTag video pricing element constant
	VideoPricingTag = "Pricing"
	//VideoPricingModel video model attribute constant
	VideoPricingModel = "model"
	//VideoPricingModelCPM video cpm attribute value constant
	VideoPricingModelCPM = "CPM"
	//VideoPricingCurrencyUSD video USD default currency constant
	VideoPricingCurrencyUSD = "USD"
	//VideoPricingCurrency video currency constant
	VideoPricingCurrency = "currency"
	//VideoTagLookupStart video xpath constant
	VideoTagLookupStart = "./"
	//VideoTagForwardSlash video forward slash for xpath constant
	VideoTagForwardSlash = "/"
	//VideoVAST2ExtensionPriceElement video parameter constant
	VideoVAST2ExtensionPriceElement = VideoTagLookupStart + VideoExtensionTag + VideoTagForwardSlash + VideoPricingTag
)
