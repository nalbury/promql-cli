FROM golang:1.18-buster AS build

COPY go.mod /promql-cli/go.mod
COPY go.sum /promql-cli/go.sum

WORKDIR /promql-cli/
RUN go mod download

COPY ./ /promql-cli/
ARG TARGETARCH
RUN OS=linux ARCH=${TARGETARCH} INSTALL_PATH=/promql-cli/build/bin/ make install

FROM gcr.io/distroless/base-debian11:nonroot AS promql-cli
COPY --from=build /promql-cli/build/bin/promql /promql
ENTRYPOINT [ "/promql" ]
