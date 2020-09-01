// General utilities
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
			for key, _ := range v.Metric {
				labelKeys[key] = struct{}{}
			}
		}
	case model.Vector:
		for _, v := range result.(model.Vector) {
			for key, _ := range v.Metric {
				labelKeys[key] = struct{}{}
			}
		}
	default:
		err = fmt.Errorf("Unable to parse metric labels: unknown query result type: %T", r)
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
