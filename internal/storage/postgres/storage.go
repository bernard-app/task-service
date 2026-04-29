package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
