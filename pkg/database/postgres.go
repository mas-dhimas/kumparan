package database

import (
	"context"
	"database/sql"
	"fmt"
	"kumparan-test/config"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

// NewPostgresDB initializes and returns a new PostgreSQL database connection pool.
func NewPostgresDB(sdc *config.SourceDataConfig) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := sdc.PostgresDSN()

	pgxCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database dsn config: %w", err)
	}
	connStr := stdlib.RegisterConnConfig(pgxCfg)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(sdc.PostgresDBMaxConns)
	db.SetMaxIdleConns(sdc.PostgresDBMinConns)
	db.SetConnMaxLifetime(time.Duration(sdc.PostgresDBMaxConnLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(sdc.PostgresDBMaxConnIdleTime) * time.Second)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Infof("Successfully connected to PostgreSQL database (via database/sql)")
	return db, nil
}
