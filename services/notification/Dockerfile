FROM golang:1.24-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev git
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build
COPY pkg/ pkg/
COPY services/notification/ services/notification/

WORKDIR /build/services/notification

# Generate protobuf
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/notification.proto

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/notification-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/notification-consumer ./cmd/consumer

# --- API image ---
FROM alpine:3.19 AS api
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/notification-api /app/notification-api
COPY --from=builder /build/services/notification/migrations /app/migrations
EXPOSE 50055 2112
CMD ["/app/notification-api"]

# --- Consumer image ---
FROM alpine:3.19 AS consumer
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/notification-consumer /app/notification-consumer
EXPOSE 2113
CMD ["/app/notification-consumer"]
