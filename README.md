# Promql CLI
```
Query prometheus from the command line for quick analysis

Usage:
  promql [query_string] [flags]

Flags:
      --config string   config file location (default $HOME/.promql-cli.yaml)
      --end string      Query range end (either 'now', or an RFC3339 formatted date string) (default "now")
  -h, --help            help for promql
      --host string     prometheus server url (default "http://0.0.0.0:9090")
      --no-headers      Disable table headers for instant queries
      --output string   Override the default output format (graph for range queries, and table for instant queries). Options: json,csv
      --start string    Query range start duration (either as a lookback in h,m,s e.g. 1m, or as an RFC3339 formatted date string). Required for range queries
      --step string     Results step duration (h,m,s e.g. 1m) (default "1m")
      --version         version for promql
```

## Usage
```
promql --host "http://my.prometheus.server:9090" "sum(up) by (job)"
```

This will return an instant vector of the metric `up` summed by `job`

To see this metric over time (returns a range vector) simple add the desired lookback using the start flag:
```
promql --host "http://my.prometheus.server:9090" "sum(up) by (job)" --start 1h
```

By default instant vectors will output as a tab separated table, and range vectors will print a single ascii graph (https://github.com/guptarohit/asciigraph) per series.

The values for `host`, `step`, and `output` can be set globally in a config file (default location is `$HOME/.promql-cli.yaml`). 
```
host: https://my.prometheus.server:9090
output: json
step: 5m
```
