/*
Copyright © 2019 Nick Albury nickalbury@gmail.com

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
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/guptarohit/asciigraph"
)

var (
	cfgFile   string
	start     string
	end       string
	noHeaders bool
)

func getLabels(result model.Value) []model.LabelName {
	// first we need to create our list of columns
	// We use a map to createa uniq set of keys without searching the existing list everytime
	// An empty struct is used as the value to minimize the mem of the map
	labelKeys := make(map[model.LabelName]struct{})
	if result.Type() == model.ValVector {
		values := result.(model.Vector)
		for _, v := range values {
			for key, _ := range v.Metric {
				labelKeys[key] = struct{}{}
			}
		}
	} else if result.Type() == model.ValMatrix {
		values := result.(model.Matrix)
		for _, v := range values {
			for key, _ := range v.Metric {
				labelKeys[key] = struct{}{}
			}
		}
	} else {
		fmt.Println("Can't determine result type")
	}
	// Once we have our map, we can create a slice of label names
	var labelKeySlice []model.LabelName
	for key := range labelKeys {
		labelKeySlice = append(labelKeySlice, key)
	}

	//Finally we sort the slice to ensure consistency between each query
	sort.Slice(labelKeySlice, func(i, j int) bool {
		return string(labelKeySlice[i]) < string(labelKeySlice[j])
	})
	return labelKeySlice
}

func instantTable(result model.Value) {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	labelKeySlice := getLabels(result)
	if !noHeaders {
		var titles []string
		for _, k := range labelKeySlice {
			titles = append(titles, strings.ToUpper(string(k)))
		}

		titles = append(titles, "VALUE")
		titles = append(titles, "TIMESTAMP")
		titleRow := strings.Join(titles, "\t")
		fmt.Fprintln(w, titleRow)
	}
	vectorVal := result.(model.Vector)
	for _, v := range vectorVal {
		data := make([]string, len(labelKeySlice))
		for i, key := range labelKeySlice {
			data[i] = string(v.Metric[key])
		}
		data = append(data, v.Value.String())
		data = append(data, v.Timestamp.String())
		row := strings.Join(data, "\t")
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

func instantJson(result model.Value) {
	vectorVal := result.(model.Vector)
	if o, err := json.Marshal(vectorVal); err == nil {
		fmt.Println(string(o))
	}
}

func instantCsv(result model.Value) {
	vectorVal := result.(model.Vector)
	w := csv.NewWriter(os.Stdout)

	var rows [][]string
	labelKeySlice := getLabels(result)

	// And a slice of label names as a string for our title row
	if !noHeaders {
		var titleRow []string
		for _, k := range labelKeySlice {
			titleRow = append(titleRow, string(k))
		}

		titleRow = append(titleRow, "value")
		titleRow = append(titleRow, "timestamp")

		rows = append(rows, titleRow)
	}

	for _, v := range vectorVal {
		row := make([]string, len(labelKeySlice))
		for i, key := range labelKeySlice {
			row[i] = string(v.Metric[key])
		}
		row = append(row, v.Value.String())
		row = append(row, v.Timestamp.Time().Format(time.RFC3339))
		rows = append(rows, row)
	}
	w.WriteAll(rows)
}

func instantQuery(host, queryString, output string) {
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
	result, warnings, err := v1api.Query(ctx, queryString, time.Now())
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	if result.Type() == model.ValVector {
		if output == "json" {
			instantJson(result)
		} else if output == "csv" {
			instantCsv(result)
		} else {
			instantTable(result)
		}
	} else {
		fmt.Printf("Did not receive an instant vector")
	}
}

func rangeGraph(result model.Value) {
	matrixVal := result.(model.Matrix)
	for _, m := range matrixVal {
		var data []float64
		for _, v := range m.Values {
			data = append(data, float64(v.Value))
		}
		fmt.Println("")
		fmt.Println("Metric:", m.Metric.String())
		graph := asciigraph.Plot(data)
		fmt.Println(graph)
		fmt.Println("")
	}
}

func rangeJson(result model.Value) {
	matrixVal := result.(model.Matrix)
	if o, err := json.Marshal(matrixVal); err == nil {
		fmt.Println(string(o))
	}
}

func rangeCsv(result model.Value) {
	w := csv.NewWriter(os.Stdout)
	var rows [][]string
	labelKeySlice := getLabels(result)
	if !noHeaders {
		var titleRow []string
		for _, k := range labelKeySlice {
			titleRow = append(titleRow, string(k))
		}

		titleRow = append(titleRow, "value")
		titleRow = append(titleRow, "timestamp")

		rows = append(rows, titleRow)
	}
	matrixVal := result.(model.Matrix)
	for _, m := range matrixVal {
		for _, v := range m.Values {
			row := make([]string, len(labelKeySlice))
			for i, key := range labelKeySlice {
				row[i] = string(m.Metric[key])
			}
			row = append(row, v.Value.String())
			row = append(row, v.Timestamp.Time().Format(time.RFC3339))
			rows = append(rows, row)
		}
	}
	w.WriteAll(rows)
}

func getRange(step, start, end string) v1.Range {
	var rangeStart time.Time
	var rangeEnd time.Time
	var rangeStep time.Duration

	if step != "" {
		if s, err := time.ParseDuration(step); err == nil {
			rangeStep = s
		} else {
			fmt.Println(err)
		}
	} else {
		rangeStep = time.Minute
	}

	if s, err := time.Parse(time.RFC3339, start); err == nil {
		rangeStart = s
	} else if l, err := time.ParseDuration(start); err == nil {
		rangeStart = time.Now().Add(-l)
	} else {
		fmt.Printf("No valid range start provided: %e", err)
	}

	if e, err := time.Parse(time.RFC3339, end); err == nil {
		rangeEnd = e
	} else {
		rangeEnd = time.Now()
	}

	r := v1.Range{
		Start: rangeStart,
		End:   rangeEnd,
		Step:  rangeStep,
	}
	return r
}

func rangeQuery(host, queryString, output string, r v1.Range) {
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
	result, warnings, err := v1api.QueryRange(ctx, queryString, r)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	if result.Type() == model.ValMatrix {
		if output == "json" {
			rangeJson(result)
		} else if output == "csv" {
			rangeCsv(result)
		} else {
			rangeGraph(result)
		}
	} else {
		fmt.Printf("Did not receive an instant vector")
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: "0.1.0",
	Use:     "promql [query_string]",
	Short:   "Query prometheus from the command line",
	Long:    `Query prometheus from the command line for quick analysis`,
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		host := viper.GetString("host")
		step := viper.GetString("step")
		output := viper.GetString("output")
		if start != "" {
			r := getRange(step, start, end)
			rangeQuery(host, query, output, r)
		} else {
			instantQuery(host, query, output)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file location (default $HOME/.promql-cli.yaml)")
	rootCmd.PersistentFlags().String("host", "http://0.0.0.0:9090", "prometheus server url")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	rootCmd.PersistentFlags().String("step", "1m", "Results step duration (h,m,s e.g. 1m)")
	viper.BindPFlag("step", rootCmd.PersistentFlags().Lookup("step"))
	rootCmd.PersistentFlags().StringVar(&start, "start", "", "Query range start duration (either as a lookback in h,m,s e.g. 1m, or as an RFC3339 formatted date string). Required for range queries")
	rootCmd.PersistentFlags().StringVar(&end, "end", "now", "Query range end (either 'now', or an RFC3339 formatted date string)")
	rootCmd.PersistentFlags().String("output", "", "Override the default output format (graph for range queries, and table for instant queries). Options: json,csv")
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	rootCmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "Disable table headers for instant queries")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".promql-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".promql-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}
}
