package writer

import (
	"fmt"
	"testing"
	"time"

	"github.com/nalbury/promql-cli/pkg/util"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestRangeGraph(t *testing.T) {
	now := model.Now()
	cases := []struct {
		Result   RangeResult
		Expected string
	}{
		{
			Result: RangeResult{
				model.Matrix{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
						},
						Values: []model.SamplePair{
							{
								Timestamp: model.TimeFromUnix(now.Unix() - 60),
								Value:     1.0,
							},
							{
								Timestamp: now,
								Value:     1.0,
							},
						},
					},
				},
			},

			Expected: fmt.Sprintf(
				"\n##################################################\n# TIME_RANGE: %s -> %s #\n# METRIC: my_metric                              #\n##################################################\n 1.00 ┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── \n",
				model.TimeFromUnix(now.Unix()-60).Time().Format(time.Stamp),
				now.Time().Format(time.Stamp),
			),
		},
	}
	for i, c := range cases {
		dim := util.TermDimensions{
			Height: 49,
			Width:  178,
		}
		buf, err := c.Result.Graph(dim)
		assert.NoError(t, err, "Unexpected err for case %d, %v", i, err)
		assert.Equal(t, c.Expected, buf.String(), "Unexpected output for case %d", i)
	}
}

func TestJson(t *testing.T) {
	now := model.Now()
	cases := []struct {
		Result   Writer
		Expected string
	}{
		{
			Result: &RangeResult{
				model.Matrix{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
						},
						Values: []model.SamplePair{
							{
								Timestamp: model.TimeFromUnix(now.Unix() - 60),
								Value:     1.0,
							},
							{
								Timestamp: now,
								Value:     1.0,
							},
						},
					},
				},
			},
			Expected: fmt.Sprintf(
				"[{\"metric\":{\"__name__\":\"my_metric\"},\"values\":[[%s,\"1\"],[%s,\"1\"]]}]",
				model.TimeFromUnix(now.Unix()-60).String(),
				now.String(),
			),
		},
		{
			Result: &InstantResult{
				model.Vector{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
						},
						Value:     1.0,
						Timestamp: now,
					},
				},
			},
			Expected: fmt.Sprintf(
				"[{\"metric\":{\"__name__\":\"my_metric\"},\"value\":[%s,\"1\"]}]",
				now.String(),
			),
		},
		{
			Result: &MetricsResult{
				"my_metric",
				"my_other_metric",
			},
			Expected: "[\"my_metric\",\"my_other_metric\"]",
		},
		{
			Result: &LabelsResult{
				model.Vector{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
							"label":    "value",
						},
						Value:     1.0,
						Timestamp: now,
					},
				},
			},
			Expected: "[\"__name__\",\"label\"]",
		},
		{
			Result: &MetaResult{
				"my_metric": []v1.Metadata{
					{
						Type: "counter",
						Help: "The best metric you've ever recorded",
						Unit: "",
					},
				},
			},
			Expected: "{\"my_metric\":[{\"type\":\"counter\",\"help\":\"The best metric you've ever recorded\",\"unit\":\"\"}]}",
		},
	}
	for i, c := range cases {
		buf, err := c.Result.Json()
		assert.NoError(t, err, "Unexpected err for case %d, %v", i, err)
		assert.JSONEq(t, c.Expected, buf.String(), "Unexpected output for case %d", i)
	}
}

func TestCsv(t *testing.T) {
	now := model.Now()
	cases := []struct {
		Result   Writer
		Expected string
	}{
		{
			Result: &RangeResult{
				model.Matrix{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
						},
						Values: []model.SamplePair{
							{
								Timestamp: model.TimeFromUnix(now.Unix() - 60),
								Value:     1.0,
							},
							{
								Timestamp: now,
								Value:     1.0,
							},
						},
					},
				},
			},
			Expected: fmt.Sprintf(
				"__name__,value,timestamp\nmy_metric,1,%s\nmy_metric,1,%s\n",
				model.TimeFromUnix(now.Unix()-60).Time().Format(time.RFC3339),
				now.Time().Format(time.RFC3339),
			),
		},
		{
			Result: &InstantResult{
				model.Vector{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
						},
						Value:     1.0,
						Timestamp: now,
					},
				},
			},
			Expected: fmt.Sprintf(
				"__name__,value,timestamp\nmy_metric,1,%s\n",
				now.Time().Format(time.RFC3339),
			),
		},
		{
			Result: &MetricsResult{
				"my_metric",
				"my_other_metric",
			},
			Expected: "metrics\nmy_metric\nmy_other_metric\n",
		},
		{
			Result: &LabelsResult{
				model.Vector{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
							"label":    "value",
						},
						Value:     1.0,
						Timestamp: now,
					},
				},
			},
			Expected: "labels\n__name__\nlabel\n",
		},
		{
			Result: &MetaResult{
				"my_metric": []v1.Metadata{
					{
						Type: "counter",
						Help: "The best metric you've ever recorded",
						Unit: "",
					},
				},
			},
			Expected: "metric,type,help,unit\nmy_metric,counter,The best metric you've ever recorded,\n",
		},
	}
	for i, c := range cases {
		buf, err := c.Result.Csv(false)
		assert.NoError(t, err, "Unexpected err for case %d, %v", i, err)
		assert.Equal(t, c.Expected, buf.String(), "Unexpected output for case %d", i)
	}
}

func TestTable(t *testing.T) {
	now := model.Now()
	cases := []struct {
		Result   InstantWriter
		Expected string
	}{
		{
			Result: &InstantResult{
				model.Vector{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
						},
						Value:     1.0,
						Timestamp: now,
					},
				},
			},
			Expected: fmt.Sprintf(
				"__NAME__     VALUE    TIMESTAMP\nmy_metric    1        %s\n",
				now.Time().Format(time.RFC3339),
			),
		},
		{
			Result: &MetricsResult{
				"my_metric",
				"my_other_metric",
			},
			Expected: "METRICS\nmy_metric\nmy_other_metric\n",
		},
		{
			Result: &LabelsResult{
				model.Vector{
					{
						Metric: map[model.LabelName]model.LabelValue{
							"__name__": "my_metric",
							"label":    "value",
						},
						Value:     1.0,
						Timestamp: now,
					},
				},
			},
			Expected: "LABELS\n__name__\nlabel\n",
		},
		{
			Result: &MetaResult{
				"my_metric": []v1.Metadata{
					{
						Type: "counter",
						Help: "The best metric you've ever recorded",
						Unit: "",
					},
				},
			},
			Expected: "METRIC       TYPE       HELP                                    UNIT\nmy_metric    counter    The best metric you've ever recorded    \n",
		},
	}
	for i, c := range cases {
		buf, err := c.Result.Table(false)
		assert.NoError(t, err, "Unexpected err for case %d, %v", i, err)
		assert.Equal(t, c.Expected, buf.String(), "Unexpected output for case %d", i)
	}
}
