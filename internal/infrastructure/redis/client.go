package redis

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type DistributedLocker interface {
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
}

type Client struct {
	rdb        *redis.Client
	isFallback bool
	localMu    sync.Mutex
	localLocks map[string]time.Time
}

func NewClient(redisURL string) *Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("[Redis] Failed to parse Redis URL (%s): %v. Using in-memory fallback lock.", redisURL, err)
		return &Client{
			isFallback: true,
			localLocks: make(map[string]time.Time),
		}
	}

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[Redis] Cannot connect to Redis at %s (%v). Using in-memory fallback lock.", redisURL, err)
		return &Client{
			isFallback: true,
			localLocks: make(map[string]time.Time),
		}
	}

	log.Printf("[Redis] Connected to Redis at %s successfully.", opts.Addr)
	return &Client{
		rdb:        rdb,
		isFallback: false,
		localLocks: make(map[string]time.Time),
	}
}

// AcquireLock attempts to obtain a distributed lock (Redlock style)
func (c *Client) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if c.isFallback {
		c.localMu.Lock()
		defer c.localMu.Unlock()

		expiry, exists := c.localLocks[key]
		if exists && time.Now().Before(expiry) {
			return false, nil
		}
		c.localLocks[key] = time.Now().Add(ttl)
		return true, nil
	}

	return c.rdb.SetNX(ctx, "lock:"+key, "locked", ttl).Result()
}

// ReleaseLock releases the acquired lock
func (c *Client) ReleaseLock(ctx context.Context, key string) error {
	if c.isFallback {
		c.localMu.Lock()
		defer c.localMu.Unlock()
		delete(c.localLocks, key)
		return nil
	}

	return c.rdb.Del(ctx, "lock:"+key).Err()
}
