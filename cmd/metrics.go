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
	"github.com/nalbury/promql-cli/pkg/writer"
	"github.com/spf13/cobra"
)

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics [query_string]",
	Short: "Get a list of prometheus metric names matching the provided query",
	Long:  `Get a list of prometheus metric names matching the provided query. If no query is provided, all metric names will be returned.`,
	Run: func(cmd *cobra.Command, args []string) {
		var r writer.SeriesResult
		if query == "" {
			query = `{job=~".+"}`
		}
		result, warnings, err := pql.SeriesQuery(query)
		if len(warnings) > 0 {
			errlog.Printf("Warnings: %v\n", warnings)
		}
		if err != nil {
			errlog.Fatalln(err)
		}
		r = result
		var m writer.MetricsResult = r.Metrics()
		if err := writer.WriteInstant(&m, pql.Output, pql.NoHeaders); err != nil {
			errlog.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
