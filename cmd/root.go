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
	"log"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prometheus/common/config"

	"github.com/nalbury/promql-cli/pkg/promql"
	"github.com/nalbury/promql-cli/pkg/writer"
)

// global error logger
var errlog = log.New(os.Stderr, "", 0)

// cmd line args
var (
	pql   promql.PromQL
	query string
	// This is placeholder for the initial flag value. We ultimately parse it into the TimeoutDuration paramater of our config
	timeout int
	// timeStr is a placeholder for the inital "time" flag value. We parse it to a time.Time for use in our queries
	timeStr string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: "v0.2.1",
	Use:     "promql [query_string]",
	Short:   "Query prometheus from the command line",
	Long:    `Query prometheus from the command line for quick analysis.`,
	Args:    cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		pql.Auth.Type = viper.GetString("auth-type")
		pql.Auth.Credentials = config.Secret(viper.GetString("auth-credentials"))
		pql.Auth.CredentialsFile = viper.GetString("auth-credentials-file")
		pql.TLSConfig = config.TLSConfig{
			CAFile:             viper.GetString("tls_config.ca_cert_file"),
			CertFile:           viper.GetString("tls_config.cert_file"),
			KeyFile:            viper.GetString("tls_config.key_file"),
			ServerName:         viper.GetString("tls_config.servername"),
			InsecureSkipVerify: viper.GetBool("tls_config.insecure_skip_verify"),
		}

		pql.Host = viper.GetString("host")
		pql.Step = viper.GetString("step")
		pql.Output = viper.GetString("output")
		// Convert our timeout flag into a time.Duration
		timeout = viper.GetInt("timeout")
		pql.TimeoutDuration = time.Duration(int64(timeout)) * time.Second
		// Parse the timeStr from our --time flag if it was provided
		pql.Time = time.Now()
		if timeStr != "now" {
			t, err := time.Parse(time.RFC3339, timeStr)
			if err != nil {
				errlog.Fatalln(err)
			}
			pql.Time = t
		}
		// Create and set client interface
		cl, err := promql.CreateClientWithAuth(pql.Host, pql.Auth, pql.TLSConfig)
		if err != nil {
			errlog.Fatalln(err)
		}
		pql.Client = cl

		// Set query string if present
		// Downstream consumption of the query variable should handle any validation they need
		if len(args) > 0 {
			query = args[0]
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If we have a start time for the query, assume we're doing a range query
		if pql.Start != "" {
			result, warnings, err := pql.RangeQuery(query)
			if len(warnings) > 0 {
				errlog.Printf("Warnings: %v\n", warnings)
			}
			if err != nil {
				errlog.Fatalln(err)
			}
			r := writer.RangeResult{Matrix: result}
			if err := writer.WriteRange(&r, pql.Output, pql.NoHeaders); err != nil {
				errlog.Println(err)
			}
		} else {
			// Run query
			result, warnings, err := pql.InstantQuery(query)
			if len(warnings) > 0 {
				errlog.Printf("Warnings: %v\n", warnings)
			}
			if err != nil {
				errlog.Fatalln(err)
			}
			// Write out result
			r := writer.InstantResult{Vector: result}
			if err := writer.WriteInstant(&r, pql.Output, pql.NoHeaders); err != nil {
				errlog.Fatalln(err)
			}
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

	rootCmd.PersistentFlags().StringVar(&pql.CfgFile, "config", "", "config file location (default $HOME/.promql-cli.yaml)")
	rootCmd.PersistentFlags().String("host", "http://0.0.0.0:9090", "prometheus server url")
	if err := viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("step", "1m", "results step duration (h,m,s e.g. 1m)")
	if err := viper.BindPFlag("step", rootCmd.PersistentFlags().Lookup("step")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().StringVar(&pql.Start, "start", "", "query range start duration (either as a lookback in h,m,s e.g. 1m, or as an ISO 8601 formatted date string). Required for range queries")
	rootCmd.PersistentFlags().StringVar(&pql.End, "end", "now", "query range end (either 'now', or an ISO 8601 formatted date string)")
	rootCmd.PersistentFlags().StringVar(&timeStr, "time", "now", "time for instant queries (either 'now', or an ISO 8601 formatted date string)")
	rootCmd.PersistentFlags().String("output", "", "override the default output format (graph for range queries, table for instant queries and metric names). Options: json,csv")
	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().BoolVar(&pql.NoHeaders, "no-headers", false, "disable table headers for instant queries")
	rootCmd.PersistentFlags().String("timeout", "10", "the timeout in seconds for all queries")
	if err := viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("auth-type", "", "optional auth scheme for http requests to prometheus e.g. \"Basic\" or \"Bearer\"")
	if err := viper.BindPFlag("auth-type", rootCmd.PersistentFlags().Lookup("auth-type")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("auth-credentials", "", "optional auth credentials string for http requests to prometheus")
	if err := viper.BindPFlag("auth-credentials", rootCmd.PersistentFlags().Lookup("auth-credentials")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("auth-credentials-file", "", "optional path to an auth credentials file for http requests to prometheus")
	if err := viper.BindPFlag("auth-credentials-file", rootCmd.PersistentFlags().Lookup("auth-credentials-file")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("tls_config.ca_cert_file", "", "CA cert Path for TLS config")
	if err := viper.BindPFlag("tls_config.ca_cert_file", rootCmd.PersistentFlags().Lookup("tls_config.ca_cert_file")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("tls_config.cert_file", "", "client cert Path for TLS config")
	if err := viper.BindPFlag("tls_config.cert_file", rootCmd.PersistentFlags().Lookup("tls_config.cert_file")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("tls_config.key_file", "", "client key for TLS config")
	if err := viper.BindPFlag("tls_config.key_file", rootCmd.PersistentFlags().Lookup("tls_config.key_file")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().String("tls_config.servername", "", "server name for TLS config")
	if err := viper.BindPFlag("tls_config.servername", rootCmd.PersistentFlags().Lookup("tls_config.servername")); err != nil {
		errlog.Fatalln(err)
	}
	rootCmd.PersistentFlags().Bool("tls_config.insecure_skip_verify", false, "disable the TLS verification of server certificates.")
	if err := viper.BindPFlag("tls_config.insecure_skip_verify", rootCmd.PersistentFlags().Lookup("tls_config.insecure_skip_verify")); err != nil {
		errlog.Fatalln(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if pql.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(pql.CfgFile)
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
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("promql")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			errlog.Printf("Could not read config file: %v\n", err)
		}
	}
}
