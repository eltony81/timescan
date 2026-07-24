FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o timescan-e2e e2e/app/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/timescan-e2e /app/timescan-e2e

ENTRYPOINT ["/app/timescan-e2e"]
