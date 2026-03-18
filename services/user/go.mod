module github.com/romanlovesweed/yammi/services/user

go 1.24

require (
	github.com/lib/pq v1.10.9
	github.com/nats-io/nats.go v1.37.0
	github.com/romanlovesweed/yammi/pkg/events v0.0.0
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.2
)

replace github.com/romanlovesweed/yammi/pkg/events => ../../pkg/events
