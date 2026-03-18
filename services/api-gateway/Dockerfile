FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-gateway ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/api-gateway /app/api-gateway
EXPOSE 8080
CMD ["/app/api-gateway"]
