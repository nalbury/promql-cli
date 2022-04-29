FROM golang:1.16-buster AS build

COPY go.mod /promql-cli/go.mod
COPY go.sum /promql-cli/go.sum

WORKDIR /promql-cli/
RUN go mod download

COPY ./ /promql-cli/
ARG TARGETARCH
RUN OS=linux ARCH=${TARGETARCH} INSTALL_PATH=/promql-cli/build/bin/ make install

FROM debian:bullseye-slim AS promql-cli

RUN apt-get update \
  && apt-get install --no-install-recommends -y ca-certificates \
  && rm -rf /var/lib/apt/lists/* /var/cache/*

COPY --from=build /promql-cli/build/bin/promql /bin/promql

RUN useradd -u 1001 -m promql
USER promql
# if needed then mount config under /home/promql/.promql-cli.yaml
# for example docker run -v ~/.promql-cli.yaml:/home/promql/.promql-cli.yaml:ro ...

ENTRYPOINT [ "/bin/promql" ]
