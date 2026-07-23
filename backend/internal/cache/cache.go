package cache

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(addr, password string, db int) *Client {
	if addr == "" {
		return &Client{rdb: nil}
	}
	return &Client{rdb: redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})}
}

func TenantKey(tenantID uuid.UUID, parts ...string) string {
	segments := append([]string{"tenant", tenantID.String()}, parts...)
	return strings.Join(segments, ":")
}

func GlobalKey(parts ...string) string {
	return strings.Join(parts, ":")
}

func (c *Client) Get(ctx context.Context, key string, dest any) (bool, error) {
	if c.rdb == nil {
		return false, nil
	}
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		log.Printf("cache: GET %s failed: %v", key, err)
		return false, nil
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		log.Printf("cache: unmarshaling %s failed: %v", key, err)
		return false, nil
	}
	return true, nil
}

func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	if c.rdb == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache: marshaling %s failed: %v", key, err)
		return
	}
	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("cache: SET %s failed: %v", key, err)
	}
}

func (c *Client) Delete(ctx context.Context, keys ...string) {
	if c.rdb == nil || len(keys) == 0 {
		return
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache: DEL %v failed: %v", keys, err)
	}
}
