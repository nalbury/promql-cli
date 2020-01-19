/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func metricsTable(result model.LabelValues) {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	if !noHeaders {
		titleRow := "METRICS"
		fmt.Fprintln(w, titleRow)
	}
	for _, l := range result {
		row := strings.ToLower(string(l))
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

func metricsJson(result model.LabelValues) {
	if o, err := json.Marshal(result); err == nil {
		fmt.Println(string(o))
	}
}

func metricsCsv(result model.LabelValues) {
	w := csv.NewWriter(os.Stdout)
	var rows [][]string
	if !noHeaders {
		titleRow := []string{"metrics"}
		rows = append(rows, titleRow)
	}
	for _, l := range result {
		row := []string{string(l)}
		rows = append(rows, row)
	}
	w.WriteAll(rows)
}

func metricsQuery(host, output string) {
	client, err := api.NewClient(api.Config{
		Address: host,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.LabelValues(ctx, "__name__")
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	if output == "json" {
		metricsJson(result)
	} else if output == "csv" {
		metricsCsv(result)
	} else {
		metricsTable(result)
	}

}

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Get a list of all prometheus metric names",
	Long:  `Get a list of all prometheus metric names`,
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		output := viper.GetString("output")
		metricsQuery(host, output)
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
