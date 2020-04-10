package server

import (
	"redis-proxy/pkg/cache"
	"redis-proxy/pkg/config"

	"github.com/go-redis/redis"
)

// App a generic app struct
type App struct {
	config *config.Config
	client *redis.Client
	Cache  *cache.Cache
}

// newRedis creates a new Redis client
func newRedis(config config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: config.Address,
		DB:   config.Database,
	})
	_, err := client.Ping().Result()
	return client, err
}

// NewApp creates a new App
func NewApp(config *config.Config) (*App, error) {
	client, err := newRedis(config.RedisConfig)
	if err != nil {
		return nil, err
	}

	return &App{
		config: config,
		client: client,
		Cache:  cache.NewCache(config.CacheCapacity, config.CacheExpiry),
	}, nil
}
