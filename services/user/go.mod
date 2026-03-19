module github.com/romanlovesweed/yammi/services/user

go 1.24

require (
	github.com/lib/pq v1.10.9
	github.com/nats-io/nats.go v1.37.0
	github.com/romanlovesweed/yammi/pkg/events v0.0.0
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.2
)

require (
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/romanlovesweed/yammi/pkg/events => ../../pkg/events
