package storage

import (
	"context"
	"ivanjabrony/cloud-test/internal/ratelimit"
	"sync"
)

// BucketStorage is a in-memory key value storage for buckets
type BucketStorage struct {
	buckets map[string]*ratelimit.TokenBucket
	mu      sync.RWMutex
}

func NewBucketStorage() *BucketStorage {
	return &BucketStorage{make(map[string]*ratelimit.TokenBucket), sync.RWMutex{}}
}

// Store saves a bucket with a key
func (bs *BucketStorage) Store(ctx context.Context, key string, bucket *ratelimit.TokenBucket) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.buckets[key] = bucket
}

// Load returns Bucket, true if bucket is exists on a key or nil, false if it is not
func (bs *BucketStorage) Load(ctx context.Context, key string) (bucket *ratelimit.TokenBucket, ok bool) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	value, ok := bs.buckets[key]
	return value, ok
}

// Stop calls Stop method of every bucket in a storage
func (bs *BucketStorage) Stop(ctx context.Context) {
	for _, b := range bs.buckets {
		b.Stop()
	}
}
