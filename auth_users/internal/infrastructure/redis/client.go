package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	Rdb *redis.Client
}

func New(addr string, password string, db int) *Client {
	return &Client{
		Rdb: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Rdb.Ping(ctx).Err()
}
