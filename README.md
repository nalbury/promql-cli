# Promql CLI
```
Query prometheus from the command line for quick analysis.

Usage:
  promql [query_string] [flags]
  promql [command]

Available Commands:
  help        Help about any command
  labels      Get a list of all labels for a given query
  meta        Get the type and help metadata for a metric
  metrics     Get a list of all prometheus metric names

Flags:
      --auth-credentials string        optional auth credentials string for http requests to prometheus
      --auth-credentials-file string   optional path to an auth credentials file for http requests to prometheus
      --auth-type string               optional auth scheme for http requests to prometheus e.g. "Basic" or "Bearer"
      --config string                  config file location (default $HOME/.promql-cli.yaml)
      --end string                     query range end (either 'now', or an ISO 8601 formatted date string) (default "now")
  -h, --help                           help for promql
      --host string                    prometheus server url (default "http://0.0.0.0:9090")
      --no-headers                     disable table headers for instant queries
      --output string                  override the default output format (graph for range queries, table for instant queries and metric names). Options: json,csv
      --start string                   query range start duration (either as a lookback in h,m,s e.g. 1m, or as an ISO 8601 formatted date string). Required for range queries
      --step string                    results step duration (h,m,s e.g. 1m) (default "1m")
      --timeout string                 the timeout in seconds for all queries (default "10")
  -v, --version                        version for promql

Use "promql [command] --help" for more information about a command.

```

## Installation
Binaries for macOS and Linux can be found on the [Releases page.](https://github.com/nalbury/promql-cli/releases)

They can also be built from source, you'll need golang 1.13.x or higher installed.

First clone the repo and `cd` into it.
```
git clone https://github.com/nalbury/promql-cli.git
cd  promql-cli/
```

Then use `make` to build and install the binary

```
OS=linux INSTALL_PATH=/usr/local/bin make install
```

More info on the various make variables and targets can be seen by running `make help`

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

You can also write your query in a file and run it with promql (useful for larger queries):

```
promql --host "http://my.prometheus.server:9090" "$(cat ./my-query.promql)" --start 1h
```

By default, instant vectors will output as a tab separated table, and range vectors will print a single [ascii graph](https://github.com/guptarohit/asciigraph) per series. All query results can be returned as either JSON or CSV formatted data using the `--output` flag (e.g. `--output csv`). This can be used to export prometheus data into other data analysis frameworks (pandas, google sheets, etc.).

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
➜  ~ promql 'sum(rate(apiserver_request_total[5m])) by (job)' --start 24h

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

For more advanced graphing of prometheus data in your terminal, I highly recommend [grafterm](https://github.com/slok/grafterm).


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

A query can be provided to narrow the list of metrics returned, for example to return all metrics that have the string `gc` in their name you can run:

```
➜  ~ promql metrics '{__name__=~".+gc.+"}'
METRICS
go_gc_duration_seconds
go_gc_duration_seconds_count
go_gc_duration_seconds_sum
go_memstats_gc_sys_bytes
go_memstats_last_gc_time_seconds
go_memstats_next_gc_bytes
prometheus_tsdb_head_gc_duration_seconds_count
prometheus_tsdb_head_gc_duration_seconds_sum
```

You can also view the metadata information for a metric (or all metrics) with the `promql meta` command.

```
➜  ~ promql meta go_goroutines
METRIC           TYPE     HELP                                          UNIT
go_goroutines    gauge    Number of goroutines that currently exist.
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

### HTTP Auth

If your prometheus server has an auth proxy in front of it, you an configure HTTP Authorization headers via cmdline flags, env vars, or in your config file. The credentials themselves can either be provided as a string, or as a file containing the credentials regardless of the method you choose for configuration. 

Command line flags:

```
➜  ~ echo -n "myuser:themostsecurepasswordinthemultiversewhywouldyoueventrytohackitdontwasteyourtime" |base64 > ~/.promql_token
➜  ~ promql --auth-type Basic --auth-credentials-file ~/.promql_token metrics |head
METRICS
prometheus_http_request_duration_seconds
prometheus_remote_storage_string_interner_zero_reference_releases_total
node_sockstat_TCP6_inuse
prometheus_sd_discovered_targets
alertmanager_nflog_query_errors_total
container_fs_writes_bytes_total
kubernetes_build_info
prometheus_tsdb_wal_corruptions_total
node_network_device_id
```

Environment variables:

```
➜  ~ export PROMQL_AUTH_TYPE="Basic"
➜  ~ export PROMQL_AUTH_CREDENTIALS="$(echo -n 'myuser:themostsecurepasswordinthemultiversewhywouldyoueventrytohackitdontwasteyourtime' |base64)"
➜  ~ promql metrics |head
METRICS
prometheus_http_request_duration_seconds
prometheus_remote_storage_string_interner_zero_reference_releases_total
node_sockstat_TCP6_inuse
prometheus_sd_discovered_targets
alertmanager_nflog_query_errors_total
container_fs_writes_bytes_total
kubernetes_build_info
prometheus_tsdb_wal_corruptions_total
node_network_device_id
```

Config file:

```
echo "auth-type: Basic" >> ~/.promql-cli.yaml
echo "auth-credentials: $(echo -n 'myuser:themostsecurepasswordinthemultiversewhywouldyoueventrytohackitdontwasteyourtime' |base64)" >> ~/.promql-cli.yaml
➜  ~ promql metrics |head
METRICS
prometheus_http_request_duration_seconds
prometheus_remote_storage_string_interner_zero_reference_releases_total
node_sockstat_TCP6_inuse
prometheus_sd_discovered_targets
alertmanager_nflog_query_errors_total
container_fs_writes_bytes_total
kubernetes_build_info
prometheus_tsdb_wal_corruptions_total
node_network_device_id
```
