package pubmatic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/adapters/adapterstest"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestJsonSamples(t *testing.T) {
	bidder, buildErr := Builder(openrtb_ext.BidderPubmatic, config.Adapter{
		Endpoint: "https://hbopenbid.pubmatic.com/translator?source=prebid-server"})

	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}

	adapterstest.RunJSONBidderTest(t, "pubmatictest", bidder)
}

func TestGetBidTypeVideo(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 1
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeVideo {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeVideo, actualBidTypeValue)
	}
}

func TestGetBidTypeForMissingBidTypeExt(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	actualBidTypeValue := getBidType(pubmaticExt)
	// banner is the default bid type when no bidType key is present in the bid.ext
	if actualBidTypeValue != "banner" {
		t.Errorf("Expected Bid Type value was: banner, actual value is: %v", actualBidTypeValue)
	}
}

func TestGetBidTypeBanner(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 0
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeBanner {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeBanner, actualBidTypeValue)
	}
}

func TestGetBidTypeNative(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 2
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeNative {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeNative, actualBidTypeValue)
	}
}

func TestGetBidTypeForUnsupportedCode(t *testing.T) {
	pubmaticExt := &pubmaticBidExt{}
	pubmaticExt.BidType = new(int)
	*pubmaticExt.BidType = 99
	actualBidTypeValue := getBidType(pubmaticExt)
	if actualBidTypeValue != openrtb_ext.BidTypeBanner {
		t.Errorf("Expected Bid Type value was: %v, actual value is: %v", openrtb_ext.BidTypeBanner, actualBidTypeValue)
	}
}

func TestParseImpressionObject(t *testing.T) {
	type args struct {
		imp                      *openrtb2.Imp
		extractWrapperExtFromImp bool
		extractPubIDFromImp      bool
	}
	tests := []struct {
		name                string
		args                args
		expectedWrapperExt  *pubmaticWrapperExt
		expectedPublisherId string
		wantErr             bool
		expectedBidfloor    float64
	}{
		{
			name: "imp.bidfloor empty and kadfloor set",
			args: args{
				imp: &openrtb2.Imp{
					Video: &openrtb2.Video{},
					Ext:   json.RawMessage(`{"bidder":{"kadfloor":"0.12"}}`),
				},
			},
			expectedBidfloor: 0.12,
		},
		{
			name: "imp.bidfloor set and kadfloor empty",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{}}`),
				},
			},
			expectedBidfloor: 0.12,
		},
		{
			name: "imp.bidfloor set and kadfloor invalid",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":"aaa"}}`),
				},
			},
			expectedBidfloor: 0.12,
		},
		{
			name: "imp.bidfloor set and kadfloor set, preference to kadfloor",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":"0.11"}}`),
				},
			},
			expectedBidfloor: 0.11,
		},
		{
			name: "kadfloor string set with whitespace",
			args: args{
				imp: &openrtb2.Imp{
					BidFloor: 0.12,
					Video:    &openrtb2.Video{},
					Ext:      json.RawMessage(`{"bidder":{"kadfloor":" \t  0.13  "}}`),
				},
			},
			expectedBidfloor: 0.13,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedWrapperExt, receivedPublisherId, err := parseImpressionObject(tt.args.imp, tt.args.extractWrapperExtFromImp, tt.args.extractPubIDFromImp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expectedWrapperExt, receivedWrapperExt)
			assert.Equal(t, tt.expectedPublisherId, receivedPublisherId)
			assert.Equal(t, tt.expectedBidfloor, tt.args.imp.BidFloor)
		})
	}
}

