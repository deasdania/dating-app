# Dating App

## Tech Stacks

- **Programming Language**: Go (Golang) or TypeScript
- **Framework for API**:
  - Golang:  `Echo`, or `Fiber` (lightweight and fast)
  
- **Database**: 
  - **Primary DB**: PostgreSQL (Relational DB)
  - **Caching**: Redis (for fast lookup of swipe limits, etc.)
  - **Authentication**: JWT for secure and stateless login sessions
  
## The structure of the service
- **config**: Application settings and configuration.
- **core**: Core business logic and models.
- **env**: Environment-specific settings.
- **go**.mod & go.sum: Go dependency management.
- **handlers**: HTTP handlers for API routes.
- **main**.go: Entry point of the application.
- **migrations**: SQL migration files for DB schema changes.
- **models**: Database models and related logic.
- **routes**.json: Route definitions for URL handling.
- **seeds**: Initial or mock data seeding for the database.
- **status**: Status codes or enums related to business logic.
- **storage**: Database and caching logic (PostgreSQL, Redis).
- **tests**: Tests for the application, including e2e tests(TBC).
- **utils**: Utility functions for common tasks.

## Instructions on how to run the service

### Prepare the postgres 

create a new postgres db locally before run the application.

---
#### How to Apply Migrations

To apply migrations using **Goose**, run the following command:

```bash
goose -dir migrations postgres "user=your_pg_user password=your_pg_password dbname=golang_app sslmode=disable" up
```

Example:

```bash
goose -dir migrations postgres "user=postgres password=secret dbname=dating_app sslmode=disable" up
```

This command will apply the migrations to your database.

---
#### Development notes
#####  How to Rollback Migrations

To rollback migrations, navigate to the `migrations` directory and run:

```bash
cd migrations
goose -dir migrations postgres "user=your_pg_user password=your_pg_password dbname=golang_app sslmode=disable" down
```

This command will roll back the most recent migration.

---

##### How to Add a New Migration File

To create a new migration file, run the following command:

```bash
cd migrations
goose create create_users_table sql
```

This will create a new migration file with a timestamp and a name of your choosing. Replace `create_users_table` with a descriptive name for your migration.

---

### Prepare the Redis

#### Running Redis with Docker

To run Redis, use the following Docker command:

```bash
docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest
```

Alternatively, to run **Redis Stack**, you can use:

```bash
docker run -d --name redis-stack -p 8001:8001 redis/redis-stack:latest
```

#### Access Redis

- To access Redis via the **Redis Stack UI**, open your browser and go to `http://localhost:8001`.
  
- To access Redis via the **CLI**, run the following command:

```bash
docker exec -t redis-stack redis-cli
```
---
## Setting The ENV

copy the  `env/sample.config` and rename to `config`
adjust the env values with your values, you can't run the seeds, before setting up this env

---

## Run the seeds

To run the database seeders, navigate to the `seeds` directory and execute the following command:

```bash
cd seeds
go run main.go
```

This will populate your database with initial data.

---

### Setting The ENV

copy the  `env/sample.config` and rename to `config`
adjust the env values with your values

---

# RUN THE APP 
to the root project and run this command below:
```bash
go run main.go
```

### Summary of Commands:

- **Apply Migrations**:
  ```bash
  goose -dir migrations postgres "user=your_pg_user password=your_pg_password dbname=golang_app sslmode=disable" up
  ```
  
- **Rollback Migrations**:
  ```bash
  goose -dir migrations postgres "user=your_pg_user password=your_pg_password dbname=golang_app sslmode=disable" down
  ```

- **Add a New Migration**:
  ```bash
  goose create create_users_table sql
  ```

- **Run Seeds**:
  ```bash
  cd seeds
  go run main.go
  ```

- **Run Redis (Docker)**:
  ```bash
  docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest
  ```

- **Access Redis CLI**:
  ```bash
  docker exec -t redis-stack redis-cli
  ```

- **Set Redis Connection**:
  ```bash
  export REDIS_CONNECTION=redis://localhost:6379
  ```

