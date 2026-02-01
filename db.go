package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(30)

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	slog.Info("connected to database via database/sql")
	return db, nil
}
