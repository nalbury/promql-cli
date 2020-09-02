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

func instantQuery(host, queryString, output string) {
	client, err := promql.CreateClient(host)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := client.Query(ctx, queryString, time.Now())
	if err != nil {
		errlog.Fatalf("Error querying Prometheus: %v\n", err)
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

func getRange(step, start, end string) (r v1.Range, err error) {
	if s, err := time.Parse(time.RFC3339, start); err == nil {
		r.Start = s
	} else if l, err := time.ParseDuration(start); err == nil {
		r.Start = time.Now().Add(-l)
	} else {
		err = fmt.Errorf("Unable to parse range start start time: %w", err)
		return r, err
	}

	r.End = time.Now()
	r.Step = time.Minute

	if step != "" {
		s, err := time.ParseDuration(step)
		if err != nil {
			errlog.Fatalln(err)
		}
		r.Step = s
	}

	if end != "now" {
		e, err := time.Parse(time.RFC3339, end)
		if err != nil {
			err = fmt.Errorf("Unable to parse range end time: %w", err)
			return r, err
		}
		r.End = e
	}

	return r, err
}

func rangeQuery(host, queryString, output string, r v1.Range) {
	// Create client
	client, err := promql.CreateClient(host)
	if err != nil {
		errlog.Fatalf("Error creating client: %v\n", err)
	}

	// create context with a timeout,
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// execute query
	result, warnings, err := client.QueryRange(ctx, queryString, r)
	if err != nil {
		errlog.Fatalf("Error querying Prometheus: %v\n", err)
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
	Version: "v0.1.0",
	Use:     "promql [query_string]",
	Short:   "Query prometheus from the command line",
	Long:    `Query prometheus from the command line for quick analysis.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		host := viper.GetString("host")
		step := viper.GetString("step")
		output := viper.GetString("output")
		// If we have a start time for the query, assume we're doing a range query
		if start != "" {
			// Parse the time range from the cmd line/config file options
			r, err := getRange(step, start, end)
			if err != nil {
				errlog.Fatalln(err)
			}
			// Execute our range query
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
	rootCmd.PersistentFlags().StringVar(&start, "start", "", "Query range start duration (either as a lookback in h,m,s e.g. 1m, or as an ISO 8601 formatted date string). Required for range queries")
	rootCmd.PersistentFlags().StringVar(&end, "end", "now", "Query range end (either 'now', or an ISO 8601 formatted date string)")
	rootCmd.PersistentFlags().String("output", "", "Override the default output format (graph for range queries, table for instant queries and metric names). Options: json,csv")
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
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Could not read config file: %w\n", err)
		}
	}
}
