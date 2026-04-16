package cache

import (
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	DB *redis.Client
}

func New(db *redis.Client) *Cache {
	return &Cache{DB: db}
}
