FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/board ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/board /app/board
EXPOSE 50053
CMD ["/app/board"]
