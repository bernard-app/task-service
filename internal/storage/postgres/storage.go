package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbEngine interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Storage struct {
	DB *pgxpool.Pool
}

func NewStorage(ctx context.Context, storagePath string) (*Storage, error) {
	const op = "storage.NewStorage"

	poolConfig, err := pgxpool.ParseConfig(storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) Close() error {
	s.DB.Close()
	
	return nil
}

func (s *Storage) getEngine(ctx context.Context) dbEngine {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}

	return s.DB
}