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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nalbury/promql-cli/pkg/promql"
	"github.com/nalbury/promql-cli/pkg/writer"
)

func metricsQuery(host, output string, timeout time.Duration) {
	client, err := promql.CreateClient(host)
	if err != nil {
		errlog.Fatalf("Error creating client, %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var r writer.MetaResult
	r, err = client.Metadata(ctx, "", "")
	if err != nil {
		errlog.Fatalf("Error querying Prometheus, %v\n", err)
	}

	// Returns an array of metrics from the metadata response.
	var m writer.MetricsResult = r.Metrics()
	// Write result
	if err := writer.WriteInstant(&m, output, noHeaders); err != nil {
		errlog.Println(err)
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
		timeout := viper.GetInt("timeout")
		// Convert our timeout flag into a time.Duration
		t := time.Duration(int64(timeout)) * time.Second
		metricsQuery(host, output, t)
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
