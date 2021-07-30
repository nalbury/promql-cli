FROM golang:1.16-buster AS build

ADD go.mod /promql-cli/go.mod
ADD go.sum /promql-cli/go.sum

WORKDIR /promql-cli/
RUN go mod download

RUN apt-get update && apt-get install -y make

ADD ./ /promql-cli/
ARG TARGETARCH
RUN OS=linux ARCH=${TARGETARCH} INSTALL_PATH=/promql-cli/build/bin/ make install

# TODO explore other base images here.
# Requirements:
# - small
# - correct perms available for mounting and using a config file (config > cmdline flags), right now we just run as root... (see below)
# I'm generally not a fan of alpine/busybox for a cmdline env but maybe minideb or similar?
# Stock deb slim is pretty rad already :)
FROM debian:buster-slim AS promql-cli
COPY --from=build /promql-cli/build/bin/promql /bin/promql

RUN apt-get update \
  && apt-get install -y ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# TODO don't run as root...
ENTRYPOINT [ "/bin/promql" ]
