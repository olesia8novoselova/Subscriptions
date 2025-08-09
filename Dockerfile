# build
FROM golang:1.23.4 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/subscriptions ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /bin/subscriptions /usr/local/bin/subscriptions
ENV GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["subscriptions"]
