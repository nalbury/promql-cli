# Promql CLI
```
Query prometheus from the command line for quick analysis.

Usage:
  promql [query_string] [flags]
  promql [command]

Available Commands:
  help        Help about any command
  labels      Get a list of all labels for a given query
  metrics     Get a list of all prometheus metric names

Flags:
      --config string    config file location (default $HOME/.promql-cli.yaml)
      --end string       Query range end (either 'now', or an ISO 8601 formatted date string) (default "now")
  -h, --help             help for promql
      --host string      prometheus server url (default "http://0.0.0.0:9090")
      --no-headers       Disable table headers for instant queries
      --output string    Override the default output format (graph for range queries, table for instant queries and metric names). Options: json,csv
      --start string     Query range start duration (either as a lookback in h,m,s e.g. 1m, or as an ISO 8601 formatted date string). Required for range queries
      --step string      Results step duration (h,m,s e.g. 1m) (default "1m")
      --timeout string   The timeout in seconds for all queries (default "10")
      --version          version for promql

Use "promql [command] --help" for more information about a command.

```

## Installation
```
curl -o /usr/local/bin/promql https://promql-cli.s3.amazonaws.com/latest/macos/promql && chmod +x /usr/local/bin/promql
```

Specific versions can be installed by replacing `latest` in the URL above with any version tag (e.g. v0.2.0).

## Usage

### Instant and Range Queries

```
promql --host "http://my.prometheus.server:9090" 'sum(up) by (job)'
```

This will return an instant vector of the metric `up` summed by `job`

To see this metric over time (returns a range vector) simple add the desired lookback using the start flag:
```
promql --host "http://my.prometheus.server:9090" "sum(up) by (job)" --start 1h
```

By default instant vectors will output as a tab separated table, and range vectors will print a single [ascii graph](https://github.com/guptarohit/asciigraph) per series. All query results can be returned as either JSON or CSV formated data using the `--output` flag (e.g. `--output csv`. This can be used to export prometheus data into other data analysis frameworks (pandas, google sheets, etc.).

The values for `host`, `step`, `output` and `timeout` can be set globally in a config file (default location is `$HOME/.promql-cli.yaml`).

```
host: https://my.prometheus.server:9090
output: json
step: 5m
```

#### Example Instant Vector
```
➜  ~ promql 'sum(rate(apiserver_request_total[24h])) by (instance)'
INSTANCE                VALUE                 TIMESTAMP
123.456.789.123:6443    14.868565474122951    2020-09-27T09:34:22-04:00
234.567.891.234:6443    9.148373277758477     2020-09-27T09:34:22-04:00
345.678.912.345:6443    11.77870788468218     2020-09-27T09:34:22-04:00

```

#### Example Range Vector
```
➜  ~ promql 'sum(rate(apiserver_request_total{cluster="production",clusterID="green"}[5m])) by (job)' --start 24h

##################################################
# TIME_RANGE: Sep 26 09:37:35 -> Sep 27 09:37:35 #
# METRIC: {job="apiserver"}                      #
##################################################
 38.31 ┤          ╭╮
 38.04 ┤          ││   ╭╮                            ╭╮╭╮                         ╭╮                               ╭╮
 37.76 ┤        ╭─╯╰╮ ╭╯╰─╮                   ╭╮╭───╮││││                    ╭╮╭╮╭╯│╭╮                             ││                        ╭─╮
 37.49 ┤        │   ╰╮│   │                  ╭╯╰╯   ││╰╯│╭╮                  │││╰╯ ││╰──╮                        ╭╮││                       ╭╯ │  ╭╮
 37.22 ┤        │    ╰╯   ╰─╮                │      ╰╯  ╰╯│                ╭╮│╰╯   ╰╯   │                  ╭╮╭╮  │││╰╮                    ╭╮│  │  │╰╮
 36.94 ┤        │           ╰╮               │            │                │╰╯          │                ╭╮│╰╯│ ╭╯││ │                 ╭──╯││  ╰─╮│ │
 36.67 ┤       ╭╯            ╰╮              │            │                │            │              ╭╮│╰╯  │╭╯ ╰╯ ╰╮                │   ╰╯    ││ │
 36.40 ┤       │              │              │            ╰╮               │            ╰╮             │╰╯    ╰╯      │                │         ╰╯ ╰╮
 36.12 ┤      ╭╯              │             ╭╯             ╰╮              │             │             │              ╰╮              ╭╯             │
 35.85 ┤      │               │             │               │             ╭╯             │             │               │              │              │ ╭╮
 35.58 ┤      │               ╰╮            │               │             │              │             │               │              │              │ ││
 35.31 ┤   ╭─╮│                │            │               ╰╮          ╭╮│              ╰╮      ╭╮    │               ╰╮            ╭╯              ╰╮││ ╭╮
 35.03 ┼╮╭╮│ ││                │           ╭╯                │ ╭╮     ╭╮│╰╯               │ ╭╮   ││    │                │    ╭╮      │                │││ ││╭
 34.76 ┤││╰╯ ╰╯                ╰╮ ╭────╮ ╭─╯                 ╰╮│╰─╮╭╮╭╯╰╯                 ╰─╯│ ╭╮│╰╮   │                │    ││      │                │││╭╯╰╯
 34.49 ┤╰╯                      ╰─╯    │╭╯                    ╰╯  ╰╯╰╯                       ╰─╯││ │╭╮╭╯                ╰────╯│ ╭────╯                ││╰╯
 34.21 ┤                               ╰╯                                                       ╰╯ ││╰╯                       │╭╯                     ╰╯
 33.94 ┤                                                                                           ╰╯                         ╰╯


```

### Metrics and Labels

In addition to querying prometheus data, you can also query for metrics and labels available in the dataset.

#### Metrics

The `promql metrics` command returns all current metrics available in prometheus.

```
➜  ~ promql metrics |head -10
METRICS
go_gc_duration_seconds
go_gc_duration_seconds_count
go_gc_duration_seconds_sum
go_goroutines
go_info
go_memstats_alloc_bytes
go_memstats_alloc_bytes_total
go_memstats_buck_hash_sys_bytes
go_memstats_frees_total

```

#### Labels

The `promql labels` command returns all current labels available for a given query.

```
➜  ~ promql labels apiserver_request_total
LABELS
__name__
client
code
component
contenttype
endpoint
group
instance
job
namespace
prometheus
prometheus_replica
resource
scope
service
subresource
verb
version

```
