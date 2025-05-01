module github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth

go 1.23

toolchain go1.24.2

require (
	github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto v0.0.0-00010101000000-000000000000
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.4.3
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.33.0
	google.golang.org/grpc v1.72.0
	gorm.io/gorm v1.25.7-0.20240204074919-46816ad31dde
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto => ./proto

replace github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto => ../user/proto
