module aycom/backend/services/auth

go 1.21

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.3.0
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.31.0
	golang.org/x/net v0.21.0
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
)

// No replace directives needed with local module path
