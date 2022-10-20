ARG golang_ver=alpine
ARG alpine_ver=latest
FROM golang:$golang_ver AS builder

RUN apk --no-cache add \
    build-base

WORKDIR /go/app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go test -mod=readonly ./...
RUN go build -mod=readonly ./cmd/godirectserverd


FROM alpine:$alpine_ver

COPY --from=builder /go/app/godirectserverd /usr/local/bin/godirectserverd

COPY etc/rules.conf /etc/godirectserverd.conf

ENV GDS_HOST_HEADER=host
ENV GDS_LISTEN_PORT=8080
ENV GDS_LISTEN_ADDR=0.0.0.0:$GDS_LISTEN_PORT
ENV GDS_RULE_FILE=/etc/godirectserverd.conf

EXPOSE $GDS_LISTEN_PORT

CMD exec /usr/local/bin/godirectserverd \
    -host-header="$GDS_HOST_HEADER" \
    -listen-addr="$GDS_LISTEN_ADDR" \
    -rule-file="$GDS_RULE_FILE"
