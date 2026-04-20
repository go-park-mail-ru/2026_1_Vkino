FROM golang:1.26.0-alpine3.23 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /vkinoapi

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_PATH
ARG BINARY_NAME

RUN test -n "$SERVICE_PATH"
RUN test -n "$BINARY_NAME"

RUN go build -ldflags="-w -s" -o /vkinoapi/bin/${BINARY_NAME} ./${SERVICE_PATH}

FROM alpine:3.23

ENV TZ=Europe/Moscow

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata

WORKDIR /vkinoapi

COPY --from=builder /usr/share/zoneinfo/${TZ} /etc/localtime
RUN echo $TZ > /etc/timezone

ARG BINARY_NAME
RUN test -n "$BINARY_NAME"

COPY --from=builder /vkinoapi/bin/${BINARY_NAME} ./${BINARY_NAME}

CMD ["sh", "-c", "./${BINARY_NAME}"]