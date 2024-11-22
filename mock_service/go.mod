module mock_service

go 1.22.5

require github.com/lib/pq v1.10.9

require (
	utils/json v0.0.0-00010101000000-000000000000
	utils/jwt v0.0.0-00010101000000-000000000000
	utils/redis v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.4 // indirect
	github.com/golang-jwt/jwt/v5 v5.0.0 // indirect
)

replace utils/redis => ../utils/redis

replace utils/json => ../utils/json

replace utils/jwt => ../utils/jwt
