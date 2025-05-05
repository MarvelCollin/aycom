module aycom/backend

go 1.21.6

replace (
	aycom/backend/api-gateway => ./api-gateway
	aycom/backend/event-bus => ./event-bus
	aycom/backend/services/community => ./services/community
	aycom/backend/services/thread => ./services/thread
	aycom/backend/services/user => ./services/user
)
