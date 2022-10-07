package cmd

import (
	"github.com/nalbury/promql-cli/pkg/writer"
	"github.com/spf13/cobra"
)

var targetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "Get a list of configured targets",
	Long: "Get a list of configured targets",
	Run: func(cmd *cobra.Command, args []string) {
		var r writer.SeriesResult
		result, err := pql.Targets(query)
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
	rootCmd.AddCommand(targetsCmd)
}