func TestExtractPubmaticExtFromRequest(t *testing.T) {
	type args struct {
		request *openrtb2.BidRequest
	}
	tests := []struct {
		name               string
		args               args
		expectedWrapperExt *pubmaticWrapperExt
		expectedAcat       []string
		wantErr            bool
	}{
		{
			name:    "Empty bidder param",
			wantErr: true,
		},
		{
			name: "Pubmatic wrapper ext missing/empty",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{}}}`),
				},
			},
			wantErr: false,
		},
		{
			name: "Only Pubmatic wrapper ext present",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":{"profile":123,"version":456}}}}`),
				},
			},
			expectedWrapperExt: &pubmaticWrapperExt{ProfileID: 123, VersionID: 456},
			wantErr:            false,
		},
		{
			name: "Invalid Pubmatic wrapper ext",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":{"profile":"123","version":456}}}}`),
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Pubmatic acat ext",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"acat":[" drg \t","dlu","ssr"],"wrapper":{"profile":123,"version":456}}}}`),
				},
			},
			expectedWrapperExt: &pubmaticWrapperExt{ProfileID: 123, VersionID: 456},
			expectedAcat:       []string{"drg", "dlu", "ssr"},
			wantErr:            false,
		},
		{
			name: "Invalid Pubmatic acat ext. We are ok with acat being non nil in this case as we are returning unmarshal error",
			args: args{
				request: &openrtb2.BidRequest{
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"acat":[1,3,4],"wrapper":{"profile":123,"version":456}}}}`),
				},
			},
			expectedWrapperExt: &pubmaticWrapperExt{ProfileID: 123, VersionID: 456},
			expectedAcat:       []string{"", "", ""},
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWrapperExt, gotAcat, err := extractPubmaticExtFromRequest(tt.args.request)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expectedWrapperExt, gotWrapperExt)
			assert.Equal(t, tt.expectedAcat, gotAcat)
		})
	}
}

func TestPubmaticAdapter_MakeRequests(t *testing.T) {
	type fields struct {
		URI string
	}
	type args struct {
		request *openrtb2.BidRequest
		reqInfo *adapters.ExtraRequestInfo
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		expectedReqData []*adapters.RequestData
		wantErr         bool
	}{
		// Happy paths covered by TestJsonSamples()
		// Covering only error scenarios here
		{
			name: "invalid bidderparams",
			args: args{
				request: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":"123"}}}`)},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PubmaticAdapter{
				URI: tt.fields.URI,
			}
			gotReqData, gotErr := a.MakeRequests(tt.args.request, tt.args.reqInfo)
			assert.Equal(t, tt.wantErr, len(gotErr) != 0)
			assert.Equal(t, tt.expectedReqData, gotReqData)
		})
	}
}

func Test_getAlternateBidderCodesFromRequest(t *testing.T) {
	type args struct {
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "request nil",
			want: []string{"pubmatic"},
		},
		{
			name: "request.ext nil",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: nil},
			},
			want: []string{"pubmatic"},
		},
		{
			name: "request.ext invalid json",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid": `)},
			},
			want:    []string{"pubmatic"},
			wantErr: true,
		},
		{
			name: "request.ext valid and alternatebiddercodes feature disabled",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":false,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":["groupm"]}}}}}`)},
			},
			want: []string{"pubmatic"},
		},
		{
			name: "request.ext valid and alternatebiddercodes disabled at bidder level",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":false,"allowedbiddercodes":["groupm"]}}}}}`)},
			},
			want: []string{"pubmatic"},
		},
		{
			name: "request.ext.valid with bidder code list not defined",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true}}}}}`)},
			},
			want: []string{"all"},
		},
		{
			name: "request.ext valid and wildcard bidder code",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":["*"]}}}}}`)},
			},
			want: []string{"all"},
		},
		{
			name: "request.ext.valid with empty bidder code list",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":[]}}}}}`)},
			},
			want: []string{"pubmatic"},
		},
		{
			name: "request.ext valid and groupm bidder allowed",
			args: args{
				bidRequest: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"alternatebiddercodes":{"enabled":true,"bidders":{"pubmatic":{"enabled":true,"allowedbiddercodes":["groupm"]}}}}}`)},
			},
			want: []string{"pubmatic", "groupm"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAlternateBidderCodesFromRequest(tt.args.bidRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAlternateBidderCodesFromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAlternateBidderCodesFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
