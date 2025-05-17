package ratelimit

import (
	"context"
	"errors"
)

// BucketStorage is an interface for storing buckets
//
// This implementation is based on map with mutex
type BucketStorage interface {
	Store(ctx context.Context, key string, bucket *TokenBucket)
	Load(ctx context.Context, key string) (bucket *TokenBucket, ok bool)
}

// RateLimiter is a main structure that rate-limits requests based on result of Allow() method
type RateLimiter struct {
	bucketStorage BucketStorage
	defaultCap    int     //Default capacity for a new client
	defaultRps    float64 //Default rps for a new client
}

func NewRateLimiter(bucketStorage BucketStorage, defaultCap int, defaultRps float64) (*RateLimiter, error) {
	if bucketStorage == nil {
		return nil, errors.New("nil values in ratelimiter constructor")
	}

	return &RateLimiter{bucketStorage, defaultCap, defaultRps}, nil
}

// addBucket adds new bucket to the storage and configures it
func (rl *RateLimiter) addBucket(ctx context.Context, ip string) *TokenBucket {
	bucket := NewTokenBucket(rl.defaultCap, rl.defaultRps)
	rl.bucketStorage.Store(ctx, ip, bucket)

	return bucket
}

// Allow is a method that chooses if request is allowed based on client storage state
//
//	If there are not enought tokens, TooManyRequests response will be sended
func (rl *RateLimiter) Allow(ctx context.Context, ip string) bool {
	bucket, ok := rl.bucketStorage.Load(ctx, ip)
	if !ok {
		bucket = rl.addBucket(ctx, ip)
	}

	return bucket.Allow()
}

// IsExists checks if there is a bucket for a client
func (rl *RateLimiter) IsExists(ctx context.Context, ip string) bool {
	_, ok := rl.bucketStorage.Load(ctx, ip)
	return ok
}
