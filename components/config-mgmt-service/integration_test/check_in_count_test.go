package integration_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/chef/automate/components/config-mgmt-service/backend"
	iBackend "github.com/chef/automate/components/ingest-service/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckinCount(t *testing.T) {

	cases := []struct {
		now         time.Time
		description string
		daysAgo     int
		filter      map[string][]string
		nodeSets    []struct {
			node iBackend.Node
			runs []iBackend.Run
		}
		expectedResponse []backend.CheckInPeroid
	}{
		{
			description: "Zero nodes three days window",
			now:         parseTime(t, "2020-03-15T12:34:00Z"),
			daysAgo:     3,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 0,
					Start:        parseTime(t, "2020-03-12T13:00:00Z"),
					End:          parseTime(t, "2020-03-13T12:59:59Z"),
				},
				{
					CheckInCount: 0,
					Start:        parseTime(t, "2020-03-13T13:00:00Z"),
					End:          parseTime(t, "2020-03-14T12:59:59Z"),
				},
				{
					CheckInCount: 0,
					Start:        parseTime(t, "2020-03-14T13:00:00Z"),
					End:          parseTime(t, "2020-03-15T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{},
		},
		{
			description: "One node checks-in all three days",
			now:         parseTime(t, "2020-03-15T12:34:00Z"),
			daysAgo:     3,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-12T13:00:00Z"),
					End:          parseTime(t, "2020-03-13T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-13T13:00:00Z"),
					End:          parseTime(t, "2020-03-14T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-14T13:00:00Z"),
					End:          parseTime(t, "2020-03-15T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-12T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-12T13:01:00Z"),
							EndTime:   parseTime(t, "2020-03-12T13:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-14T12:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T12:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-14T13:05:00Z"),
							EndTime:   parseTime(t, "2020-03-14T13:06:59Z"),
						},
					},
				},
			},
		},
		{
			description: "Empty last bucket",
			now:         parseTime(t, "2020-03-15T12:34:00Z"),
			daysAgo:     3,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-12T13:00:00Z"),
					End:          parseTime(t, "2020-03-13T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-13T13:00:00Z"),
					End:          parseTime(t, "2020-03-14T12:59:59Z"),
				},
				{
					CheckInCount: 0,
					Start:        parseTime(t, "2020-03-14T13:00:00Z"),
					End:          parseTime(t, "2020-03-15T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-12T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-12T13:01:00Z"),
							EndTime:   parseTime(t, "2020-03-12T13:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-14T12:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T12:02:59Z"),
						},
					},
				},
			},
		},
		{
			description: "Empty start bucket",
			now:         parseTime(t, "2020-03-15T12:34:00Z"),
			daysAgo:     3,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 0,
					Start:        parseTime(t, "2020-03-12T13:00:00Z"),
					End:          parseTime(t, "2020-03-13T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-13T13:00:00Z"),
					End:          parseTime(t, "2020-03-14T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-14T13:00:00Z"),
					End:          parseTime(t, "2020-03-15T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-12T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-14T12:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T12:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-14T13:05:00Z"),
							EndTime:   parseTime(t, "2020-03-14T13:06:59Z"),
						},
					},
				},
			},
		},
		{
			description: "Empty center bucket",
			now:         parseTime(t, "2020-03-15T12:34:00Z"),
			daysAgo:     3,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-12T13:00:00Z"),
					End:          parseTime(t, "2020-03-13T12:59:59Z"),
				},
				{
					CheckInCount: 0,
					Start:        parseTime(t, "2020-03-13T13:00:00Z"),
					End:          parseTime(t, "2020-03-14T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-14T13:00:00Z"),
					End:          parseTime(t, "2020-03-15T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-12T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-12T13:01:00Z"),
							EndTime:   parseTime(t, "2020-03-12T13:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-14T13:05:00Z"),
							EndTime:   parseTime(t, "2020-03-14T13:06:59Z"),
						},
					},
				},
			},
		},
		{
			description: "One node checks-in twice in one day",
			now:         parseTime(t, "2020-03-15T01:14:00Z"),
			daysAgo:     1,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-14T02:00:00Z"),
					End:          parseTime(t, "2020-03-15T01:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-12T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-14T18:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T18:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-14T20:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T20:02:59Z"),
						},
					},
				},
			},
		},
		{
			description: "Two nodes check-in on the same day",
			now:         parseTime(t, "2020-03-15T11:59:00Z"),
			daysAgo:     1,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 2,
					Start:        parseTime(t, "2020-03-14T12:00:00Z"),
					End:          parseTime(t, "2020-03-15T11:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-12T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-14T20:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T20:02:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-10T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-14T23:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T23:02:59Z"),
						},
					},
				},
			},
		},
		{
			description: "3 days over daylight savings hour forword",
			now:         parseTime(t, "2020-03-09T12:34:00Z"),
			daysAgo:     3,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-06T13:00:00Z"),
					End:          parseTime(t, "2020-03-07T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-07T13:00:00Z"),
					End:          parseTime(t, "2020-03-08T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-08T13:00:00Z"),
					End:          parseTime(t, "2020-03-09T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-03T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-06T13:01:00Z"),
							EndTime:   parseTime(t, "2020-03-06T13:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-08T12:01:00Z"),
							EndTime:   parseTime(t, "2020-03-08T12:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2020-03-08T13:05:00Z"),
							EndTime:   parseTime(t, "2020-03-08T13:06:59Z"),
						},
					},
				},
			},
		},
		{
			description: "3 days over daylight savings hour back",
			now:         parseTime(t, "2019-11-04T12:34:00Z"),
			daysAgo:     3,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2019-11-01T13:00:00Z"),
					End:          parseTime(t, "2019-11-02T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2019-11-02T13:00:00Z"),
					End:          parseTime(t, "2019-11-03T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2019-11-03T13:00:00Z"),
					End:          parseTime(t, "2019-11-04T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2019-11-01T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2019-11-01T13:01:00Z"),
							EndTime:   parseTime(t, "2019-11-01T13:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2019-11-03T12:01:00Z"),
							EndTime:   parseTime(t, "2019-11-03T12:02:59Z"),
						},
						{
							StartTime: parseTime(t, "2019-11-03T13:05:00Z"),
							EndTime:   parseTime(t, "2019-11-03T13:06:59Z"),
						},
					},
				},
			},
		},
		{
			description: "filtering environment",
			now:         parseTime(t, "2020-03-15T12:34:00Z"),
			daysAgo:     3,
			filter: map[string][]string{
				"environment": {"forest"},
			},
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-12T13:00:00Z"),
					End:          parseTime(t, "2020-03-13T12:59:59Z"),
				},
				{
					CheckInCount: 0,
					Start:        parseTime(t, "2020-03-13T13:00:00Z"),
					End:          parseTime(t, "2020-03-14T12:59:59Z"),
				},
				{
					CheckInCount: 1,
					Start:        parseTime(t, "2020-03-14T13:00:00Z"),
					End:          parseTime(t, "2020-03-15T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-01T16:02:59Z"),
						NodeInfo: iBackend.NodeInfo{
							Environment: "forest",
						},
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-12T13:01:00Z"),
							EndTime:   parseTime(t, "2020-03-12T13:02:59Z"),
							NodeInfo: iBackend.NodeInfo{
								Environment: "forest",
							},
						},
						{
							StartTime: parseTime(t, "2020-03-14T12:01:00Z"),
							EndTime:   parseTime(t, "2020-03-14T12:02:59Z"),
							NodeInfo: iBackend.NodeInfo{
								Environment: "desert",
							},
						},
						{
							StartTime: parseTime(t, "2020-03-14T13:05:00Z"),
							EndTime:   parseTime(t, "2020-03-14T13:06:59Z"),
							NodeInfo: iBackend.NodeInfo{
								Environment: "forest",
							},
						},
					},
				},
			},
		},
		{
			description: "Over 10 nodes in one period",
			now:         parseTime(t, "2020-03-15T12:34:00Z"),
			daysAgo:     1,
			expectedResponse: []backend.CheckInPeroid{
				{
					CheckInCount: 11,
					Start:        parseTime(t, "2020-03-14T13:00:00Z"),
					End:          parseTime(t, "2020-03-15T12:59:59Z"),
				},
			},
			nodeSets: []struct {
				node iBackend.Node
				runs []iBackend.Run
			}{
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-01T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-14T13:05:00Z"),
							EndTime:   parseTime(t, "2020-03-14T13:06:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T01:05:00Z"),
							EndTime:   parseTime(t, "2020-03-15T01:06:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T02:05:00Z"),
							EndTime:   parseTime(t, "2020-03-15T02:06:59Z"),
						},
					},
				},

				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T02:15:00Z"),
							EndTime:   parseTime(t, "2020-03-15T02:16:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T02:25:00Z"),
							EndTime:   parseTime(t, "2020-03-15T02:26:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T02:35:00Z"),
							EndTime:   parseTime(t, "2020-03-15T02:36:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T02:45:00Z"),
							EndTime:   parseTime(t, "2020-03-15T02:46:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T02:55:00Z"),
							EndTime:   parseTime(t, "2020-03-15T02:56:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T03:05:00Z"),
							EndTime:   parseTime(t, "2020-03-15T03:06:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T03:15:00Z"),
							EndTime:   parseTime(t, "2020-03-15T03:16:59Z"),
						},
					},
				},
				{
					node: iBackend.Node{
						Checkin: parseTime(t, "2020-03-02T16:02:59Z"),
					},
					runs: []iBackend.Run{
						{
							StartTime: parseTime(t, "2020-03-15T03:25:00Z"),
							EndTime:   parseTime(t, "2020-03-15T03:26:59Z"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			nodes := make([]iBackend.Node, len(testCase.nodeSets))
			runs := make([]iBackend.Run, 0)
			// Adding required node data
			for index := range testCase.nodeSets {
				nodeID := newUUID()
				testCase.nodeSets[index].node.Exists = true
				testCase.nodeSets[index].node.NodeInfo.EntityUuid = nodeID
				testCase.nodeSets[index].node.NodeName = strconv.Itoa(index)
				nodes = append(nodes, testCase.nodeSets[index].node)
				for runIndex := range testCase.nodeSets[index].runs {
					runID := newUUID()
					testCase.nodeSets[index].runs[runIndex].EntityUuid = nodeID
					testCase.nodeSets[index].runs[runIndex].RunID = runID
					runs = append(runs, testCase.nodeSets[index].runs[runIndex])
				}
			}

			suite.IngestNodes(nodes)
			suite.IngestRuns(runs)
			defer suite.DeleteAllDocuments()

			endTime := time.Date(testCase.now.Year(), testCase.now.Month(), testCase.now.Day(),
				testCase.now.Hour(), 0, 0, 0, time.UTC).Add(time.Hour)

			startTime := endTime.Add(-time.Hour * 24 * time.Duration(testCase.daysAgo))

			actualResponse, err := esBackend.GetCheckinCountsTimeSeries(startTime,
				endTime.Add(-time.Millisecond), testCase.filter)
			require.NoError(t, err)

			assert.Equal(t, len(testCase.expectedResponse), len(actualResponse))
			for index := range actualResponse {
				assert.Equal(t, testCase.expectedResponse[index].Start.Format(time.RFC3339),
					actualResponse[index].Start.Format(time.RFC3339))
				assert.Equal(t, testCase.expectedResponse[index].End.Format(time.RFC3339),
					actualResponse[index].End.Format(time.RFC3339))
				assert.Equal(t, testCase.expectedResponse[index].CheckInCount,
					actualResponse[index].CheckInCount)
			}
		})
	}
}

func parseTime(t *testing.T, timeString string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, timeString)
	require.NoError(t, err)

	return parsedTime
}
