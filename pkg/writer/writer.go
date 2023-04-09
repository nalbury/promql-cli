/*
Copyright © 2020 Nick Albury nickalbury@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// writer provides our stdout writers for promql query results
package writer

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/nalbury/promql-cli/pkg/util"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Writer is our base interface for promql writers
// Defines Json and Csv writers
type Writer interface {
	Json() (bytes.Buffer, error)
	Csv(noHeaders bool) (bytes.Buffer, error)
}

// RangeWriter extends the Writer interface by adding a Graph method
// Used specifically for writing the results of range queries
type RangeWriter interface {
	Writer
	Graph(dim util.TermDimensions) (bytes.Buffer, error)
}

// InstantWriter extends the Writer interface by adding a Table method
// Use specifically for writing the results of instant queries
type InstantWriter interface {
	Writer
	Table(noHeaders bool) (bytes.Buffer, error)
}

// RangeResult is wrapper of the prometheus model.Matrix type returned from range queries
// Satisfies the RangeWriter interface
type RangeResult struct {
	model.Matrix
}

// findMaxAbsFloat64 returns the maximum absolute value in a slice of float64
func findMaxAbsFloat64(values []float64) (float64, bool) {
	if len(values) == 0 {
		return 0, false
	}

	max := values[0]
	for _, number := range values {
		if math.Abs(number) > max {
			max = number
		}
	}
	return max, true
}

// Graph returns an ascii graph using https://github.com/guptarohit/asciigraph
func (r *RangeResult) Graph(dim util.TermDimensions) (bytes.Buffer, error) {
	var buf bytes.Buffer

	termHeightOpt := asciigraph.Height(dim.Height / 5)

	for _, m := range r.Matrix {
		var (
			data         []float64
			start        string
			end          string
			borderLength int
		)

		for _, v := range m.Values {
			data = append(data, float64(v.Value))
		}

		// Find max absolute value to determine width of the left column of the graph
		//
		// We fall back to 0 if there are no data points
		maxAbs, found := findMaxAbsFloat64(data)
		if !found {
			maxAbs = 0
		}

		// Determine width of the left column of the graph
		maxValueLen := len(strconv.Itoa(int(math.Ceil(maxAbs))))

		// "4" is a magic number to compensate for the asciigraph decorations
		termWidthOpt := asciigraph.Width(dim.Width - 4 - maxValueLen)

		start = m.Values[0].Timestamp.Time().Format(time.Stamp)
		end = m.Values[(len(m.Values) - 1)].Timestamp.Time().Format(time.Stamp)

		timeRange := start + " -> " + end

		// Generate the graph boxed to our terminal size
		graph := asciigraph.Plot(data, termHeightOpt, termWidthOpt)

		// Create our header for each graph
		// # TIME_RANGE: Sep 27 09:08:09 -> Sep 27 09:18:09
		timeRangeHeader := "# TIME_RANGE: " + timeRange
		// # METRIC: {instance="10.202.38.101:6443"}
		metricHeader := "# METRIC: " + m.Metric.String()
		// Truncate the metric header to the term width - 2
		// This ensures that long metric headers don't overflow onto a new line
		if len(metricHeader) > (dim.Width - 2) {
			metricHeader = metricHeader[:(dim.Width - 2)]
		}
		// Determine the longest header string and set the border (######) to it's length + 2
		// Add spacing to the shortest header
		if len(timeRangeHeader) > len(metricHeader) {
			borderLength = len(timeRangeHeader) + 2
			s := len(timeRangeHeader) - len(metricHeader)
			metricHeader = metricHeader + strings.Repeat(" ", s)
		} else {
			borderLength = len(metricHeader) + 2
			s := len(metricHeader) - len(timeRangeHeader)
			timeRangeHeader = timeRangeHeader + strings.Repeat(" ", s)
		}
		// Create the border of '#'
		border := strings.Repeat("#", borderLength)
		// Write out
		if _, err := fmt.Fprintf(&buf, "\n%s\n", border); err != nil {
			return buf, err
		}
		if _, err := fmt.Fprintf(&buf, "%s #\n", timeRangeHeader); err != nil {
			return buf, err
		}
		if _, err := fmt.Fprintf(&buf, "%s #\n", metricHeader); err != nil {
			return buf, err
		}
		if _, err := fmt.Fprintf(&buf, "%s\n", border); err != nil {
			return buf, err
		}
		if _, err := fmt.Fprintf(&buf, "%s\n", graph); err != nil {
			return buf, err
		}
	}
	return buf, nil
}

// Json returns the response from a range query as json
func (r *RangeResult) Json() (bytes.Buffer, error) {
	var buf bytes.Buffer
	o, err := json.Marshal(r.Matrix)
	if err != nil {
		return buf, err
	}
	buf.Write(o)
	return buf, nil
}

// Csv returns the response from a range query as a csv
func (r *RangeResult) Csv(noHeaders bool) (bytes.Buffer, error) {
	var (
		buf  bytes.Buffer
		rows [][]string
	)
	w := csv.NewWriter(&buf)
	labels, err := util.UniqLabels(r.Matrix)
	if err != nil {
		return buf, err
	}
	if !noHeaders {
		var titleRow []string
		for _, k := range labels {
			titleRow = append(titleRow, string(k))
		}

		titleRow = append(titleRow, "value")
		titleRow = append(titleRow, "timestamp")

		rows = append(rows, titleRow)
	}

	for _, m := range r.Matrix {
		for _, v := range m.Values {
			row := make([]string, len(labels))
			for i, key := range labels {
				row[i] = string(m.Metric[key])
			}
			row = append(row, v.Value.String())
			row = append(row, v.Timestamp.Time().Format(time.RFC3339))
			rows = append(rows, row)
		}
	}
	if err := w.WriteAll(rows); err != nil {
		return buf, err
	}
	return buf, nil
}

// WriteRange writes out the results of the query to an
// output buffer and prints it to stdout
func WriteRange(r RangeWriter, format string, noHeaders bool) error {
	var (
		buf bytes.Buffer
		err error
	)
	switch format {
	case "json":
		buf, err = r.Json()
		if err != nil {
			return err
		}
	case "csv":
		buf, err = r.Csv(noHeaders)
		if err != nil {
			return err
		}
	default:
		dim, err := util.TerminalSize()
		if err != nil {
			return err
		}
		buf, err = r.Graph(dim)
		if err != nil {
			return err
		}
	}
	fmt.Println(buf.String())
	return nil
}

// InstantResult is wrapper of the prometheus model.Matrix type returned from instant queries
// Satisfies the InstantWriter interface
type InstantResult struct {
	model.Vector
}

// Table returns the response from an instant query as a tab separated table
func (r *InstantResult) Table(noHeaders bool) (bytes.Buffer, error) {
	var buf bytes.Buffer
	const padding = 4
	w := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', 0)
	labels, err := util.UniqLabels(r.Vector)
	if err != nil {
		return buf, err
	}
	if !noHeaders {
		var titles []string
		for _, k := range labels {
			titles = append(titles, strings.ToUpper(string(k)))
		}
		titles = append(titles, "VALUE")
		titles = append(titles, "TIMESTAMP")
		titleRow := strings.Join(titles, "\t")
		if _, err := fmt.Fprintln(w, titleRow); err != nil {
			return buf, err
		}
	}

	for _, v := range r.Vector {
		data := make([]string, len(labels))
		for i, key := range labels {
			data[i] = string(v.Metric[key])
		}
		data = append(data, v.Value.String())
		data = append(data, v.Timestamp.Time().Format(time.RFC3339))
		row := strings.Join(data, "\t")
		if _, err := fmt.Fprintln(w, row); err != nil {
			return buf, err
		}
	}
	if err := w.Flush(); err != nil {
		return buf, err
	}
	return buf, nil
}

// Json returns the response from an instant query as json
func (r *InstantResult) Json() (bytes.Buffer, error) {
	var buf bytes.Buffer
	o, err := json.Marshal(r.Vector)
	if err != nil {
		return buf, err
	}
	buf.Write(o)
	return buf, nil
}

// Csv returns the response from an instant query as a csv
func (r *InstantResult) Csv(noHeaders bool) (bytes.Buffer, error) {
	var (
		buf  bytes.Buffer
		rows [][]string
	)
	w := csv.NewWriter(&buf)
	labels, err := util.UniqLabels(r.Vector)
	if err != nil {
		return buf, err
	}
	if !noHeaders {
		var titleRow []string
		for _, k := range labels {
			titleRow = append(titleRow, string(k))
		}

		titleRow = append(titleRow, "value")
		titleRow = append(titleRow, "timestamp")

		rows = append(rows, titleRow)
	}

	for _, v := range r.Vector {
		row := make([]string, len(labels))
		for i, key := range labels {
			row[i] = string(v.Metric[key])
		}
		row = append(row, v.Value.String())
		row = append(row, v.Timestamp.Time().Format(time.RFC3339))
		rows = append(rows, row)
	}
	if err := w.WriteAll(rows); err != nil {
		return buf, err
	}
	return buf, nil
}

// MetricsResult is the list of metrics names from a metadata query result
// It satisfies the InstantWriter interface as it's
// a point in time (e.g. what metrics are currently queryable)
type MetricsResult []string

// Table returns the response from a metrics query as a single column table
func (r *MetricsResult) Table(noHeaders bool) (bytes.Buffer, error) {
	var buf bytes.Buffer
	const padding = 4
	w := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', 0)
	if !noHeaders {
		titleRow := "METRICS"
		if _, err := fmt.Fprintln(w, titleRow); err != nil {
			return buf, err
		}
	}
	for _, l := range *r {
		row := string(l)
		if _, err := fmt.Fprintln(w, row); err != nil {
			return buf, err
		}
	}
	if err := w.Flush(); err != nil {
		return buf, err
	}
	return buf, nil
}

// Json returns the response from a metrics query as json
func (r *MetricsResult) Json() (bytes.Buffer, error) {
	var buf bytes.Buffer
	o, err := json.Marshal(r)
	if err != nil {
		return buf, err
	}
	buf.Write(o)
	return buf, nil
}

// Csv returns the response from a metrics query as a single column csv
func (r *MetricsResult) Csv(noHeaders bool) (bytes.Buffer, error) {
	var (
		buf  bytes.Buffer
		rows [][]string
	)
	w := csv.NewWriter(&buf)
	if !noHeaders {
		titleRow := []string{"metrics"}
		rows = append(rows, titleRow)
	}
	for _, l := range *r {
		row := []string{string(l)}
		rows = append(rows, row)
	}
	if err := w.WriteAll(rows); err != nil {
		return buf, err
	}
	return buf, nil
}

// LabelsResult is the result of an instant query
// It's really the same as an InstantResult with different methods for parsing
// the unique labels present in a query result.
type LabelsResult struct {
	model.Vector
}

// Table returns the labels from an instant query as a single column table
func (r *LabelsResult) Table(noHeaders bool) (bytes.Buffer, error) {
	var buf bytes.Buffer
	const padding = 4
	w := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', 0)
	labels, err := util.UniqLabels(r.Vector)
	if err != nil {
		return buf, err
	}
	if !noHeaders {
		titleRow := "LABELS"
		if _, err := fmt.Fprintln(w, titleRow); err != nil {
			return buf, err
		}
	}
	for _, l := range labels {
		row := string(l)
		if _, err := fmt.Fprintln(w, row); err != nil {
			return buf, err
		}
	}
	if err := w.Flush(); err != nil {
		return buf, err
	}
	return buf, nil
}

// Json returns the labels from an instant query as json
func (r *LabelsResult) Json() (bytes.Buffer, error) {
	var buf bytes.Buffer
	labels, err := util.UniqLabels(r.Vector)
	if err != nil {
		return buf, err
	}
	o, err := json.Marshal(labels)
	if err != nil {
		return buf, err
	}
	buf.Write(o)
	return buf, nil
}

// Csv returns the labels from an instant query as a single column csv
func (r *LabelsResult) Csv(noHeaders bool) (bytes.Buffer, error) {
	var (
		buf  bytes.Buffer
		rows [][]string
	)
	w := csv.NewWriter(&buf)
	labels, err := util.UniqLabels(r.Vector)
	if err != nil {
		return buf, err
	}
	if !noHeaders {
		titleRow := []string{"labels"}
		rows = append(rows, titleRow)
	}
	for _, l := range labels {
		row := []string{string(l)}
		rows = append(rows, row)
	}
	if err := w.WriteAll(rows); err != nil {
		return buf, err
	}
	return buf, nil
}

// MetaResult is the result of our metadata query
// It satisfies the InstantWriter interface
type MetaResult map[string][]v1.Metadata

// Metrics returns an array of metrics from a metadata query result
func (r *MetaResult) Metrics() MetricsResult {
	var metrics MetricsResult = make([]string, 0, len(*r))
	for k := range *r {
		metrics = append(metrics, k)
	}
	return metrics
}

// Json returns the result from a metadata query as json
func (r *MetaResult) Json() (bytes.Buffer, error) {
	var buf bytes.Buffer
	j, err := json.Marshal(r)
	if err != nil {
		return buf, err
	}
	buf.Write(j)
	return buf, nil
}

// Csv returns the result from a metadata query as csv
func (r *MetaResult) Csv(noHeaders bool) (bytes.Buffer, error) {
	var (
		buf  bytes.Buffer
		rows [][]string
	)
	w := csv.NewWriter(&buf)
	if !noHeaders {
		titleRow := []string{"metric", "type", "help", "unit"}
		rows = append(rows, titleRow)
	}

	for metric, meta := range *r {
		for _, m := range meta {
			data := make([]string, 0, 4)
			data = append(data, metric)
			data = append(data, string(m.Type))
			data = append(data, m.Help)
			data = append(data, m.Unit)
			rows = append(rows, data)
		}
	}

	if err := w.WriteAll(rows); err != nil {
		return buf, err
	}
	return buf, nil
}

// Table returns the result from a metadata query as tab separated table
func (r *MetaResult) Table(noHeaders bool) (bytes.Buffer, error) {
	var buf bytes.Buffer
	const padding = 4
	w := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', 0)
	if !noHeaders {
		titles := []string{"METRIC", "TYPE", "HELP", "UNIT"}
		titleRow := strings.Join(titles, "\t")
		if _, err := fmt.Fprintln(w, titleRow); err != nil {
			return buf, err
		}
	}
	for metric, meta := range *r {
		for _, m := range meta {
			data := make([]string, 0, 4)
			data = append(data, metric)
			data = append(data, string(m.Type))
			data = append(data, m.Help)
			data = append(data, m.Unit)
			row := strings.Join(data, "\t")
			if _, err := fmt.Fprintln(w, row); err != nil {
				return buf, err
			}
		}
	}
	if err := w.Flush(); err != nil {
		return buf, err
	}
	return buf, nil
}

// SeriesResult currently doesn't write anything itself but helps create other result types like metrics from the series api
type SeriesResult []model.LabelSet

// Metrics creates a MetricsResult from a SeriesResult
func (r *SeriesResult) Metrics() MetricsResult {
	u := make(map[string]struct{})
	var m MetricsResult
	for _, l := range *r {
		name := string(l["__name__"])
		if _, ok := u[name]; !ok {
			u[name] = struct{}{}
			m = append(m, name)
		}
	}
	return m
}

// WriteInstant writes out the results of the query to an
// output buffer and prints it to stdout
func WriteInstant(i InstantWriter, format string, noHeaders bool) error {
	var (
		buf bytes.Buffer
		err error
	)
	switch format {
	case "json":
		buf, err = i.Json()
		if err != nil {
			return err
		}
	case "csv":
		buf, err = i.Csv(noHeaders)
		if err != nil {
			return err
		}
	default:
		buf, err = i.Table(noHeaders)
		if err != nil {
			return err
		}
	}
	fmt.Println(buf.String())
	return nil
}
