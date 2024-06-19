module github.com/pewe21/userService

go 1.22.0

require (
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/pewe21/imageProto v0.0.0-00010101000000-000000000000
	github.com/pewe21/library v1.0.0
	github.com/pewe21/userProto v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
	google.golang.org/grpc v1.64.0
)

require (
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)

require (
	golang.org/x/crypto v0.24.0
	golang.org/x/sys v0.21.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/pewe21/library => ../library

replace github.com/pewe21/userProto => ../userProto

replace github.com/pewe21/imageProto => ../imageProto
