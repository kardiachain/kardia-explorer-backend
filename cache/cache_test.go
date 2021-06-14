// Package cache
package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var testRedisURL = "54.179.162.247:6379"

func SetupTestCache() (Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: testRedisURL,
		DB:   0,
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	lgr := logger.With(zap.String("cache", "redis"))
	client := &Redis{
		client: redisClient,
		logger: lgr,
	}
	return client, nil
}
