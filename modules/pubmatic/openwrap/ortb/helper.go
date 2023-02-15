package ortb

// GetPublisherID returns publisher ID from request
func GetPublisherID(req *BidRequest) string {
	pubID := ""
	if req.Site != nil {
		pubID = *req.Site.Publisher.Id
	} else {
		pubID = *req.App.Publisher.Id
	}
	return pubID
}
