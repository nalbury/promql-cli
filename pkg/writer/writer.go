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
	"github.com/guptarohit/asciigraph"
	"github.com/nalbury/promql-cli/pkg/util"
	"github.com/prometheus/common/model"
	"strings"
	"text/tabwriter"
	"time"
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
	Graph() (bytes.Buffer, error)
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

func (r *RangeResult) Graph() (bytes.Buffer, error) {
	var buf bytes.Buffer

	dim, err := util.TerminalSize()
	if err != nil {
		return buf, err
	}
	termHeightOpt := asciigraph.Height(dim.Height / 3)
	termWidthOpt := asciigraph.Width(dim.Width - 8)

	for _, m := range r.Matrix {
		var data []float64

		for _, v := range m.Values {
			data = append(data, float64(v.Value))
		}

		graph := asciigraph.Plot(data, termHeightOpt, termWidthOpt)
		fmt.Fprintf(&buf, "\nMetric: %s\n", m.Metric.String())
		fmt.Fprintf(&buf, "%v\n", graph)
	}
	return buf, nil
}

func (r *RangeResult) Json() (bytes.Buffer, error) {
	var buf bytes.Buffer
	o, err := json.Marshal(r.Matrix)
	if err != nil {
		return buf, err
	}
	buf.Write(o)
	return buf, nil
}

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
	w.WriteAll(rows)
	return buf, nil
}

func WriteRange(r RangeWriter, format string, noHeaders bool) error {
	var (
		o   bytes.Buffer
		err error
	)
	switch format {
	case "json":
		o, err = r.Json()
		if err != nil {
			return err
		}
	case "csv":
		o, err = r.Csv(noHeaders)
		if err != nil {
			return err
		}
	default:
		o, err = r.Graph()
		if err != nil {
			return err
		}
	}
	fmt.Println(o.String())
	return nil
}

// InstantResult is wrapper of the prometheus model.Matrix type returned from instant queries
// Satisfies the InstantWriter interface
type InstantResult struct {
	model.Vector
}

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
		fmt.Fprintln(w, titleRow)
	}

	for _, v := range r.Vector {
		data := make([]string, len(labels))
		for i, key := range labels {
			data[i] = string(v.Metric[key])
		}
		data = append(data, v.Value.String())
		data = append(data, v.Timestamp.Time().Format(time.RFC3339))
		row := strings.Join(data, "\t")
		fmt.Fprintln(w, row)
	}
	w.Flush()
	return buf, nil
}

func (r *InstantResult) Json() (bytes.Buffer, error) {
	var buf bytes.Buffer
	o, err := json.Marshal(r.Vector)
	if err != nil {
		return buf, err
	}
	buf.Write(o)
	return buf, nil
}

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
	w.WriteAll(rows)
	return buf, nil
}

func WriteInstant(i InstantWriter, format string, noHeaders bool) error {
	var (
		o   bytes.Buffer
		err error
	)
	switch format {
	case "json":
		o, err = i.Json()
		if err != nil {
			return err
		}
	case "csv":
		o, err = i.Csv(noHeaders)
		if err != nil {
			return err
		}
	default:
		o, err = i.Table(noHeaders)
		if err != nil {
			return err
		}
	}
	fmt.Println(o.String())
	return nil
}
