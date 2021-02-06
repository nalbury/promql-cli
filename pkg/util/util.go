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

// util provides general utility functions for promql-cli
package util

import (
	"fmt"
	"github.com/prometheus/common/model"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// UniqLabels takes an interface model.Value and returns a slice of label names.
func UniqLabels(result model.Value) (labels []model.LabelName, err error) {
	labelKeys := make(map[model.LabelName]struct{})
	switch r := result.(type) {
	case model.Matrix:
		for _, v := range result.(model.Matrix) {
			for key := range v.Metric {
				labelKeys[key] = struct{}{}
			}
		}
	case model.Vector:
		for _, v := range result.(model.Vector) {
			for key := range v.Metric {
				labelKeys[key] = struct{}{}
			}
		}
	default:
		err = fmt.Errorf("unable to parse metric labels: unknown query result type: %T", r)
		return labels, err
	}

	for key := range labelKeys {
		labels = append(labels, key)
	}

	sort.Slice(labels, func(i, j int) bool {
		return string(labels[i]) < string(labels[j])
	})
	return labels, err
}

// TermDimensions stores the width and height of the current terminal window
// Used when setting the ascii graph size for range queries
type TermDimensions struct {
	Height int
	Width  int
}

// TerminalSize returns the current height and width [h,w]
// of the terminal promql is executed in.
func TerminalSize() (dimensions TermDimensions, err error) {
	var (
		stdout []byte
	)
	sttySize := exec.Command("stty", "size")
	sttySize.Stdin = os.Stdin
	stdout, err = sttySize.Output()
	if err != nil {
		return dimensions, err
	}
	o := strings.TrimSuffix(string(stdout), "\n")
	d := strings.Split(o, " ")
	dimensions.Height, err = strconv.Atoi(d[0])
	if err != nil {
		return dimensions, err
	}
	dimensions.Width, err = strconv.Atoi(d[1])
	if err != nil {
		return dimensions, err
	}
	return dimensions, nil
}
