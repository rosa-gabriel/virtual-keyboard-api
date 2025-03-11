package postgresql

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func Connect() {
	config, err := pgxpool.ParseConfig("user=admin password=password host=localhost port=5432 dbname=virtualkeyboard")

	if err != nil {
		slog.Error("Invalid database URL", "eror", err)
        panic(err)
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		slog.Error("Failed to connect to database pool", "eror", err)
        panic(err)
	}

	slog.Info("Connected to database")
}

func GetConection() (*pgxpool.Conn, error) {
	if pool == nil {
        return nil, errors.New("Not connected to database")
	}

	conn, err := pool.Acquire(context.Background())

	if err != nil {
		slog.Error("Failed to aquire database connection", "eror", err)
        return nil, err
	}

	return conn, nil
}

func Close() {
	if pool != nil {
		pool.Close()
	}

	pool = nil
}
