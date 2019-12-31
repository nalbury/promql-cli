#Promql CLI
```
Query prometheus from the command line for quick analysis

Usage:
  promql [query_string] [flags]

Flags:
      --config string   config file (default "$HOME/.promql-cli.yaml")
      --end string      Query range end (either 'now', or an RFC3339 formatted date string) (default "now")
  -h, --help            help for promql
      --host string     prometheus server url (default "http://0.0.0.0:9090")
      --no-headers      Disable table headers for instant queries
      --start string    Query range start duration (either as a lookback in h,m,s e.g. 1m, or as an RFC3339 formatted date string). Required for range queries
      --step string     Results step duration (h,m,s e.g. 1m) (default "1m")
      --version         version for promql
```
