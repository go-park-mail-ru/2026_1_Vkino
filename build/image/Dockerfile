FROM golang:1.26.0-alpine3.23 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk update --no-cache && apk add --no-cache tzdata

COPY go.mod go.sum /vkinoapi/

WORKDIR /vkinoapi

RUN go mod download

COPY . /vkinoapi

RUN go build -ldflags="-w -s" -o /vkinoapi/bin/vkinoapi ./cmd/api/

FROM golang:1.26.0-alpine3.23

ENV TZ=Europe/Moscow

EXPOSE 8080

WORKDIR /vkinoapi

COPY --from=builder /usr/share/zoneinfo/${TZ} /etc/localtime
RUN echo $TZ > /etc/timezone

COPY --from=builder /vkinoapi/bin/vkinoapi .