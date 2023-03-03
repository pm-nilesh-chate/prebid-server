package prometheusmetrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRecordRejectedBids(t *testing.T) {
	type testIn struct {
		pubid, bidder, code string
	}
	type testOut struct {
		expCount int
	}
	testCases := []struct {
		description string
		in          testIn
		out         testOut
	}{
		{
			description: "record rejected bids",
			in: testIn{
				pubid:  "1010",
				bidder: "bidder",
				code:   "100",
			},
			out: testOut{
				expCount: 1,
			},
		},
	}
	for _, test := range testCases {
		pm := createMetricsForTesting()
		pm.RecordRejectedBids(test.in.pubid, test.in.bidder, test.in.code)

		assertCounterVecValue(t,
			"",
			"rejected_bids",
			pm.rejectedBids,
			float64(test.out.expCount),
			prometheus.Labels{
				pubIDLabel:  test.in.pubid,
				bidderLabel: test.in.bidder,
				codeLabel:   test.in.code,
			})
	}
}