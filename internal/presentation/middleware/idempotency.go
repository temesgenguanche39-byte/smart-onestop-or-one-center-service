package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const HeaderIdempotencyKey = "X-Idempotency-Key"

type IdempotencyStore struct {
	mu   sync.Mutex
	keys map[string]time.Time
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{
		keys: make(map[string]time.Time),
	}
}

// Idempotency deduplicates webhooks, intake posts, and state transitions
func Idempotency(store *IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(HeaderIdempotencyKey)
		if key == "" {
			c.Next()
			return
		}

		store.mu.Lock()
		expiry, exists := store.keys[key]
		if exists && time.Now().Before(expiry) {
			store.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error":           "Duplicate request detected for idempotency key",
				"idempotency_key": key,
			})
			return
		}

		// Keep key for 10 minutes
		store.keys[key] = time.Now().Add(10 * time.Minute)
		store.mu.Unlock()

		c.Next()
	}
}
