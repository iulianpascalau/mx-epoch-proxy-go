package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const missingKeyError = "redis: nil"

type redisWrapper struct {
	redisInstance *redis.Client
}

// NewRedisWrapper creates a new instance of type redis wrapper
func NewRedisWrapper(url string, password string) *redisWrapper {
	redisInstance := redis.NewClient(
		&redis.Options{
			Addr:     url,
			Password: password,
			DB:       0, // use default DB
		},
	)

	return &redisWrapper{
		redisInstance: redisInstance,
	}
}

// SetWithoutExpiry will set the provided key & value in storage without an expiry date
func (wrapper *redisWrapper) SetWithoutExpiry(ctx context.Context, key string, value string) error {
	return wrapper.redisInstance.Set(ctx, key, value, redis.KeepTTL).Err()
}

// Set will set the provided key & value in storage with the provided expiry date
func (wrapper *redisWrapper) Set(ctx context.Context, key string, value string, expiry time.Duration) error {
	return wrapper.redisInstance.Set(ctx, key, value, expiry).Err()
}

// Get retrieves the data from the storage
func (wrapper *redisWrapper) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := wrapper.redisInstance.Get(ctx, key).Result()
	if err == nil {
		return value, true, nil
	}

	if err.Error() == missingKeyError {
		return "", false, nil
	}

	return "", false, err
}

// Increment will increment the value for the provided key
func (wrapper *redisWrapper) Increment(ctx context.Context, key string) error {
	return wrapper.redisInstance.Incr(ctx, key).Err()
}

// Delete will delete the data from the provided key
func (wrapper *redisWrapper) Delete(ctx context.Context, key string) error {
	return wrapper.redisInstance.Del(ctx, key).Err()
}

// GetAllKeys will try to get all keys
func (wrapper *redisWrapper) GetAllKeys(ctx context.Context) ([]string, error) {
	result := make([]string, 0)

	iter := wrapper.redisInstance.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		result = append(result, iter.Val())
	}

	err := iter.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Close closes the current instance
func (wrapper *redisWrapper) Close() error {
	return wrapper.redisInstance.Close()
}

// IsInterfaceNil returns true if the value under the interface is nil
func (wrapper *redisWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
