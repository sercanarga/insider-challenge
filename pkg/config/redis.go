package config

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// MessageCache represents the cached message data stored in redis
type MessageCache struct {
	SentAt    int64  `json:"sent_at"`
	MessageID string `json:"message_id"`
}

// InitRedis initializes the redis client with configuration from env variables
func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connection test
	_, err := RedisClient.Ping(ctx).Result()
	return err
}

// CacheMessageID caches a messageId with it sending time
func CacheMessageID(ctx context.Context, messageID string, webhookMessageID string) error {
	cache := MessageCache{
		SentAt:    time.Now().Unix(),
		MessageID: webhookMessageID,
	}

	jsonData, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return RedisClient.Set(ctx, "message:"+messageID, jsonData, 24*time.Hour).Err()
}

// GetMessageCache retrieves the sending time and messageId for message
func GetMessageCache(ctx context.Context, messageID string) (*MessageCache, error) {
	data, err := RedisClient.Get(ctx, "message:"+messageID).Bytes()
	if err != nil {
		return nil, err
	}

	var cache MessageCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}
