module github.com/RomaLytar/yammi/services/auth

go 1.24

require (
	github.com/RomaLytar/yammi/pkg/events v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/nats-io/nats.go v1.37.0
	golang.org/x/crypto v0.31.0
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.2
)

replace github.com/RomaLytar/yammi/pkg/events => ../../pkg/events
