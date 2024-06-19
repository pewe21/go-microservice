module github.com/pewe21/postService

go 1.22.0

replace github.com/pewe21/library => ../library

replace github.com/pewe21/userProto => ../userProto

replace github.com/pewe21/imageProto => ../imageProto

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/pewe21/imageProto v0.0.0-00010101000000-000000000000
	github.com/pewe21/library v0.0.0-00010101000000-000000000000
	github.com/pewe21/userProto v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
	google.golang.org/grpc v1.64.0
)

require (
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
