/*
Copyright © 2020 NAME HERE nickalbury@gmail.com

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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prometheus/common/model"

	"github.com/nalbury/promql-cli/pkg/promql"
	"github.com/nalbury/promql-cli/pkg/writer"
)

func labelsQuery(host, query, output string, timeout time.Duration) {
	client, err := promql.CreateClient(host)
	if err != nil {
		errlog.Fatalf("Error creating client, %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, warnings, err := client.Query(ctx, query, time.Now())
	if err != nil {
		errlog.Fatalf("Error querying Prometheus, %v\n", err)
	}
	if len(warnings) > 0 {
		errlog.Printf("Warnings: %v\n", warnings)
	}

	// if result is the expected type, Write it out in the
	// desired output format
	if result, ok := result.(model.Vector); ok {
		r := writer.LabelsResult{result}
		if err := writer.WriteInstant(&r, output, noHeaders); err != nil {
			errlog.Println(err)
		}
	} else {
		errlog.Println("Did not receive an instant vector")
	}
}

// labelsCmd represents the labels command
var labelsCmd = &cobra.Command{
	Use:   "labels [query_string]",
	Short: "Get a list of all labels for a given query",
	Long:  `Get a list of all labels for a given query`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		host := viper.GetString("host")
		output := viper.GetString("output")
		timeout := viper.GetInt("timeout")

		// Convert our timeout flag into a time.Duration
		t := time.Duration(int64(timeout)) * time.Second
		labelsQuery(host, query, output, t)
	},
}

func init() {
	rootCmd.AddCommand(labelsCmd)
}
