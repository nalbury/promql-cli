FROM golang:1.16-buster AS build

ADD go.mod /promql-cli/go.mod
ADD go.sum /promql-cli/go.sum

WORKDIR /promql-cli/
RUN go mod download

RUN apt-get update && apt-get install -y make

ADD ./ /promql-cli/
ARG TARGETARCH
RUN OS=linux ARCH=${TARGETARCH} make build 

FROM debian:buster-slim AS promql-cli 
COPY --from=build /promql-cli/build/bin/linux/${TARGETARCH}/promql /bin/promql
ENTRYPOINT [ "/bin/promql" ]
