module aycom/backend/services/thread

go 1.24.2

require (
	aycom/backend/proto v0.0.0-00010101000000-000000000000
	aycom/backend/services/user v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.26.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
)

replace aycom/backend/services/user => ../user
replace aycom/backend/proto => ../../proto
