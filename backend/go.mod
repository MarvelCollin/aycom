module aycom/backend

go 1.24.2

replace (
	aycom/backend/api-gateway => ./api-gateway
	aycom/backend/event-bus => ./event-bus
	aycom/backend/services/community => ./services/community
	aycom/backend/services/thread => ./services/thread
	aycom/backend/services/user => ./services/user
)

require (
	aycom/backend/api-gateway v0.0.0-00010101000000-000000000000 // indirect
	github.com/gin-gonic/gin v1.10.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
)
