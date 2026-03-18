FROM golang:1.24-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev git
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build
COPY pkg/ pkg/
COPY services/user/ services/user/

WORKDIR /build/services/user

# Generate protobuf
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/user.proto

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/user /app/user
COPY --from=builder /build/services/user/migrations /app/migrations
EXPOSE 50052
CMD ["/app/user"]
