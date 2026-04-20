FROM golang:1.26.0-alpine3.23 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /vkinoapi

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o /vkinoapi/bin/user-service ./services/user-service/cmd/user-service

FROM alpine:3.23

ENV TZ=Europe/Moscow

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata

WORKDIR /vkinoapi

COPY --from=builder /usr/share/zoneinfo/${TZ} /etc/localtime
RUN echo $TZ > /etc/timezone

COPY --from=builder /vkinoapi/bin/user-service ./user-service

EXPOSE 9091

CMD ["./user-service", "-config", "/etc/vkino/user.yaml"]