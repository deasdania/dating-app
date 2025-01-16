dating app

how to apply migrations
goose -dir migrations postgres "user=your_pg_user password=your_pg_password dbname=golang_app sslmode=disable" up

goose -dir migrations postgres "user=postgres password=secret dbname=dating_app sslmode=disable" up


how to rollback migrations
cd migrations 
goose -dir migrations postgres "user=your_pg_user password=your_pg_password dbname=golang_app sslmode=disable" down


to add new migration file, 
cd migrations
goose create create_users_table sql


how to run the seeds
cd seeds
go run main.go

use redis 
docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest
docker run -d --name redis-stack -p 8001:8001 redis/redis-stack:latest


open the redis with localhost:8001
or cli 
docker exec -t redis-stack redis-cli

export REDIS_CONNECTION=redis://localhost:6379