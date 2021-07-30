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

// labelsCmd represents the labels command
var labelsCmd = &cobra.Command{
	Use:   "labels [query_string]",
	Short: "Get a list of all labels for a given query",
	Long:  `Get a list of all labels for a given query`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, warnings, err := pql.LabelsQuery(query)
		if len(warnings) > 0 {
			errlog.Printf("Warnings: %v\n", warnings)
		}
		if err != nil {
			errlog.Fatalln(err)
		}
		// Write out result
		r := writer.LabelsResult{Vector: result}
		if err := writer.WriteInstant(&r, pql.Output, pql.NoHeaders); err != nil {
			errlog.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(labelsCmd)
}
