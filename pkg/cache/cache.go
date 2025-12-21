package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// CacheKey generates a cache key from request parameters
func CacheKey(model string, prompt string, params map[string]interface{}) string {
	data := struct {
		Model  string                 `json:"model"`
		Prompt string                 `json:"prompt"`
		Params map[string]interface{} `json:"params"`
	}{
		Model:  model,
		Prompt: prompt,
		Params: params,
	}

	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// ResponseCache interface for LLM response caching
type ResponseCache interface {
	Get(ctx context.Context, key string) (*CachedResponse, error)
	Set(ctx context.Context, key string, response *CachedResponse, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// CachedResponse represents a cached LLM response
type CachedResponse struct {
	Content  string    `json:"content"`
	Model    string    `json:"model"`
	Tokens   int       `json:"tokens"`
	CachedAt time.Time `json:"cached_at"`
}
