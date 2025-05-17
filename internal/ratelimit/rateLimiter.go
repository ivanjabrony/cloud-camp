package ratelimit

import (
	"context"
	"errors"
)

type BucketStorage interface {
	Store(ctx context.Context, key string, bucket *TokenBucket)
	Load(ctx context.Context, key string) (bucket *TokenBucket, ok bool)
}

type RateLimiter struct {
	bucketStorage BucketStorage
	defaultCap    int
	defaultRps    float64
}

func NewRateLimiter(bucketStorage BucketStorage, defaultCap int, defaultRps float64) (*RateLimiter, error) {
	if bucketStorage == nil {
		return nil, errors.New("nil values in ratelimiter constructor")
	}

	return &RateLimiter{bucketStorage, defaultCap, defaultRps}, nil
}

func (rl *RateLimiter) addBucket(ctx context.Context, ip string) *TokenBucket {
	bucket := NewTokenBucket(rl.defaultCap, rl.defaultRps)
	rl.bucketStorage.Store(ctx, ip, bucket)

	return bucket
}

func (rl *RateLimiter) Allow(ctx context.Context, ip string) bool {
	bucket, ok := rl.bucketStorage.Load(ctx, ip)
	if !ok {
		bucket = rl.addBucket(ctx, ip)
	}

	return bucket.Allow()
}

func (rl *RateLimiter) IsExists(ctx context.Context, ip string) bool {
	_, ok := rl.bucketStorage.Load(ctx, ip)
	return ok
}
