package router

import (
	"github.com/julienschmidt/httprouter"
)

const (
	OpenWrapAuction = "/pbs/openrtb2/auction"
	OpenWrapV25     = "/openrtb/2.5"
	OpenWrapVideo   = "/openrtb/video"
	OpenWrapAmp     = "/openrtb/amp"
)

// Support legacy APIs for a grace period.
// not implementing middleware to avoid duplicate processing like read, unmarshal, write, etc.
// handling the temporary middleware stuff in EntryPoint hook.
func (r *Router) registerOpenWrapEndpoints(openrtbEndpoint, ampEndpoint httprouter.Handle) {
	//OpenWrap hybrid
	r.POST(OpenWrapAuction, openrtbEndpoint)

	// OpenWrap 2.5 in-app, etc
	r.POST(OpenWrapV25, openrtbEndpoint)

	// OpenWrap 2.5 video
	r.POST(OpenWrapVideo, openrtbEndpoint)

	// OpenWrap AMP
	r.POST(OpenWrapAmp, ampEndpoint)
}
