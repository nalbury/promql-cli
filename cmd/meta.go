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

// metaCmd represents the meta command
var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Get the type and help metadata for a metric",
	Long:  "Get the type and help metadata for a metric",
	Run: func(cmd *cobra.Command, args []string) {
		var r writer.MetaResult
		result, err := pql.MetaQuery(query)
		if err != nil {
			errlog.Fatalln(err)
		}
		r = result
		if err := writer.WriteInstant(&r, pql.Output, pql.NoHeaders); err != nil {
			errlog.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(metaCmd)
}
