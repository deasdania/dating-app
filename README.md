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

