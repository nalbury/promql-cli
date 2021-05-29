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

	"github.com/nalbury/promql-cli/pkg/promql"
	"github.com/nalbury/promql-cli/pkg/writer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func metaQuery(host, query, output string, timeout time.Duration) {
	client, err := promql.CreateClient(host)
	if err != nil {
		errlog.Fatalf("Error creating client, %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var r writer.MetaResult
	r, err = client.Metadata(ctx, query, "")
	if err != nil {
		errlog.Fatalf("Error querying Prometheus, %v\n", err)
	}

	// if result is the expected type, Write it out in the
	// desired output format
	if err := writer.WriteInstant(&r, output, noHeaders); err != nil {
		errlog.Println(err)
	}
}

// metaCmd represents the meta command
var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Get the type and help metadata for a metric",
	Long:  "Get the type and help metadata for a metric",
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}
		host := viper.GetString("host")
		output := viper.GetString("output")
		timeout := viper.GetInt("timeout")
		// Convert our timeout flag into a time.Duration
		t := time.Duration(int64(timeout)) * time.Second
		metaQuery(host, query, output, t)
	},
}

func init() {
	rootCmd.AddCommand(metaCmd)
}
