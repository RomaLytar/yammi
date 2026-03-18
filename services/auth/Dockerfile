FROM golang:1.24-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev git
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build
COPY pkg/ pkg/
COPY services/auth/ services/auth/

WORKDIR /build/services/auth

# Generate protobuf
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/auth.proto

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/auth /app/auth
COPY --from=builder /build/services/auth/migrations /app/migrations
EXPOSE 50051
CMD ["/app/auth"]
