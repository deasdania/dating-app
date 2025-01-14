package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/deasdania/dating-app/storage/postgresutil"
)

const (
	dbConnEnv = "DATABASE_CONNECTION"
	localEnv  = "local"
	devEnv    = "development"
)

//go:generate mockgen -source=storage.go -destination=mock/mock_storage.go -aux_files brank.as/data/backend/statement/storage/statement=reader.go,brank.as/data/backend/statement/storage/statement=writer.go
//go:generate gofumpt -s -w mock/mock_storage.go

type IStore interface {
	IReaderStore
	IWriterStore
}

type GetterContext interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Storage provides a wrapper around an sql database and provides
// required methods for interacting with the database
type Storage struct {
	logger logrus.FieldLogger
	db     *sqlx.DB
	env    string

	// dataEncryptionKey is the key used for protecting the secret fields
	dataEncryptionKey []byte
	defaultPageSize   int
}

// NewStorageFromConn returns a new Storage from the provides psql databse string
func NewStorageFromConn(logger logrus.FieldLogger, db *sqlx.DB, env string) (*Storage, error) {
	logger.Info("env:", env)
	return &Storage{
		logger:          logger,
		db:              db,
		defaultPageSize: 10,
		env:             env,
	}, nil
}

// NewStorageWithTracing returns a new StorageUtil from config that has distributed tracing capability.
func NewStorageWithTracing(logger logrus.FieldLogger, config *viper.Viper) (*sqlx.DB, error) {
	db, err := postgresutil.Connectx(config)
	if err != nil {
		return nil, err
	}
	// TODO: is this a sane default?
	// The current max_connections in postgres is 100.
	db.SetMaxOpenConns(50)
	db.SetConnMaxLifetime(time.Hour)
	return db, nil
}

type pgxTx struct{}
type Tx struct {
	Tx        *sqlx.Tx
	committed bool
}

func (s *Storage) BeginTx(ctx context.Context) (*Tx, error) {
	btx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	return &Tx{Tx: btx}, nil
}

// Commit commits the transaction
func (s *Storage) Commit(ctx context.Context) error {
	t := getTx(ctx)
	if t == nil {
		return fmt.Errorf("not a transaction context")
	}
	if t.committed {
		return nil // Already committed
	}
	if err := t.Tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	t.committed = true
	return nil
}

// Rollback rolls back the transaction
func (s *Storage) Rollback(ctx context.Context) error {
	t := getTx(ctx)
	if t == nil {
		return fmt.Errorf("not a transaction context")
	}
	if t.committed {
		return nil // Already committed
	}
	if err := t.Tx.Rollback(); err != nil {
		return fmt.Errorf("error rolling back transaction: %w", err)
	}
	t.committed = true
	return nil
}

// GetDBConn returns the underlying sql.DB object
func (s Storage) GetDBConn() *sql.DB {
	return s.db.DB
}

// SetDataEncryptionKey stores the DEK in the storage layer
// created/fetched by secret handler.
// This allows the table field to be encrypted and decrypted
func (s *Storage) SetDataEncryptionKey(key []byte) {
	s.dataEncryptionKey = key
}

func getTx(ctx context.Context) *Tx {
	if t, ok := ctx.Value(pgxTx{}).(*Tx); ok {
		return t
	}
	return nil
}

func newTestStorage(t *testing.T) (*Storage, func()) {
	ddlConnStr := os.Getenv(dbConnEnv)
	if ddlConnStr == "" {
		t.Skipf("%s is not set", dbConnEnv)
	}

	db, teardown := postgresutil.MustNewDevelopmentDB(ddlConnStr, filepath.Join("..", "..", "migrations"))
	return &Storage{db: db, defaultPageSize: 10}, teardown
}

func envTest(t *testing.T) {
	config := viper.NewWithOptions(
		viper.EnvKeyReplacer(
			strings.NewReplacer(".", "_"),
		),
	)
	config.SetConfigFile("../../env/config")
	config.SetConfigType("ini")
	config.AutomaticEnv()
	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("error loading configuration: %v", err)
	}
	dConfig, err := postgresutil.NewDBFromConfig(config)
	if err != nil {
		t.Fatalf("error configure db, please set your config first")
	}
	// DATABASE_CONNECTION="user={user} host={host} port={port} dbname={dbname} password={password} sslmode=disable"
	t.Setenv(dbConnEnv, fmt.Sprintf("user=%s host=%s port=%s dbname=%s password=%s sslmode=%s",
		dConfig.User,
		dConfig.Host,
		dConfig.Port,
		dConfig.DBName,
		dConfig.Password,
		dConfig.SSLMode,
	))
}

func addQueryString(query, clause string) string {
	if strings.Contains(query, "WHERE") {
		return query + " AND " + clause
	}
	return query + " WHERE " + clause
}

func (s *Storage) SetupTx(ctx context.Context) (*Tx, func() error, error) {
	tx, err := s.BeginTx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error starting transaction: %w", err)
	}

	rollbackFn := func() error {
		err := s.Rollback(ctx)
		if err != nil {
			return fmt.Errorf("error rolling back transaction: %w", err)
		}
		return nil
	}

	return tx, rollbackFn, nil
}
