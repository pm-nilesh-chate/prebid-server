package router

import "github.com/julienschmidt/httprouter"

func (r *Router) registerOpenWrapEndpoints(openrtbEndpoint httprouter.Handle) {
	r.POST("/openrtb/2.5", openrtbEndpoint)
}
