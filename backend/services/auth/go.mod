module github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth

go 1.21

require (
	github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto v0.0.0-00010101000000-000000000000
	github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto v0.0.0-00010101000000-000000000000
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx/v5 v5.7.4
	golang.org/x/crypto v0.31.0
	google.golang.org/grpc v1.55.0
	gorm.io/gorm v1.26.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

// Fix the import paths
replace (
	github.com/AYCOM/backend/services/auth => ./
	github.com/AYCOM/backend/services/auth/handler => ./handler
	github.com/AYCOM/backend/services/auth/model => ./model
	github.com/AYCOM/backend/services/auth/repository => ./repository
	github.com/AYCOM/backend/services/auth/service => ./service
	github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto => ./proto
	github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto => ../user/proto
)
