FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 go build -o /dlq main.go

FROM alpine:3.19
COPY --from=builder /dlq /dlq
ENTRYPOINT ["/dlq"]
