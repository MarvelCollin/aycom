module aycom/backend

go 1.24.2

replace (
	aycom/backend/api-gateway => ./api-gateway
	aycom/backend/event-bus => ./event-bus
	aycom/backend/services/community => ./services/community
	aycom/backend/services/thread => ./services/thread
	aycom/backend/services/user => ./services/user
)

replace aycom/backend/proto/user => ./proto/user

require (
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
)

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
)
