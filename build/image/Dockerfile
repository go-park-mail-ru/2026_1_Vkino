FROM golang:1.26.0-alpine3.23 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV TZ=Europe/Moscow

RUN apk update --no-cache && apk add --no-cache tzdata ca-certificates

WORKDIR /vkinoapi

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_PATH
ARG BINARY_NAME

RUN test -n "$SERVICE_PATH"
RUN test -n "$BINARY_NAME"

RUN go build -ldflags="-w -s" -o /vkinoapi/bin/${BINARY_NAME} ./${SERVICE_PATH}

FROM golang:1.26.0-alpine3.23

ENV TZ=Europe/Moscow

WORKDIR /vkinoapi

ARG BINARY_NAME
ARG PORT=8080

RUN test -n "$BINARY_NAME"

COPY --from=builder /usr/share/zoneinfo/${TZ} /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN echo $TZ > /etc/timezone

COPY --from=builder /vkinoapi/bin/${BINARY_NAME} ./${BINARY_NAME}

EXPOSE ${PORT}

CMD ["sh", "-c", "./${BINARY_NAME}"]