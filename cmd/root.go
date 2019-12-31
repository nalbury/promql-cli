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
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
	host      string
	start     string
	end       string
	step      string
	noHeaders bool
)

func instantTable(result model.Value) {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	if !noHeaders {
		titleAttributes := []string{"METRIC", "VALUE"}
		titleRow := strings.Join(titleAttributes, "\t")
		fmt.Fprintln(w, titleRow)
	}
	vectorVal := result.(model.Vector)
	for _, v := range vectorVal {
		columns := []string{v.Metric.String(), v.Value.String()}
		row := strings.Join(columns, "\t")
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

func instantQuery(queryString string) {
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
		instantTable(result)
	} else {
		fmt.Printf("Did not receive and instant vector")
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

func rangeQuery(queryString string, r v1.Range) {
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
	rangeGraph(result)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: "0.1.0",
	Use:     "promql [query_string]",
	Short:   "Query prometheus from the command line",
	Long:    `Query prometheus from the command line for quick analysis`,
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		if start != "" {
			r := getRange(step, start, end)
			rangeQuery(query, r)
		} else {
			instantQuery(query)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "$HOME/.promql-cli.yaml", "config file")
	rootCmd.PersistentFlags().StringVar(&host, "host", "http://0.0.0.0:9090", "prometheus server url")
	rootCmd.PersistentFlags().StringVar(&step, "step", "1m", "Results step duration (h,m,s e.g. 1m)")
	rootCmd.PersistentFlags().StringVar(&start, "start", "", "Query range start duration (either as a lookback in h,m,s e.g. 1m, or as an RFC3339 formatted date string). Required for range queries")
	rootCmd.PersistentFlags().StringVar(&end, "end", "now", "Query range end (either 'now', or an RFC3339 formatted date string)")
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
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
