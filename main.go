package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	cfg "github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/handlers"
	"github.com/deasdania/dating-app/storage/postgresql"
	"github.com/deasdania/dating-app/storage/postgresutil"
	"github.com/deasdania/dating-app/storage/redis"
	"github.com/faiface/mainthread"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

const (
	serviceName = "dating-app"
)

var (
	logger      *logrus.Entry
	config      *viper.Viper
	appMetadata = &cfg.AppMetadata{}
	dbCon       *sqlx.DB
)

func main() {
	mainthread.Run(runServer)
}

func initLogger(config *viper.Viper) (*logrus.Entry, error) {
	l := cfg.NewLogger()
	var logLevel logrus.Level

	llStr := config.GetString("server.logLevel")
	appEnvStr := config.GetString("server.appEnv")
	if appEnvStr == "" {
		logger.Fatal("no configured app environment")
	}
	if llStr == "fromenv" {
		switch config.GetString("runtime.environment") {
		case "staging", "development":
			logLevel = logrus.DebugLevel // to simplify debugging
		default: // including production
			logLevel = logrus.InfoLevel
		}
	} else {
		var err error
		logLevel, err = logrus.ParseLevel(llStr)
		if err != nil {
			return nil, err
		}
	}

	l.SetLevel(logLevel)
	return l.WithFields(logrus.Fields{
		"service": serviceName,
		"app_env": appEnvStr,
	}), nil
}

func init() {
	config = viper.NewWithOptions(
		viper.EnvKeyReplacer(
			strings.NewReplacer(".", "_"),
		),
	)
	config.SetConfigFile("env/config")
	config.SetConfigType("ini")
	config.AutomaticEnv()
	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	var err error
	logger, err = initLogger(config)
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	appEnvStr := config.GetString("server.appEnv")
	if appEnvStr == "" {
		logger.Fatal("no configured app environment")
	}
	appEnvStr = strings.Title(strings.ToLower(appEnvStr))

	e := strings.ToLower(config.GetString("runtime.environment"))
	switch e {
	case "staging":
		appMetadata.Env = cfg.Env_Staging
	case "production":
		appMetadata.Env = cfg.Env_Production
	default:
		appMetadata.Env = cfg.Env_Development
	}
}

func runServer() {
	validate := cfg.NewValidator()
	e := cfg.NewEcho(config, validate)

	// Log before database initialization
	logger.Info("Initializing PostgreSQL storage...")
	var err error
	dbCon, err = postgresutil.NewStorageWithTracing(logger, config)
	if err != nil {
		cfg.WithError(err, logger).Fatal("error initializing postgres connection")
	}
	defer dbCon.Close()

	// Log after database initialization
	logger.Info("PostgreSQL storage initialized successfully.")

	redisConfig := redis.Config{
		Host:     config.GetString("redis.host"),
		Port:     config.GetInt("redis.port"),
		Password: config.GetString("redis.password"),
		Database: config.GetInt("redis.database"),
		Timeout:  config.GetDuration("redis.timeout"),
		SSL:      config.GetBool("redis.ssl"),
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rc, err := redis.NewRedisConnection(ctx, redisConfig)
	if err != nil {
		logger.Fatalf("error connecting to Redis: %v", err)
	}
	defer rc.Cl.Close()
	logger.Info("successfully connected to the Redis!")

	appEnvStr := config.GetString("server.appEnv")
	logger.Info("appEnv:", appEnvStr)

	// Create store from DB connection
	store, err := postgresql.NewStorageFromConn(logger, dbCon, appEnvStr)
	if err != nil {
		cfg.WithError(err, logger).Fatal("error initializing store")
	}
	logger.Info("Store initialized successfully.")

	handlers.Bootstrap(&handlers.API{
		App:      e,
		Log:      logger,
		Validate: validate.Validator,
		Config:   config,
		Storage:  store,
		RC:       rc,
	})
	logger.Info("Handlers bootstrap completed.")

	if config.GetBool(`debug`) {
		logger.Info("Service RUN on DEBUG mode")
	}

	// Ensure the app starts the server or handles requests
	logger.Info("Server is starting...")

	// Create a channel to signal when the server stops
	shutdownCh := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			shutdownCh <- fmt.Errorf("failed to start server: %w", err)
		}
	}()

	// Wait for an OS signal or server shutdown signal
	select {
	case err := <-shutdownCh:
		// If the server encounters an error
		logger.Fatal("Server error:", err)
	case <-ctx.Done():
		// If an interrupt signal is received (Ctrl+C, SIGTERM)
		logger.Info("Gracefully shutting down server...")

		// Set a timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		// Gracefully stop the Echo server
		if shutdownErr := e.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Fatal("Graceful shutdown failed:", shutdownErr)
		}
		logger.Info("Server gracefully shut down")
	}
}
