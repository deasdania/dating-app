package postgresutil

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
)

type NamedExecerContext interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type GetterContext interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

const driver = "postgres"

type DBConfig struct {
	User     string `mapstructure:"user"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DBName   string `mapstructure:"dbname"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslMode"`
}

// NewStorageWithTracing returns a new StorageUtil from config that has distributed tracing capability.
func NewStorageWithTracing(logger logrus.FieldLogger, config *viper.Viper) (*sqlx.DB, error) {
	db, err := Connectx(config)
	if err != nil {
		return nil, err
	}
	// TODO: is this a sane default?
	// The current max_connections in postgres is 100.
	db.SetMaxOpenConns(50)
	db.SetConnMaxLifetime(time.Hour)
	return db, nil
}

// NewDBFromConfig build database connection string from config file.
func NewDBFromConfig(config *viper.Viper) (*DBConfig, error) {
	var allConfig struct {
		Database DBConfig `mapstructure:"database"`
	}
	if err := config.Unmarshal(&allConfig); err != nil {
		return nil, fmt.Errorf("cannot unmarshal db config: %w", err)
	}

	return &allConfig.Database, nil
}

// NewDBStringFromConfig build database connection string from config file.
func NewDBStringFromConfig(config *viper.Viper) (string, error) {
	var allConfig struct {
		Database DBConfig `mapstructure:"database"`
	}
	if err := config.Unmarshal(&allConfig); err != nil {
		return "", fmt.Errorf("cannot unmarshal db config: %w", err)
	}

	return NewDBStringFromDBConfig(allConfig.Database)
}

func NewDBStringFromDBConfig(config DBConfig) (string, error) {
	var dbParams []string
	dbParams = append(dbParams, fmt.Sprintf("user=%s", config.User))
	dbParams = append(dbParams, fmt.Sprintf("host=%s", config.Host))
	dbParams = append(dbParams, fmt.Sprintf("port=%s", config.Port))
	dbParams = append(dbParams, fmt.Sprintf("dbname=%s", config.DBName))
	if password := config.Password; password != "" {
		dbParams = append(dbParams, fmt.Sprintf("password=%s", password))
	}
	dbParams = append(dbParams, fmt.Sprintf("sslmode=%s",
		config.SSLMode))
	return strings.Join(dbParams, " "), nil
}

// Connectx opens a connection to database with given connection string using sqlx opener
// and verify the connection with a ping.
func Connectx(config *viper.Viper) (*sqlx.DB, error) {
	dbString, err := NewDBStringFromConfig(config)
	if err != nil {
		return nil, err
	}

	db, err := otelsqlx.Connect(driver, dbString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func usage() {
	const (
		usageRun      = `goose [OPTIONS] COMMAND`
		usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with next version`
	)
	fmt.Println(usageRun)
	flag.PrintDefaults()
	fmt.Println(usageCommands)
}

func Migrate() error {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Print or log the current directory
	log.Printf("Current directory: %s", dir)
	configPath := flag.String("config", "env/config", "config file")
	format := flag.String("format", "ini", "config format")

	flag.Usage = usage

	flag.Parse()
	args := flag.Args()

	config := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
	)
	config.SetConfigFile(*configPath)
	config.SetConfigType(*format)
	config.AutomaticEnv()
	if err := config.ReadInConfig(); err != nil {
		return fmt.Errorf("error loading configuration for migration: %v", err)
	}

	db, err := Open(config)
	if err != nil {
		return fmt.Errorf("error opening db connection: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err = goose.SetDialect(driver); err != nil {
		return fmt.Errorf("failed to set goose dialect: %v", err)
	}

	if len(args) == 0 {
		return errors.New("expected at least one arg")
	}

	command := args[0]

	migrationDir := config.GetString("database.migrationDir")
	if err = goose.Run(command, db, migrationDir, args[1:]...); err != nil {
		return fmt.Errorf("goose run: %v", err)

	}
	return db.Close()
}

func Open(config *viper.Viper) (*sql.DB, error) {
	dbString, err := NewDBStringFromConfig(config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driver, dbString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// replaceDBName replaces the dbname option in connection string with given db name in parameter.
func replaceDBName(connStr, dbName string) string {
	r := regexp.MustCompile(`dbname=([^\s]+)\s`)
	return r.ReplaceAllString(connStr, fmt.Sprintf("dbname=%s ", dbName))
}

// MustNewDevelopmentDB creates a new isolated database for the use of a package test
// The checking of dbconn is expected to be done in the package test using this
func MustNewDevelopmentDB(ddlConnStr, migrationDir string) (*sqlx.DB, func()) {
	const driver = "postgres"

	dbName := uuid.New().String()
	ddlDB := otelsqlx.MustConnect(driver, ddlConnStr)
	ddlDB.MustExec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	if err := ddlDB.Close(); err != nil {
		panic(err)
	}

	connStr := replaceDBName(ddlConnStr, dbName)
	db := otelsqlx.MustConnect(driver, connStr)

	if err := goose.Run("up", db.DB, migrationDir); err != nil {
		panic(err)
	}

	tearDownFn := func() {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close database connection: %s", err.Error())
		}
		ddlDB, err := otelsqlx.Connect(driver, ddlConnStr)
		if err != nil {
			log.Fatalf("failed to connect database: %s", err.Error())
		}

		if _, err = ddlDB.Exec(fmt.Sprintf(`DROP DATABASE "%s"`, dbName)); err != nil {
			log.Fatalf("failed to drop database: %s", err.Error())
		}

		if err = ddlDB.Close(); err != nil {
			log.Fatalf("failed to close DDL database connection: %s", err.Error())
		}
	}

	return db, tearDownFn
}
