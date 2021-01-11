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
package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/nalbury/promql-cli/pkg/promql"
	"github.com/nalbury/promql-cli/pkg/writer"
)

// cmd line args
var (
	cfgFile   string
	start     string
	end       string
	noHeaders bool
)

// global error logger
var errlog *log.Logger = log.New(os.Stderr, "", 0)

// instantQuery performs an instant query and writes the results to stdout
func instantQuery(host, queryString, output string, timeout time.Duration) {
	client, err := promql.CreateClient(host)
	if err != nil {
		errlog.Fatalf("Error creating client, %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, warnings, err := client.Query(ctx, queryString, time.Now())
	if err != nil {
		errlog.Fatalf("Error querying Prometheus, %v\n", err)
	}
	if len(warnings) > 0 {
		errlog.Printf("Warnings: %v\n", warnings)
	}

	// if result is the expected type, Write it out in the
	// desired output format
	if result, ok := result.(model.Vector); ok {
		r := writer.InstantResult{result}
		if err := writer.WriteInstant(&r, output, noHeaders); err != nil {
			errlog.Println(err)
		}
	} else {
		errlog.Println("Did not receive an instant vector")
	}
}

// getRange creates a prometheus range from the provided start, end, and step options
func getRange(step, start, end string) (r v1.Range, err error) {
	// At minimum we need a start time so we attempt to parse that first
	if s, err := time.Parse(time.RFC3339, start); err == nil {
		r.Start = s
	} else if l, err := time.ParseDuration(start); err == nil {
		r.Start = time.Now().Add(-l)
	} else {
		err = fmt.Errorf("Unable to parse range start time, %v", err)
		return r, err
	}

	// Set up defaults for the range end and step values
	r.End = time.Now()
	r.Step = time.Minute

	// If the user provided a step value, parse it as a time.Duration and override the default
	if step != "" {
		r.Step, err = time.ParseDuration(step)
		if err != nil {
			err = fmt.Errorf("Unable to parse step duration, %v", err)
			return r, err
		}
	}

	// If the user provided an end value, parse it to a time struct and override the default
	if end != "now" {
		e, err := time.Parse(time.RFC3339, end)
		if err != nil {
			err = fmt.Errorf("Unable to parse range end time, %v", err)
			return r, err
		}
		r.End = e
	}

	return r, err
}

// rangeQuery performs a range query and writes the results to stdout
func rangeQuery(host, queryString, output string, timeout time.Duration, r v1.Range) {
	// Create client
	client, err := promql.CreateClient(host)
	if err != nil {
		errlog.Fatalf("Error creating client, %v\n", err)
	}

	// create context with a timeout,
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// execute query
	result, warnings, err := client.QueryRange(ctx, queryString, r)
	if err != nil {
		errlog.Fatalf("Error querying Prometheus, %v\n", err)
	}
	// print warnings to stderr and continue
	if len(warnings) > 0 {
		errlog.Printf("Warnings: %v\n", warnings)
	}

	// if result is the expected type, Write it out in the
	// desired output format
	if result, ok := result.(model.Matrix); ok {
		r := writer.RangeResult{result}
		if err := writer.WriteRange(&r, output, noHeaders); err != nil {
			errlog.Println(err)
		}
	} else {
		errlog.Println("Did not receive a range result")
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: "v0.2.1",
	Use:     "promql [query_string]",
	Short:   "Query prometheus from the command line",
	Long:    `Query prometheus from the command line for quick analysis.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		host := viper.GetString("host")
		step := viper.GetString("step")
		output := viper.GetString("output")
		timeout := viper.GetInt("timeout")

		// Convert our timeout flag into a time.Duration
		t := time.Duration(int64(timeout)) * time.Second
		// If we have a start time for the query, assume we're doing a range query
		if start != "" {
			// Parse the time range from the cmd line/config file options
			r, err := getRange(step, start, end)
			if err != nil {
				errlog.Fatalln(err)
			}
			// Execute our range query
			rangeQuery(host, query, output, t, r)
		} else {
			instantQuery(host, query, output, t)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		errlog.Fatalln(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file location (default $HOME/.promql-cli.yaml)")
	rootCmd.PersistentFlags().String("host", "http://0.0.0.0:9090", "prometheus server url")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	rootCmd.PersistentFlags().String("step", "1m", "Results step duration (h,m,s e.g. 1m)")
	viper.BindPFlag("step", rootCmd.PersistentFlags().Lookup("step"))
	rootCmd.PersistentFlags().StringVar(&start, "start", "", "Query range start duration (either as a lookback in h,m,s e.g. 1m, or as an ISO 8601 formatted date string). Required for range queries")
	rootCmd.PersistentFlags().StringVar(&end, "end", "now", "Query range end (either 'now', or an ISO 8601 formatted date string)")
	rootCmd.PersistentFlags().String("output", "", "Override the default output format (graph for range queries, table for instant queries and metric names). Options: json,csv")
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	rootCmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "Disable table headers for instant queries")
	rootCmd.PersistentFlags().String("timeout", "10", "The timeout in seconds for all queries")
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

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
			errlog.Fatalln(err)
		}

		// Search config in home directory with name ".promql-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".promql-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			errlog.Printf("Could not read config file: %v\n", err)
		}
	}
}
