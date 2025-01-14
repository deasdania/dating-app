package main

import (
	"log"
	"strings"

	"github.com/faiface/mainthread"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	cfg "github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/handlers"
	storage "github.com/deasdania/dating-app/storage/postgresql"
	"github.com/deasdania/dating-app/storage/postgresutil"
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

func main() {
	mainthread.Run(runServer)
}

func runServer() {
	validate := cfg.NewValidator()
	app := cfg.NewEcho(config, validate)

	// Create a new postgres storage
	var err error
	dbCon, err = postgresutil.NewStorageWithTracing(logger, config)
	if err != nil {
		cfg.WithError(err, logger).Fatal("error initializing postgres connection")
	}
	defer dbCon.Close()

	appEnvStr := config.GetString("server.appEnv")
	logger.Info("appEnv:", appEnvStr)
	store, err := storage.NewStorageFromConn(logger, dbCon, appEnvStr)
	if err != nil {
		cfg.WithError(err, logger).Fatal("error initializing postgres connection")
	}

	handlers.Bootstrap(&handlers.API{
		App:      app,
		Log:      logger,
		Validate: validate.Validator,
		Config:   config,
	})

	if config.GetBool(`debug`) {
		logger.Info("Service RUN on DEBUG mode")
	}

	logger.Fatal(app.Start(config.GetString("server.address")))
}
