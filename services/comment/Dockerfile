FROM golang:1.24-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev git
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build

# Копируем board service (нужен для replace directive — proto для IsMember gRPC client)
COPY services/board /board

# Копируем comment service
COPY services/comment/go.mod services/comment/go.sum* ./
RUN go mod download
COPY services/comment/ .

# Generate protobuf
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/comment.proto

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/comment ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/comment /app/comment
COPY services/comment/migrations /app/migrations
EXPOSE 50054 2112
CMD ["/app/comment"]
